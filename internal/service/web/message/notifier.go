package message

import (
	"strings"
	"time"

	"gorm.io/gorm"

	responseMsg "tuyou/internal/model/dto/response/web/message"
	"tuyou/internal/model/entity"
	repoMsg "tuyou/internal/repository/web/message"
	"tuyou/internal/stream"
)

// Notifier 消息通知器：给指定会员生成一条消息（查 msg_template 渲染正文）并推送实时通知。
// 供点赞/关注/评论/报名/后台删除等服务调用；best-effort，失败不影响主流程。
type Notifier struct {
	msgRepo repoMsg.MessageRepository
	db      *gorm.DB
	hub     *stream.SSEHub
}

// NewNotifier 创建通知器
func NewNotifier(msgRepo repoMsg.MessageRepository, db *gorm.DB, hub *stream.SSEHub) *Notifier {
	return &Notifier{msgRepo: msgRepo, db: db, hub: hub}
}

// NotifyLike 点赞通知（模板1）
func (n *Notifier) NotifyLike(contentOwnerID, senderID uint, title string) {
	n.send(contentOwnerID, senderID, 1, "like", map[string]string{
		"nickname": n.nickname(senderID),
		"title":    title,
	})
}

// NotifyFollow 关注通知（模板2）
func (n *Notifier) NotifyFollow(followeeID, followerID uint) {
	n.send(followeeID, followerID, 2, "follow", map[string]string{
		"nickname": n.nickname(followerID),
	})
}

// NotifyCollect 收藏通知（模板12）
func (n *Notifier) NotifyCollect(ownerID, collectorID uint, title string) {
	n.send(ownerID, collectorID, 12, "collect", map[string]string{
		"nickname": n.nickname(collectorID),
		"title":    title,
	})
}

// NotifySignup 报名通知（给活动领队，模板6）
func (n *Notifier) NotifySignup(leaderID, memberID uint, title string) {
	n.send(leaderID, memberID, 6, "signup", map[string]string{
		"nickname": n.nickname(memberID),
		"title":    title,
	})
}

// NotifySignupSuccess 报名成功通知（给报名者本人，模板8）
func (n *Notifier) NotifySignupSuccess(memberID uint, title string) {
	n.send(memberID, 0, 8, "signup-success", map[string]string{
		"title": title,
	})
}

// NotifyActivityDeleted 活动删除通知（给所有已报名成员，模板11）
func (n *Notifier) NotifyActivityDeleted(memberIDs []uint, leaderID uint, title string) {
	nickname := n.nickname(leaderID)
	for _, memberID := range memberIDs {
		n.send(memberID, leaderID, 11, "activity-deleted", map[string]string{
			"nickname": nickname,
			"title":    title,
		})
	}
}

// NotifyComment 评论通知（给内容作者，模板3）。
// commentID 关联新评论，接收者可从消息中心回复该评论。
func (n *Notifier) NotifyComment(ownerID, commenterID uint, title string, commentID uint) {
	n.sendTo(ownerID, commenterID, 3, "comment", commentID, map[string]string{
		"nickname": n.nickname(commenterID),
		"title":    title,
	})
}

// NotifyCommentReply 回复评论通知（给被回复的评论作者，模板10）。
// commentID 关联新回复评论，接收者可从消息中心继续回复。
func (n *Notifier) NotifyCommentReply(commentOwnerID, replyerID uint, content string, commentID uint) {
	n.sendTo(commentOwnerID, replyerID, 10, "comment-reply", commentID, map[string]string{
		"nickname": n.nickname(replyerID),
		"comment":  content,
	})
}

// NotifyCommentLike 点赞评论通知（给评论作者，模板13）
func (n *Notifier) NotifyCommentLike(commentOwnerID, likerID uint) {
	n.send(commentOwnerID, likerID, 13, "comment-like", map[string]string{
		"nickname": n.nickname(likerID),
	})
}

// NotifyActivityStart 活动快开始提醒（给已报名者本人，模板9）
func (n *Notifier) NotifyActivityStart(memberID uint, title string) {
	n.send(memberID, 0, 9, "activity-start", map[string]string{
		"title": title,
	})
}

// NotifyNoteDeleted 笔记删除通知（给作者，模板14）
func (n *Notifier) NotifyNoteDeleted(ownerID uint, title string) {
	n.send(ownerID, 0, 14, "content-deleted", map[string]string{
		"title": title,
	})
}

// NotifyGuideDeleted 攻略删除通知（给作者，模板15）
func (n *Notifier) NotifyGuideDeleted(ownerID uint, title string) {
	n.send(ownerID, 0, 15, "content-deleted", map[string]string{
		"title": title,
	})
}

// NotifyNoteOffline 笔记下架通知（给作者，模板16）
func (n *Notifier) NotifyNoteOffline(ownerID uint, title string) {
	n.send(ownerID, 0, 16, "content-offline", map[string]string{
		"title": title,
	})
}

// NotifyGuideOffline 攻略下架通知（给作者，模板17）
func (n *Notifier) NotifyGuideOffline(ownerID uint, title string) {
	n.send(ownerID, 0, 17, "content-offline", map[string]string{
		"title": title,
	})
}

// send 通用：查模板渲染正文 → 写 msg_info + msg_user_record（同一事务）→ 推送事件 + 最新未读数。
// 不带关联评论，CommentID 固定为 0；评论类消息请用 sendTo。
func (n *Notifier) send(targetMemberID, senderID uint, templateID uint, event string, vars map[string]string) {
	n.sendTo(targetMemberID, senderID, templateID, event, 0, vars)
}

// sendTo 通用：同 send，额外带 commentID（评论消息关联新评论，供消息中心回复评论）。
// msg_type 从模板表读取；目标为 0 或触达自己时直接跳过。
func (n *Notifier) sendTo(targetMemberID, senderID uint, templateID uint, event string, commentID uint, vars map[string]string) {
	if targetMemberID == 0 || targetMemberID == senderID {
		return // 无人可通知 / 自己触达自己不通知
	}

	tpl, err := n.findTemplate(templateID)
	if err != nil {
		return // 模板不存在，best-effort 跳过
	}
	content := renderContent(tpl.ContentTemplate, vars)

	info := &entity.MsgInfo{
		MsgType:        tpl.MsgType,
		ReceiveObjType: 1,
		SenderID:       senderID,
		TemplateID:     templateID,
		MsgContent:     content,
		SendTime:       time.Now(),
		CommentID:      commentID,
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

// findTemplate 查消息模板
func (n *Notifier) findTemplate(templateID uint) (*entity.MsgTemplate, error) {
	var tpl entity.MsgTemplate
	if err := n.db.Table("msg_template").Where("id = ?", templateID).First(&tpl).Error; err != nil {
		return nil, err
	}
	return &tpl, nil
}

// Render 查模板并渲染出最终正文（供定时任务幂等查重等需提前拿到正文的场景使用）
func (n *Notifier) Render(templateID uint, vars map[string]string) string {
	tpl, err := n.findTemplate(templateID)
	if err != nil {
		return ""
	}
	return renderContent(tpl.ContentTemplate, vars)
}

// renderContent 用变量替换模板占位符 {var}
func renderContent(tpl string, vars map[string]string) string {
	for k, v := range vars {
		tpl = strings.ReplaceAll(tpl, "{"+k+"}", v)
	}
	return tpl
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
