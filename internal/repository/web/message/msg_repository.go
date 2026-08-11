package message

import (
	"gorm.io/gorm"
	"tuyou/internal/model/entity"
)

type msgRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &msgRepository{db: db}
}

// FindMessages 查询指定会员某类型消息列表（msg_user_record JOIN msg_info JOIN member）
// 列表仅含该会员的接收记录（receive_member_id 命中），发送者联查 member 表取昵称头像。
func (r *msgRepository) FindMessages(memberID uint, msgType, page, size int) ([]MessageRow, int64, error) {
	var total int64
	err := r.db.Table("msg_user_record").
		Joins("JOIN msg_info ON msg_info.id = msg_user_record.msg_id AND msg_info.is_del = 0").
		Where("msg_user_record.receive_member_id = ? AND msg_info.msg_type = ? AND msg_user_record.is_user_del = 0", memberID, msgType).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var rows []MessageRow
	err = r.db.Table("msg_user_record").
		Select("msg_info.id, msg_info.msg_type, msg_info.msg_content, msg_info.send_time, msg_info.sender_id, msg_user_record.is_read, mb.nickname, mb.avatar").
		Joins("JOIN msg_info ON msg_info.id = msg_user_record.msg_id AND msg_info.is_del = 0").
		Joins("LEFT JOIN member mb ON mb.id = msg_info.sender_id AND msg_info.sender_id > 0").
		Where("msg_user_record.receive_member_id = ? AND msg_info.msg_type = ? AND msg_user_record.is_user_del = 0", memberID, msgType).
		Order("msg_info.send_time DESC, msg_info.id DESC").
		Offset((page - 1) * size).Limit(size).
		Scan(&rows).Error
	return rows, total, err
}

// MarkRead 标记单条消息已读（该会员无此消息的接收记录时返回 gorm.ErrRecordNotFound）
// 先查存在性再更新：避免「已读状态下重复标记」RowsAffected=0 被误判为不存在。
func (r *msgRepository) MarkRead(memberID, msgID uint) error {
	var cnt int64
	if err := r.db.Model(&entity.MsgUserRecord{}).
		Where("msg_id = ? AND receive_member_id = ? AND is_user_del = 0", msgID, memberID).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return gorm.ErrRecordNotFound
	}
	return r.db.Model(&entity.MsgUserRecord{}).
		Where("msg_id = ? AND receive_member_id = ? AND is_user_del = 0", msgID, memberID).
		UpdateColumn("is_read", 1).Error
}

// MarkAllRead 标记全部消息已读
func (r *msgRepository) MarkAllRead(memberID uint) error {
	return r.db.Model(&entity.MsgUserRecord{}).
		Where("receive_member_id = ? AND is_read = 0 AND is_user_del = 0", memberID).
		UpdateColumn("is_read", 1).Error
}

// UnreadCount 未读消息数（按 msg_type 分组：1=互动，2=系统）
func (r *msgRepository) UnreadCount(memberID uint) (interactions, system int64, err error) {
	var rows []struct {
		MsgType int   `gorm:"column:msg_type"`
		Num     int64 `gorm:"column:num"`
	}
	err = r.db.Table("msg_user_record").
		Select("msg_info.msg_type AS msg_type, COUNT(*) AS num").
		Joins("JOIN msg_info ON msg_info.id = msg_user_record.msg_id AND msg_info.is_del = 0").
		Where("msg_user_record.receive_member_id = ? AND msg_user_record.is_read = 0 AND msg_user_record.is_user_del = 0", memberID).
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

// FindCommentIDByMsg 按消息+会员查询关联评论id（校验消息归属，无记录返回 gorm.ErrRecordNotFound）
func (r *msgRepository) FindCommentIDByMsg(msgID, memberID uint) (uint, error) {
	var row struct {
		CommentID uint `gorm:"column:comment_id"`
	}
	err := r.db.Table("msg_user_record").
		Select("msg_info.comment_id").
		Joins("JOIN msg_info ON msg_info.id = msg_user_record.msg_id AND msg_info.is_del = 0").
		Where("msg_user_record.msg_id = ? AND msg_user_record.receive_member_id = ? AND msg_user_record.is_user_del = 0", msgID, memberID).
		Scan(&row).Error
	if err != nil {
		return 0, err
	}
	return row.CommentID, nil
}

// FindCommentByID 查询评论
func (r *msgRepository) FindCommentByID(commentID uint) (*entity.Comment, error) {
	var c entity.Comment
	if err := r.db.First(&c, commentID).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// CreateComment 在事务中插入回复评论
func (r *msgRepository) CreateComment(tx *gorm.DB, c *entity.Comment) error {
	return tx.Create(c).Error
}
