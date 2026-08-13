package message

import (
	"time"

	"gorm.io/gorm"

	responseMsg "tuyou/internal/model/dto/response/web/message"
	"tuyou/internal/model/entity"
	repoMsg "tuyou/internal/repository/web/message"
	"tuyou/internal/stream"
)

// Notifier 互动消息通知器：给指定会员生成一条互动消息并推送实时通知。
// 供点赞/关注/评论/报名等服务调用；best-effort，失败不影响主流程。
type Notifier struct {
	msgRepo repoMsg.MessageRepository
	db      *gorm.DB
	hub     *stream.SSEHub
}

// NewNotifier 创建通知器
func NewNotifier(msgRepo repoMsg.MessageRepository, db *gorm.DB, hub *stream.SSEHub) *Notifier {
	return &Notifier{msgRepo: msgRepo, db: db, hub: hub}
}

// NotifyLike 点赞通知：「nickname」赞了你的内容《title》
func (n *Notifier) NotifyLike(contentOwnerID, senderID uint, title string) {
	n.send(contentOwnerID, senderID, msgTypeInteraction, 1, "like",
		"「"+n.nickname(senderID)+"」赞了你的内容《"+title+"》")
}

// NotifyFollow 关注通知：「nickname」关注了你
func (n *Notifier) NotifyFollow(followeeID, followerID uint) {
	n.send(followeeID, followerID, msgTypeInteraction, 2, "follow",
		"「"+n.nickname(followerID)+"」关注了你")
}

// NotifyCollect 收藏通知（给内容作者）：「nickname」收藏了你的内容《title》
// 模板 msg_type=1，互动消息
func (n *Notifier) NotifyCollect(ownerID, collectorID uint, title string) {
	n.send(ownerID, collectorID, msgTypeInteraction, 12, "collect",
		"「"+n.nickname(collectorID)+"」收藏了你的内容《"+title+"》")
}

// NotifySignup 报名通知（给活动领队）：「nickname」报名了你的活动《title》
// 模板 msg_type=2，系统通知
func (n *Notifier) NotifySignup(leaderID, memberID uint, title string) {
	n.send(leaderID, memberID, msgTypeSystem, 6, "signup",
		"「"+n.nickname(memberID)+"」报名了你的活动《"+title+"》")
}

// NotifySignupSuccess 报名成功通知（给报名者本人）：你已成功报名活动《title》
// 模板 msg_type=2，系统通知，发送者=系统(0)
func (n *Notifier) NotifySignupSuccess(memberID uint, title string) {
	n.send(memberID, 0, msgTypeSystem, 8, "signup-success",
		"你已成功报名活动《"+title+"》")
}

// NotifyActivityDeleted 活动删除通知（给所有已报名成员）：「nickname」删除了活动《title》
// 模板 msg_type=2，系统通知，发送者=活动领队。逐个成员发送，best-effort。
func (n *Notifier) NotifyActivityDeleted(memberIDs []uint, leaderID uint, title string) {
	nickname := n.nickname(leaderID)
	for _, memberID := range memberIDs {
		n.send(memberID, leaderID, msgTypeSystem, 11, "activity-deleted",
			"「"+nickname+"」删除了活动《"+title+"》")
	}
}

// NotifyComment 评论通知（一级评论，给内容作者）：「nickname」评论了你的内容《title》
// 模板 msg_type=1，互动消息，正文不含评论内容（模板3）。
// 注：暂未关联 comment_id；将来若要让作者能「回复」这条评论，调用处需让 send 带上新评论id。
func (n *Notifier) NotifyComment(ownerID, commenterID uint, title string) {
	n.send(ownerID, commenterID, msgTypeInteraction, 3, "comment",
		"「"+n.nickname(commenterID)+"」评论了你的内容《"+title+"》")
}

// NotifyCommentLike 点赞评论通知（给评论作者）：「nickname」赞了你的评论
// 模板 msg_type=1，互动消息。注：web 层评论点赞功能尚未实现，待接线。
func (n *Notifier) NotifyCommentLike(commentOwnerID, likerID uint) {
	n.send(commentOwnerID, likerID, msgTypeInteraction, 13, "comment-like",
		"「"+n.nickname(likerID)+"」赞了你的评论")
}

// NotifyActivityStart 活动快开始提醒（给已报名者本人，content 为最终正文）
// 模板 msg_type=2，系统通知，发送者=系统(0)。content 由定时任务生成并用于幂等查重。
func (n *Notifier) NotifyActivityStart(memberID uint, content string) {
	n.send(memberID, 0, msgTypeSystem, 9, "activity-start", content)
}

// send 通用：写 msg_info + msg_user_record（同一事务）→ 推送事件 + 最新未读数。
// msgType 按模板的 msg_type 传（1=互动，2=系统）；目标为 0 或触达自己时直接跳过。
func (n *Notifier) send(targetMemberID, senderID uint, msgType, templateID int, event, content string) {
	if targetMemberID == 0 || targetMemberID == senderID {
		return // 无人可通知 / 自己触达自己不通知
	}

	info := &entity.MsgInfo{
		MsgType:        msgType,
		ReceiveObjType: 1,
		SenderID:       senderID,
		TemplateID:     uint(templateID),
		MsgContent:     content,
		SendTime:       time.Now(),
		CommentID:      0,
	}
	record := &entity.MsgUserRecord{
		ReceiveMemberID: targetMemberID,
		IsRead:          0,
		IsUserDel:       0,
	}
	if err := n.db.Transaction(func(tx *gorm.DB) error {
		if err := n.msgRepo.CreateMessage(tx, info); err != nil {
			return err
		}
		record.MsgID = info.ID // Create 后 gorm 回填主键
		return n.msgRepo.CreateUserRecord(tx, record)
	}); err != nil {
		return // best-effort：通知失败不影响主流程
	}

	// 推送实时事件
	n.hub.Publish(targetMemberID, event, map[string]any{
		"content":  content,
		"senderId": senderID,
	})
	// 推送最新未读数，角标实时更新
	n.pushUnread(targetMemberID)
}

// nickname 查发送者昵称，查不到兜底"用户"
func (n *Notifier) nickname(memberID uint) string {
	nick, _ := n.msgRepo.FindMemberNickname(memberID)
	if nick == "" {
		return "用户"
	}
	return nick
}

// pushUnread 推送该会员最新未读数
func (n *Notifier) pushUnread(targetMemberID uint) {
	interactions, system, err := n.msgRepo.UnreadCount(targetMemberID)
	if err != nil {
		return
	}
	n.hub.Publish(targetMemberID, "unread", &responseMsg.UnreadCount{
		Total:        interactions + system,
		Interactions: interactions,
		System:       system,
	})
}
