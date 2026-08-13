package center

import (
	"context"
	"time"

	"gorm.io/gorm"
	"tuyou/internal/model/entity"
)

// adminMsgType 后台个人消息的消息类型（msg_info.msg_type）：2 系统通知
const adminMsgType = 2

type adminMsgRepository struct {
	db *gorm.DB
}

func NewAdminMsgRepository(db *gorm.DB) AdminMsgRepository {
	return &adminMsgRepository{db: db}
}

// Create 生成一条管理员系统通知：msg_info + msg_user_record 同一事务，best-effort 由调用方决定
func (r *adminMsgRepository) Create(ctx context.Context, adminID uint, content string) error {
	info := &entity.MsgInfo{
		MsgType:        adminMsgType,               // 2 系统通知
		ReceiveObjType: entity.ReceiveObjTypeAdmin, // 2 指定管理员
		SenderID:       0,                          // 0=系统
		TemplateID:     0,
		MsgContent:     content,
		SendTime:       time.Now(),
	}
	record := &entity.MsgUserRecord{
		ReceiveMemberID: adminID,
		IsRead:          0,
		IsUserDel:       0,
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(info).Error; err != nil {
			return err
		}
		record.MsgID = info.ID // Create 后 gorm 回填主键
		return tx.Create(record).Error
	})
}

// List 分页查询管理员个人消息（仅 receive_obj_type=2 的消息）
func (r *adminMsgRepository) List(ctx context.Context, adminID uint, page, size int) ([]AdminMsgRow, int64, error) {
	var total int64
	err := r.db.Table("msg_user_record").
		Joins("JOIN msg_info ON msg_info.id = msg_user_record.msg_id AND msg_info.is_del = 0").
		Where("msg_user_record.receive_member_id = ? AND msg_info.receive_obj_type = ? AND msg_user_record.is_user_del = 0", adminID, entity.ReceiveObjTypeAdmin).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var rows []AdminMsgRow
	err = r.db.Table("msg_user_record").
		Select("msg_info.id, msg_info.msg_type AS type, msg_info.msg_content AS content, msg_user_record.is_read, msg_info.send_time").
		Joins("JOIN msg_info ON msg_info.id = msg_user_record.msg_id AND msg_info.is_del = 0").
		Where("msg_user_record.receive_member_id = ? AND msg_info.receive_obj_type = ? AND msg_user_record.is_user_del = 0", adminID, entity.ReceiveObjTypeAdmin).
		Order("msg_info.send_time DESC, msg_info.id DESC").
		Offset((page - 1) * size).Limit(size).
		Scan(&rows).Error
	return rows, total, err
}

// UnreadCount 未读消息数（仅 receive_obj_type=2 的消息）
func (r *adminMsgRepository) UnreadCount(ctx context.Context, adminID uint) (int64, error) {
	var cnt int64
	err := r.db.Table("msg_user_record").
		Joins("JOIN msg_info ON msg_info.id = msg_user_record.msg_id AND msg_info.is_del = 0").
		Where("msg_user_record.receive_member_id = ? AND msg_info.receive_obj_type = ? AND msg_user_record.is_read = 0 AND msg_user_record.is_user_del = 0", adminID, entity.ReceiveObjTypeAdmin).
		Count(&cnt).Error
	return cnt, err
}

// MarkRead 标记单条消息已读（无该消息记录返回 gorm.ErrRecordNotFound）
// 子查询限定 msg_id 属于 receive_obj_type=2 的消息，防止与前台会员消息混淆
func (r *adminMsgRepository) MarkRead(ctx context.Context, adminID, msgID uint) error {
	where := "msg_id = ? AND receive_member_id = ? AND is_user_del = 0 AND msg_id IN (SELECT id FROM msg_info WHERE is_del = 0 AND receive_obj_type = ?)"
	var cnt int64
	if err := r.db.Model(&entity.MsgUserRecord{}).
		Where(where, msgID, adminID, entity.ReceiveObjTypeAdmin).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return gorm.ErrRecordNotFound
	}
	return r.db.Model(&entity.MsgUserRecord{}).
		Where(where, msgID, adminID, entity.ReceiveObjTypeAdmin).
		UpdateColumn("is_read", 1).Error
}

// MarkAllRead 标记全部消息已读
func (r *adminMsgRepository) MarkAllRead(ctx context.Context, adminID uint) error {
	return r.db.Model(&entity.MsgUserRecord{}).
		Where("receive_member_id = ? AND is_read = 0 AND is_user_del = 0 AND msg_id IN (SELECT id FROM msg_info WHERE is_del = 0 AND receive_obj_type = ?)", adminID, entity.ReceiveObjTypeAdmin).
		UpdateColumn("is_read", 1).Error
}

// Delete 软删除单条消息（is_user_del=1，无该消息记录返回 gorm.ErrRecordNotFound）
func (r *adminMsgRepository) Delete(ctx context.Context, adminID, msgID uint) error {
	where := "msg_id = ? AND receive_member_id = ? AND is_user_del = 0 AND msg_id IN (SELECT id FROM msg_info WHERE is_del = 0 AND receive_obj_type = ?)"
	var cnt int64
	if err := r.db.Model(&entity.MsgUserRecord{}).
		Where(where, msgID, adminID, entity.ReceiveObjTypeAdmin).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return gorm.ErrRecordNotFound
	}
	return r.db.Model(&entity.MsgUserRecord{}).
		Where(where, msgID, adminID, entity.ReceiveObjTypeAdmin).
		UpdateColumn("is_user_del", 1).Error
}

// ClearAll 清空该管理员全部消息（软删除）
func (r *adminMsgRepository) ClearAll(ctx context.Context, adminID uint) error {
	return r.db.Model(&entity.MsgUserRecord{}).
		Where("receive_member_id = ? AND is_user_del = 0 AND msg_id IN (SELECT id FROM msg_info WHERE is_del = 0 AND receive_obj_type = ?)", adminID, entity.ReceiveObjTypeAdmin).
		UpdateColumn("is_user_del", 1).Error
}
