package message

import (
	stdErrors "errors"
	"math"
	"strconv"
	"strings"

	"gorm.io/gorm"
	requestMsg "tuyou/internal/model/dto/request/web/message"
	responseMsg "tuyou/internal/model/dto/response/web/message"
	"tuyou/internal/model/entity"
	repoMsg "tuyou/internal/repository/web/message"
	"tuyou/pkg/errors"
)

// 消息类型（与 msg_info.msg_type 约定一致）
const (
	msgTypeInteraction = 1 // 互动消息
	msgTypeSystem      = 2 // 系统通知
)

type msgService struct {
	msgRepo repoMsg.MessageRepository
	db      *gorm.DB
}

// NewMessageService 创建消息服务实例
func NewMessageService(msgRepo repoMsg.MessageRepository, db *gorm.DB) MessageService {
	return &msgService{msgRepo: msgRepo, db: db}
}

// GetInteractions 获取互动消息列表
func (s *msgService) GetInteractions(memberID uint, req *requestMsg.MessageListRequest) (*responseMsg.MessagePage, error) {
	return s.getList(memberID, msgTypeInteraction, req)
}

// GetSystem 获取系统通知列表
func (s *msgService) GetSystem(memberID uint, req *requestMsg.MessageListRequest) (*responseMsg.MessagePage, error) {
	return s.getList(memberID, msgTypeSystem, req)
}

// getList 通用列表查询
func (s *msgService) getList(memberID uint, msgType int, req *requestMsg.MessageListRequest) (*responseMsg.MessagePage, error) {
	page, size := normalizePage(req.Page, req.Size)
	rows, total, err := s.msgRepo.FindMessages(memberID, msgType, page, size)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询消息列表失败", err)
	}
	return &responseMsg.MessagePage{
		List:      buildItems(rows),
		Total:     total,
		Page:      page,
		Size:      size,
		TotalPage: int(math.Ceil(float64(total) / float64(size))),
	}, nil
}

// MarkRead 标记单条消息已读
func (s *msgService) MarkRead(memberID, msgID uint) error {
	err := s.msgRepo.MarkRead(memberID, msgID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.CodeNotFound, "消息不存在")
		}
		return errors.NewWithErr(errors.CodeInternalError, "标记已读失败", err)
	}
	return nil
}

// MarkAllRead 标记全部消息已读
func (s *msgService) MarkAllRead(memberID uint) error {
	if err := s.msgRepo.MarkAllRead(memberID); err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "标记全部已读失败", err)
	}
	return nil
}

// UnreadCount 未读消息数（互动+系统拆分）
func (s *msgService) UnreadCount(memberID uint) (*responseMsg.UnreadCount, error) {
	interactions, system, err := s.msgRepo.UnreadCount(memberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询未读数失败", err)
	}
	return &responseMsg.UnreadCount{
		Total:        interactions + system,
		Interactions: interactions,
		System:       system,
	}, nil
}

// Reply 回复互动消息关联的评论：校验消息归属 → 取评论 → 以回复身份在评论表插一条 reply
func (s *msgService) Reply(memberID, msgID uint, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New(errors.CodeBadRequest, "回复内容不能为空")
	}

	commentID, err := s.msgRepo.FindCommentIDByMsg(msgID, memberID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.CodeNotFound, "消息不存在")
		}
		return errors.NewWithErr(errors.CodeInternalError, "查询消息失败", err)
	}
	if commentID == 0 {
		return errors.New(errors.CodeBadRequest, "该消息不可回复")
	}

	c, err := s.msgRepo.FindCommentByID(commentID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.CodeNotFound, "评论不存在")
		}
		return errors.NewWithErr(errors.CodeInternalError, "查询评论失败", err)
	}
	if c.Status != 1 {
		return errors.New(errors.CodeBadRequest, "评论已删除，无法回复")
	}

	reply := &entity.Comment{
		TargetType: c.TargetType,
		TargetID:   c.TargetID,
		MemberID:   memberID,
		Content:    content,
		ReplyToID:  &commentID,
		Status:     1,
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		return s.msgRepo.CreateComment(tx, reply)
	}); err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "回复失败", err)
	}
	return nil
}

// normalizePage 分页参数归一化：page 默认 1，size 默认 10，上限 100
func normalizePage(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

// buildItems 联查结果 → 响应项（系统消息 sender 为 null）
func buildItems(rows []repoMsg.MessageRow) []responseMsg.MessageItem {
	items := make([]responseMsg.MessageItem, 0, len(rows))
	for _, r := range rows {
		item := responseMsg.MessageItem{
			ID:       strconv.FormatUint(uint64(r.ID), 10),
			MsgType:  r.MsgType,
			Content:  r.Content,
			SendTime: r.SendTime.Format("2006-01-02 15:04:05"),
			IsRead:   r.IsRead,
		}
		if r.SenderID > 0 {
			item.Sender = &responseMsg.Sender{
				ID:       strconv.FormatUint(uint64(r.SenderID), 10),
				Nickname: r.Nickname,
				Avatar:   r.Avatar,
			}
		}
		items = append(items, item)
	}
	return items
}
