package message

import (
	"context"
	stdErrors "errors"
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	request "tuyou/internal/model/dto/request/admin/message"
	response "tuyou/internal/model/dto/response/admin/message"
	responseMsg "tuyou/internal/model/dto/response/web/message"
	"tuyou/internal/model/entity"
	repo "tuyou/internal/repository/admin/message"
	"tuyou/internal/stream"
	"tuyou/pkg/errors"
)

// 消息类型（与 msg_info.msg_type 约定一致）
const (
	msgTypeInteraction = 1 // 互动消息
	msgTypeSystem      = 2 // 系统消息
)

// 互动消息：后台类型 → template_id 集合（筛选翻译用）
var interactionTypeMap = map[string][]uint{
	"点赞": {1, 13}, // 1=赞笔记/攻略，13=赞评论
	"关注": {2},
	"评论": {3, 10}, // 3=评论，10=回复评论
	"收藏": {12},
}

// 系统消息：后台类型 → template_id 集合（筛选翻译用）
// 前端系统消息筛选项值为「系统」，与展示名「系统通知」对齐两种都收。
var systemTypeMap = map[string][]uint{
	"活动":   {6, 8, 9, 11}, // 报名通知/报名成功/活动开始/活动删除
	"系统":   {5},           // 手动发布
	"系统通知": {5},           // 手动发布（别名）
}

type msgService struct {
	msgRepo repo.MessageRepository
	hub     *stream.SSEHub // 会员端 SSE，发布后推前台未读数
}

// NewMessageService 创建后台消息管理服务实例
func NewMessageService(msgRepo repo.MessageRepository, hub *stream.SSEHub) MessageService {
	return &msgService{msgRepo: msgRepo, hub: hub}
}

// ListInteractions 分页查询互动消息
func (s *msgService) ListInteractions(ctx context.Context, req *request.ListMessageReq) (*response.ListMessageResp, error) {
	return s.list(ctx, msgTypeInteraction, req)
}

// ListSystem 分页查询系统消息
func (s *msgService) ListSystem(ctx context.Context, req *request.ListMessageReq) (*response.ListMessageResp, error) {
	return s.list(ctx, msgTypeSystem, req)
}

// list 通用列表查询：拼装筛选条件 → 查询 → 映射为中文类型
func (s *msgService) list(ctx context.Context, msgType int, req *request.ListMessageReq) (*response.ListMessageResp, error) {
	filters := make(map[string]interface{})
	if req.Type != "" {
		var m map[string][]uint
		if msgType == msgTypeInteraction {
			m = interactionTypeMap
		} else {
			m = systemTypeMap
		}
		templateIDs, ok := m[req.Type]
		if !ok {
			return nil, errors.New(errors.CodeInvalidParam, "消息类型参数不合法")
		}
		filters["template_ids"] = templateIDs
	}
	if req.BeginTime != "" {
		filters["begin_time"] = req.BeginTime
	}
	if req.FinishTime != "" {
		filters["finish_time"] = req.FinishTime + " 23:59:59"
	}

	rows, total, err := s.msgRepo.List(ctx, msgType, filters, req.Page, req.Size)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询消息列表失败", err)
	}
	return &response.ListMessageResp{
		List:      buildItems(rows, msgType),
		Total:     total,
		Page:      req.Page,
		Size:      req.Size,
		TotalPage: int(math.Ceil(float64(total) / float64(req.Size))),
	}, nil
}

// PublishSystemMsg 发布系统消息：校验 → 确定接收者 → 事务落库 → 推前台未读数
func (s *msgService) PublishSystemMsg(ctx context.Context, req *request.PublishSystemMsgReq) error {
	// 1. 校验
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return errors.New(errors.CodeBadRequest, "消息内容不能为空")
	}
	if req.ReceiveType != entity.ReceiveObjTypeAll && req.ReceiveType != entity.ReceiveObjTypeMember {
		return errors.New(errors.CodeInvalidParam, "接收类型不合法")
	}
	if req.ReceiveType == entity.ReceiveObjTypeMember && len(req.MemberIDs) == 0 {
		return errors.New(errors.CodeInvalidParam, "指定接收者不能为空")
	}

	// 2. 确定接收者：全体=所有会员（含领队）；指定=前端 id 列表
	var memberIDs []uint
	if req.ReceiveType == entity.ReceiveObjTypeAll {
		ids, err := s.msgRepo.AllMemberIDs(ctx)
		if err != nil {
			return errors.NewWithErr(errors.CodeInternalError, "查询接收者失败", err)
		}
		memberIDs = ids
	} else {
		memberIDs = req.MemberIDs
	}

	// 3. 组装消息并事务落库（template_id=5 系统通知模板，sender=0 系统）
	info := &entity.MsgInfo{
		MsgType:        msgTypeSystem, // 2 系统消息
		ReceiveObjType: req.ReceiveType,
		SenderID:       0, // 0=系统
		TemplateID:     5, // 系统通知模板
		MsgContent:     content,
		SendTime:       time.Now(),
	}
	if err := s.msgRepo.Create(ctx, info, memberIDs); err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "发布消息失败", err)
	}

	// 4. 逐个接收者推最新未读数到前台（best-effort，失败不影响发布结果）
	for _, mid := range memberIDs {
		s.pushUnread(ctx, mid)
	}
	return nil
}

// pushUnread 推送该会员最新未读数（前台角标实时更新），best-effort
func (s *msgService) pushUnread(ctx context.Context, memberID uint) {
	interactions, system, err := s.msgRepo.UnreadCount(ctx, memberID)
	if err != nil {
		return
	}
	s.hub.Publish(memberID, "unread", &responseMsg.UnreadCount{
		Total:        interactions + system,
		Interactions: interactions,
		System:       system,
	})
}

// Delete 硬删除一条消息（连带其全部接收记录）
func (s *msgService) Delete(ctx context.Context, id uint) error {
	if _, err := s.msgRepo.FindByID(ctx, id); err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.CodeNotFound, "消息不存在")
		}
		return errors.NewWithErr(errors.CodeInternalError, "查询消息失败", err)
	}
	if err := s.msgRepo.Delete(ctx, id); err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "删除消息失败", err)
	}
	return nil
}

// buildItems 联查结果 → 响应项（类型/接收者/发送者映射为中文）
func buildItems(rows []repo.MessageRow, msgType int) []response.MessageItem {
	items := make([]response.MessageItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, response.MessageItem{
			ID:          strconv.FormatUint(uint64(r.ID), 10),
			MsgType:     msgTypeName(msgType, r.TemplateID),
			Content:     r.MsgContent,
			ReceiveType: receiveTypeName(r.ReceiveObjType),
			Sender:      senderName(r.SenderID, r.SenderNickname, r.SenderAdminName),
			SendTime:    r.SendTime.Format("2006-01-02 15:04:05"),
		})
	}
	return items
}

// msgTypeName template_id → 后台展示类型
func msgTypeName(msgType int, templateID uint) string {
	if msgType == msgTypeInteraction {
		switch templateID {
		case 1, 13:
			return "点赞"
		case 2:
			return "关注"
		case 3, 10:
			return "评论"
		case 12:
			return "收藏"
		}
		return "其他"
	}
	switch templateID {
	case 6, 8, 9, 11:
		return "活动"
	default:
		return "系统通知"
	}
}

// receiveTypeName receive_obj_type → 接收类型
func receiveTypeName(t int) string {
	if t == entity.ReceiveObjTypeAll {
		return "全体"
	}
	return "指定会员"
}

// senderName 发送者中文：0=系统；优先会员昵称，否则管理员姓名
func senderName(id uint, nickname, adminName string) string {
	if id == 0 {
		return "系统"
	}
	if nickname != "" {
		return nickname
	}
	if adminName != "" {
		return adminName
	}
	return "系统"
}
