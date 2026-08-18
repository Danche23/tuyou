package message

import (
	stdErrors "errors"
	"math"
	"strconv"
	"strings"
	"tuyou/internal/stream"

	"gorm.io/gorm"
	requestComment "tuyou/internal/model/dto/request/web/comment"
	requestMsg "tuyou/internal/model/dto/request/web/message"
	responseMsg "tuyou/internal/model/dto/response/web/message"
	repoMsg "tuyou/internal/repository/web/message"
	serviceComment "tuyou/internal/service/web/comment"
	"tuyou/pkg/errors"
	"tuyou/pkg/sensitive"
)

// 消息类型（与 msg_info.msg_type 约定一致）
const (
	msgTypeInteraction = 1 // 互动消息
	msgTypeSystem      = 2 // 系统通知
)

type msgService struct {
	msgRepo    repoMsg.MessageRepository
	db         *gorm.DB
	hub        *stream.SSEHub
	commentSvc serviceComment.CommentService // 回复评论时复用评论模块的发表评论逻辑
	sensitive  *sensitive.SensitiveMatcher
}

// NewMessageService 创建消息服务实例
func NewMessageService(msgRepo repoMsg.MessageRepository, db *gorm.DB, hub *stream.SSEHub, commentSvc serviceComment.CommentService, matcher *sensitive.SensitiveMatcher) MessageService {
	return &msgService{msgRepo: msgRepo, db: db, hub: hub, commentSvc: commentSvc, sensitive: matcher}
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
		List:      s.buildItems(rows, memberID),
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

// Delete 删除单条消息（软删除：is_user_del=1，列表/未读数自动排除）
func (s *msgService) Delete(memberID, msgID uint) error {
	err := s.msgRepo.DeleteUserMessage(memberID, msgID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.CodeNotFound, "消息不存在")
		}
		return errors.NewWithErr(errors.CodeInternalError, "删除消息失败", err)
	}

	// 删除可能影响未读数（删了未读消息），推送最新未读数让角标实时更新
	if uc, err := s.UnreadCount(memberID); err == nil {
		s.hub.Publish(memberID, "unread", uc)
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

// Reply 回复互动消息关联的评论：校验消息归属 → 取评论 → 调评论模块 CreateComment 以回复身份插一条二级评论。
// 评论创建、comment_count/热度维护、给被回复人生成消息（模板10）与推送未读数，全部由评论模块统一处理，这里不再重复造轮子。
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

	// 以回复身份走评论模块发表二级评论：被回复人 = 原评论作者。
	resp, err := s.commentSvc.CreateComment(memberID, &requestComment.CreateCommentRequest{
		TargetType:      c.TargetType,
		TargetID:        c.TargetID,
		Content:         content,
		ReplyToID:       &commentID,
		ReplyToMemberID: &c.MemberID,
	})
	if err != nil {
		return err
	}

	// SSE 推送给原评论作者：带上新回复ID，前端可即时刷新对应评论的回复列表。
	// 非阻塞：对方离线/缓冲满直接丢弃，下次上线靠 /stream 连接时的未读数兜底。
	replyID, _ := strconv.ParseUint(resp.ID, 10, 64)
	s.hub.Publish(c.MemberID, "comment-reply", map[string]any{
		"replyId":   replyID,
		"replyToId": commentID,
		"content":   content,
		"senderId":  memberID,
	})
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

// buildItems 联查结果 → 响应项（系统消息 sender 为 null）。
// 评论类消息（comment_id>0）附带 comment：本条消息（"you"）+ 只含"我"发的回复（"me"）。
func (s *msgService) buildItems(rows []repoMsg.MessageRow, memberID uint) []responseMsg.MessageItem {
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
				Nickname: s.mask(r.Nickname),
				Avatar:   r.Avatar,
			}
		}
		if r.CommentID > 0 {
			item.Comment = s.buildComment(r, memberID)
		}
		items = append(items, item)
	}
	return items
}

// buildComment 组装评论类消息的 comment：
// 本条消息（sender="you"）+ 回复列表只含"我"发的回复（sender="me"，时间正序）。
// 回复列表按 parent_id 绑定：直接回复本条评论的、当前用户自己的回复。
// best-effort：任何查询失败返回 nil（该消息退化为不带 comment）。
func (s *msgService) buildComment(r repoMsg.MessageRow, memberID uint) *responseMsg.CommentMessage {
	c, err := s.msgRepo.FindCommentByID(r.CommentID)
	if err != nil {
		return nil
	}
	myReplies, err := s.msgRepo.FindRepliesByParent(c.ID, memberID)
	if err != nil {
		return nil
	}

	likeIDs := []uint{c.ID}
	for _, tc := range myReplies {
		likeIDs = append(likeIDs, tc.ID)
	}
	likes, err := s.msgRepo.FindCommentLikes(likeIDs, memberID)
	if err != nil {
		likes = map[uint]bool{}
	}

	msg := &responseMsg.CommentMessage{
		CommentID: strconv.FormatUint(uint64(c.ID), 10),
		Sender:    "you",
		Content:   c.Content,
		SendTime:  c.CreatedAt.Format("2006-01-02 15:04:05"),
		LikeCount: c.LikeCount,
		IsLiked:   likes[c.ID],
	}
	for _, tc := range myReplies {
		msg.Replies = append(msg.Replies, &responseMsg.CommentMessage{
			CommentID: strconv.FormatUint(uint64(tc.ID), 10),
			Sender:    "me",
			Content:   tc.Content,
			SendTime:  tc.CreatedAt.Format("2006-01-02 15:04:05"),
			LikeCount: tc.LikeCount,
			IsLiked:   likes[tc.ID],
		})
	}
	return msg
}

// mask 前台文本出口统一打星（nil 安全）
func (s *msgService) mask(content string) string {
	if s.sensitive == nil {
		return content
	}
	return s.sensitive.Mask(content)
}
