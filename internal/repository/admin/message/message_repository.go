package message

import (
	"context"

	"gorm.io/gorm"
	"tuyou/internal/model/entity"
)

type msgRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建后台消息管理仓储实例
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &msgRepository{db: db}
}

// applyFilters 拼装动态查询条件（仅 msg_info 字段，COUNT 与列表共用，列名无歧义）
func applyFilters(query *gorm.DB, msgType int, filters map[string]interface{}) *gorm.DB {
	query = query.Where("msg_type = ? AND is_del = 0", msgType)
	if v, ok := filters["template_ids"]; ok {
		query = query.Where("template_id IN ?", v)
	}
	if v, ok := filters["begin_time"]; ok {
		query = query.Where("send_time >= ?", v)
	}
	if v, ok := filters["finish_time"]; ok {
		query = query.Where("send_time <= ?", v)
	}
	return query
}

// List 分页查询消息主表（一条消息一行，联查发送者中文，按 send_time 倒序）
func (r *msgRepository) List(ctx context.Context, msgType int, filters map[string]interface{}, page, size int) ([]MessageRow, int64, error) {
	var total int64
	err := applyFilters(r.db.WithContext(ctx).Table("msg_info"), msgType, filters).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var rows []MessageRow
	err = applyFilters(r.db.WithContext(ctx).Table("msg_info").
		Select("msg_info.id, msg_info.msg_type, msg_info.msg_content, msg_info.receive_obj_type, msg_info.sender_id, msg_info.template_id, msg_info.send_time, mb.nickname AS sender_nickname, adm.admin_name AS sender_admin_name").
		Joins("LEFT JOIN member mb ON mb.id = msg_info.sender_id AND msg_info.sender_id > 0").
		Joins("LEFT JOIN sys_admin adm ON adm.id = msg_info.sender_id AND msg_info.sender_id > 0"),
		msgType, filters).
		Order("msg_info.send_time DESC, msg_info.id DESC").
		Offset((page - 1) * size).Limit(size).
		Scan(&rows).Error
	return rows, total, err
}

// FindByID 按 id 查询消息主表，不存在返回 gorm.ErrRecordNotFound
func (r *msgRepository) FindByID(ctx context.Context, id uint) (*entity.MsgInfo, error) {
	var info entity.MsgInfo
	err := r.db.WithContext(ctx).Table("msg_info").
		Where("id = ? AND is_del = 0", id).
		First(&info).Error
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// Create 事务内创建消息：msg_info + 批量接收记录（任一步失败整体回滚）
func (r *msgRepository) Create(ctx context.Context, info *entity.MsgInfo, memberIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(info).Error; err != nil {
			return err
		}
		records := make([]entity.MsgUserRecord, 0, len(memberIDs))
		for _, mid := range memberIDs {
			records = append(records, entity.MsgUserRecord{
				MsgID:           info.ID, // Create 后 gorm 回填主键
				ReceiveMemberID: mid,
				IsRead:          0,
				IsUserDel:       0,
			})
		}
		if len(records) > 0 {
			return tx.Create(&records).Error
		}
		return nil
	})
}

// AllMemberIDs 查询所有会员 id（含领队 role=0 和普通会员 role=1）
func (r *msgRepository) AllMemberIDs(ctx context.Context) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&entity.Member{}).Pluck("id", &ids).Error
	return ids, err
}

// UnreadCount 会员未读消息数（互动/系统拆分），与前台查询口径一致（receive_obj_type IN 0,1）
func (r *msgRepository) UnreadCount(ctx context.Context, memberID uint) (interactions, system int64, err error) {
	var rows []struct {
		MsgType int   `gorm:"column:msg_type"`
		Num     int64 `gorm:"column:num"`
	}
	err = r.db.WithContext(ctx).Table("msg_user_record").
		Select("msg_info.msg_type AS msg_type, COUNT(*) AS num").
		Joins("JOIN msg_info ON msg_info.id = msg_user_record.msg_id AND msg_info.is_del = 0").
		Where("msg_user_record.receive_member_id = ? AND msg_info.receive_obj_type IN (?, ?) AND msg_user_record.is_read = 0 AND msg_user_record.is_user_del = 0",
			memberID, entity.ReceiveObjTypeAll, entity.ReceiveObjTypeMember).
		Group("msg_info.msg_type").
		Scan(&rows).Error
	if err != nil {
		return 0, 0, err
	}
	for _, row := range rows {
		switch row.MsgType {
		case 1: // 互动消息
			interactions = row.Num
		case 2: // 系统通知
			system = row.Num
		}
	}
	return interactions, system, nil
}

// Delete 硬删除消息主表及其全部接收记录（事务，Unscoped 绕过软删物理移除）
func (r *msgRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("msg_id = ?", id).Delete(&entity.MsgUserRecord{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&entity.MsgInfo{}, id).Error
	})
}
