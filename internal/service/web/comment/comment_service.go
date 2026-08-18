package comment

// 前台评论模块 —— 服务层实现（业务规则都在这）
//
// 一句话概括本模块的职责：把 Controller 传来的"用户想干什么"，
// 转成一串"对仓储的调用 + 业务校验"，并保证正确性。
//
// 值得留意的业务设计（也是面试/评审常问的点）：
//  1. 两级评论模型：所有回复统一落成二级，reply_to_id 永远指向根（一级）评论。
//  2. 评论列表 = "热评置顶 + 时间倒序分页"：前2条按热度 heat 排，其余按 create_time 倒序。
//     热评不占分页计数（total=一级评论真实总数，≤2 条时不排除热评），前端首屏就能看到优质评论。
//  3. 点赞用"行锁 + 先删后插"：锁住评论行（SELECT FOR UPDATE）把并发点赞串行化，
//     避免"取消再点赞"撞唯一索引 idx_comment_member 报 500。
//  4. 防 N+1：列表里每个评论都要查用户、点赞态、回复数……全部批量一次查回 map。
//  5. 写操作几乎都包在事务里：写评论时评论表、目标 comment_count、热度三处必须同时成功。
//  6. 删除是软删（藏起来）而非硬删（抹掉），与后台的硬删语义不同——见 task/comment_heat.go。
//
// 本项目事务的用法：s.db.Transaction(func(tx) error { ... }) 里所有带 tx 的仓储方法
// 都在同一个数据库事务里，任一步出错整体回滚。
import (
	"context"
	"math"
	"strconv"
	"strings"
	"time"

	requestComment "tuyou/internal/model/dto/request/web/comment"
	responseComment "tuyou/internal/model/dto/response/web/comment"
	"tuyou/internal/model/entity"
	"tuyou/internal/repository"
	repoComment "tuyou/internal/repository/web/comment"
	"tuyou/pkg/errors"
	"tuyou/pkg/response"
	"tuyou/pkg/sensitive"

	"gorm.io/gorm"
)

// CommentNotifier 评论消息通知接口：由消息模块的 Notifier 实现（Go 结构类型，无需显式声明实现）。
// 接口定义在评论包而不是直接依赖 message 包，避免 comment↔message 循环依赖。
type CommentNotifier interface {
	NotifyComment(ownerID, commenterID uint, title string, commentID uint)             // 一级评论：通知内容作者（模板3）
	NotifyCommentReply(commentOwnerID, replyerID uint, content string, commentID uint) // 二级评论：通知被回复的评论作者（模板10）
	NotifyCommentLike(commentOwnerID, likerID uint)                                    // 点赞评论：通知评论作者（模板13）
}

type commentService struct {
	db        *gorm.DB
	repo      repoComment.CommentRepository
	sensitive *sensitive.SensitiveMatcher
	redis     repository.RedisRepository
	notifier  CommentNotifier
}

func NewCommentService(db *gorm.DB, repo repoComment.CommentRepository, sensitive *sensitive.SensitiveMatcher, redis repository.RedisRepository, notifier CommentNotifier) CommentService {
	return &commentService{db: db, repo: repo, sensitive: sensitive, redis: redis, notifier: notifier}
}

// timeFormat 时间格式化样式（与项目其他模块一致）
const timeFormat = "2006-01-02 15:04:05"

const (
	// heatQuotaDaily 每用户每天可贡献的热度分上限（全局配额，设计 §2.2）
	heatQuotaDaily int64 = 50
	// heatReplyPerComment 单评论每用户最多计入热度的回复数（设计 §2.2）
	heatReplyPerComment int64 = 3
	// heatLikeScore 点赞贡献的热度分
	heatLikeScore int64 = 1
	// heatReplyScore 回复贡献的热度分
	heatReplyScore int64 = 2
)

// idStr uint 转字符串 ID
func idStr(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

// optionalIDStr 可空 ID 转字符串（nil→""）
func optionalIDStr(id *uint) string {
	if id == nil {
		return ""
	}
	return idStr(*id)
}

// heatQuotaKey 每日热度配额 key（自然日维度，次日换 key 天然重置）
func heatQuotaKey(memberID uint) string {
	return "member:heat:quota:" + time.Now().Format("20060102") + ":" + idStr(memberID)
}

// consumeHeatQuota 消费热度配额（Redis INCRBY 原子扣减）。返回 true=配额内（可计分），false=已超限（不计分）。
//   - Redis 不可用：fail-open 返回 true（不限制，与登录白名单降级一致）；
//   - 超限：回滚本次扣减（用掉的不回涨，超出的部分不占配额）；
//   - 首次创建 key：顺带设置"当天结束"过期，避免 key 无限堆积。
func (s *commentService) consumeHeatQuota(memberID uint, score int64) bool {
	if s.redis == nil {
		return true
	}
	ctx := context.Background()
	key := heatQuotaKey(memberID)
	newVal, err := s.redis.IncrBy(ctx, key, score)
	if err != nil {
		return true // Redis 不可用 → fail-open
	}
	if newVal == score {
		now := time.Now()
		endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
		_ = s.redis.Expire(ctx, key, endOfDay.Sub(now)+time.Second)
	}
	if newVal > heatQuotaDaily {
		_, _ = s.redis.IncrBy(ctx, key, -score)
		return false
	}
	return true
}

// shouldScoreReply 判断某条回复能否为父级贡献热度：① 未达单评论上限；② 配额有余额。
// 用掉不回涨：删除已计分回复不恢复名额、不扣父级热度（设计 §2.2）。
func shouldScoreReply(scoredReplies int64, quotaOK bool) bool {
	return scoredReplies < heatReplyPerComment && quotaOK
}

// sensitiveSnapshot 发布时敏感词快照：命中→(is_sensitive=1, 命中词分号连接)；未命中→(0, "")。
// 内容照常存原文（设计 §8.2），快照供后台筛选。
func sensitiveSnapshot(m *sensitive.SensitiveMatcher, content string) (int, string) {
	if m == nil {
		return 0, ""
	}
	words := m.FindWords(content)
	if len(words) == 0 {
		return 0, ""
	}
	return 1, strings.Join(words, ";")
}

// maskContent 前台出口统一打星（nil 安全：匹配器未注入时原样返回）
func (s *commentService) maskContent(content string) string {
	if s.sensitive == nil {
		return content
	}
	return s.sensitive.Mask(content)
}

// commentUser 组装评论用户信息（昵称打星）
func (s *commentService) commentUser(m *entity.Member) *responseComment.CommentUser {
	if m == nil {
		return &responseComment.CommentUser{}
	}
	return &responseComment.CommentUser{
		ID:       idStr(m.ID),
		Nickname: s.maskContent(m.Nickname),
		Avatar:   m.Avatar,
	}
}

// replyToNickname 取被回复用户的昵称（nil→""，昵称打星）
func (s *commentService) replyToNickname(memberID *uint, members map[uint]*entity.Member) string {
	if memberID == nil {
		return ""
	}
	if m := members[*memberID]; m != nil {
		return s.maskContent(m.Nickname)
	}
	return ""
}

// dedupe 去重并保持顺序
func dedupe(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// hotExcludeIDs 决定分页查询是否排除热评：一级评论总数 >2 时排除前2条热评（沿用原逻辑）；
// ≤2 时热评即全部，返回 nil 不排除——否则 total 被排空成 0（历史 bug：note60 仅 2 条一级评论时 total=0）。
func hotExcludeIDs(top []entity.Comment, totalTop int64) []uint {
	if totalTop <= 2 {
		return nil
	}
	ids := make([]uint, 0, len(top))
	for _, c := range top {
		ids = append(ids, c.ID)
	}
	return ids
}

// dedupeComments 按 ID 去重并保持顺序（总数≤2 时不排除热评，rest 与 top 会重复，合并后需去重）
func dedupeComments(comments []entity.Comment) []entity.Comment {
	seen := make(map[uint]struct{}, len(comments))
	result := make([]entity.Comment, 0, len(comments))
	for _, c := range comments {
		if _, ok := seen[c.ID]; ok {
			continue
		}
		seen[c.ID] = struct{}{}
		result = append(result, c)
	}
	return result
}

// buildReplyItems 组装二级评论列表项
func (s *commentService) buildReplyItems(replies []entity.Comment, members map[uint]*entity.Member, likes map[uint]bool, memberID, targetAuthorID uint) []*responseComment.CommentReplyItem {
	items := make([]*responseComment.CommentReplyItem, 0, len(replies))
	for _, r := range replies {
		items = append(items, &responseComment.CommentReplyItem{
			ID:              idStr(r.ID),
			User:            s.commentUser(members[r.MemberID]),
			Content:         s.maskContent(r.Content),
			LikeCount:       r.LikeCount,
			IsLiked:         likes[r.ID],
			ReplyToID:       optionalIDStr(r.ReplyToID),
			ReplyToNickname: s.replyToNickname(r.ReplyToMemberID, members),
			CreatedAt:       r.CreatedAt.Format(timeFormat),
			CanDelete:       r.MemberID == memberID || targetAuthorID == memberID,
		})
	}
	return items
}

// ToggleLike 点赞/取消点赞（翻转）
func (s *commentService) ToggleLike(commentID uint, memberID uint) (*responseComment.LikeToggleResponse, error) {
	var isLiked bool
	var likeCount int64
	var commentOwnerID uint
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 先锁评论行：挡掉 commentID=0 的非法请求，并串行化并发点赞——
		// 不锁行时"删点赞记录 + 插点赞记录"在并发下会撞唯一索引 idx_comment_member 导致 500。
		comment, err := s.repo.LockCommentByID(tx, commentID)
		if err != nil {
			return errors.NewWithErr(errors.CodeNotFound, "评论不存在", err)
		}
		commentOwnerID = comment.MemberID

		// 先试取消：删到了 → 之前已点赞，本次取消，like_count -1
		deleted, err := s.repo.DeleteCommentLike(tx, commentID, memberID)
		if err != nil {
			return err
		}
		if deleted {
			if err := s.repo.DecrLikeCount(tx, commentID); err != nil {
				return err
			}
			isLiked = false
		} else {
			// 没删到 → 之前未点赞，本次点赞，like_count +1
			if err := s.repo.CreateCommentLike(tx, commentID, memberID); err != nil {
				return err
			}
			if err := s.repo.IncrLikeCount(tx, commentID); err != nil {
				return err
			}
			isLiked = true
		}
		// 点赞贡献热度（配额内才计分）；取消点赞不回扣（用掉不回涨，设计 §2.2）
		if isLiked && s.consumeHeatQuota(memberID, heatLikeScore) {
			if err := s.repo.IncrHeatScore(tx, commentID, heatLikeScore); err != nil {
				return err
			}
		}

		// 最新点赞数 = 锁内读到的值 ± 1（持有行锁，无并发扰动，值准确）；-1 用 max 防负
		likeCount = comment.LikeCount
		if isLiked {
			likeCount++
		} else {
			likeCount = max(comment.LikeCount-1, 0)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// 点赞成功才通知评论作者（取消点赞不通知；点赞自己的评论由 notifier 内部跳过）
	if isLiked {
		s.notifier.NotifyCommentLike(commentOwnerID, memberID)
	}
	return &responseComment.LikeToggleResponse{IsLiked: isLiked, LikeCount: likeCount}, nil
}

// CreateComment 发表评论
func (s *commentService) CreateComment(memberID uint, req *requestComment.CreateCommentRequest) (*responseComment.CommentItem, error) {
	// 目标类型校验
	if req.TargetType != 1 && req.TargetType != 2 {
		return nil, errors.New(errors.CodeInvalidParam, "目标类型错误")
	}
	// 内容去首尾空格后不能为空（binding 的 required 挡不住纯空格），存库前统一 trim
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		return nil, errors.New(errors.CodeInvalidParam, "评论内容不能为空")
	}
	// 敏感词快照：命中不拦截，is_sensitive/hit_words 落库，内容存原文（设计 §8.4）
	isSensitive, hitWords := sensitiveSnapshot(s.sensitive, req.Content)
	// 校验目标存在（笔记/攻略），并取作品作者 memberID 与标题（标题用于一级评论通知正文）
	targetAuthorID, title, err := s.repo.FindTargetAuthor(req.TargetType, req.TargetID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "目标不存在", err)
	}

	// 二级评论：对任何评论的回复都统一存二级，reply_to_id 指向归属的一级评论
	var replyToID, replyToMemberID, parentID *uint
	if req.ReplyToID != nil {
		replied, err := s.repo.FindByID(*req.ReplyToID)
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeNotFound, "被回复的评论不存在", err)
		}
		if replied.TargetType != req.TargetType || replied.TargetID != req.TargetID {
			return nil, errors.New(errors.CodeInvalidParam, "被回复的评论不属于当前目标")
		}
		// 被回复的若是一级评论，它即归属；若是二级评论，取其归属的一级评论ID
		rootID := *req.ReplyToID
		if replied.ReplyToID != nil {
			rootID = *replied.ReplyToID
		}
		replyToID = &rootID
		// 直接父 = 真正回复的那条评论（消息中心回复列表按此绑定；reply_to_id 仍指向根保线程结构）
		parentID = &replied.ID
		// C5：忽略客户端传的被回复用户，一律取被回复评论的作者（防伪造被回复人）
		mid := replied.MemberID
		replyToMemberID = &mid
	}

	comment := &entity.Comment{
		TargetType:      req.TargetType,
		TargetID:        req.TargetID,
		MemberID:        memberID,
		Content:         req.Content,
		ReplyToID:       replyToID,
		ReplyToMemberID: replyToMemberID,
		ParentID:        parentID,
		LikeCount:       0,
		Heat:            0,
		Status:          1,
		IsSensitive:     isSensitive,
		HitWords:        hitWords,
	}
	// 事务：写评论 + 目标 comment_count +1 + 重算新评论热度；
	// 若是回复，父一级评论的 reply_count 变了，同事务重算父评论热度，避免热评榜/回复预览滞后到小时任务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// C3：锁序统一"先锁根一级评论行"，与 DeleteComment 一致，消除死锁。
		// 回复场景锁父一级行，也让"新回复"与"删父"串行化（C1 计数准确）。
		if replyToID != nil {
			if _, err := s.repo.LockCommentByID(tx, *replyToID); err != nil {
				return errors.NewWithErr(errors.CodeNotFound, "被回复的评论不存在", err)
			}
			// 回复计分（设计 §2.3）：单评论上限（该用户对根评论已计分回复 <3）+ 全局配额。
			// 用掉不回涨：删回复不扣父级热度、不恢复配额（heat_scored 标记保持 1）。
			scoredReplies, err := s.repo.CountScoredReplies(tx, *replyToID, memberID)
			if err != nil {
				return err
			}
			// 短路求值：上限未过才消费配额（避免上限已满还白白扣配额）
			quotaOK := scoredReplies < heatReplyPerComment && s.consumeHeatQuota(memberID, heatReplyScore)
			if shouldScoreReply(scoredReplies, quotaOK) {
				if err := s.repo.IncrHeatScore(tx, *replyToID, heatReplyScore); err != nil {
					return err
				}
				comment.HeatScored = 1
			}
		}
		if err := s.repo.CreateComment(tx, comment); err != nil {
			return err
		}
		if err := s.repo.IncrTargetCommentCount(tx, req.TargetType, req.TargetID); err != nil {
			return err
		}
		if err := s.repo.RecomputeHeat(tx, comment.ID); err != nil {
			return err
		}
		if replyToID != nil {
			if err := s.repo.RecomputeHeat(tx, *replyToID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "发表评论失败", err)
	}

	// 生成评论通知消息（best-effort，失败不影响主流程；评论自己内容/回复自己不通知）。
	// 一级评论 → 通知内容作者（模板3）；二级评论 → 通知被回复的评论作者（模板10）。
	// commentID 写进消息，接收者可从消息中心继续回复这条评论。
	if replyToID == nil {
		s.notifier.NotifyComment(targetAuthorID, memberID, title, comment.ID)
	} else if replyToMemberID != nil {
		s.notifier.NotifyCommentReply(*replyToMemberID, memberID, s.maskContent(req.Content), comment.ID)
	}

	// 组装响应（作者=自己，canDelete=true）。
	// 评论者和被回复者一起去重查询，否则被回复者的昵称永远取不到（恒为空）。
	queryIDs := []uint{memberID}
	if replyToMemberID != nil {
		queryIDs = append(queryIDs, *replyToMemberID)
	}
	members, err := s.repo.FindMembersByIDs(dedupe(queryIDs))
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取用户信息失败", err)
	}
	var replyToNicknameStr string
	if replyToMemberID != nil {
		if m := members[*replyToMemberID]; m != nil {
			replyToNicknameStr = s.maskContent(m.Nickname)
		}
	}
	return &responseComment.CommentItem{
		ID:              idStr(comment.ID),
		User:            s.commentUser(members[memberID]),
		Content:         s.maskContent(comment.Content),
		LikeCount:       0,
		IsLiked:         false,
		ReplyToID:       optionalIDStr(replyToID),
		ReplyToNickname: replyToNicknameStr,
		CreatedAt:       comment.CreatedAt.Format(timeFormat),
		Replies:         []*responseComment.CommentReplyItem{},
		ReplyCount:      0,
		CanDelete:       true,
	}, nil
}

// ListComments 评论列表：前2条热评仅第1页展示（分页排除热评）+ 其余按发布时间倒序分页，total=一级评论真实总数（≤2 条时不排除热评）
func (s *commentService) ListComments(memberID uint, req *requestComment.ListCommentRequest) (*responseComment.CommentListResponse, error) {
	if req.TargetType != 1 && req.TargetType != 2 {
		return nil, errors.New(errors.CodeInvalidParam, "目标类型错误")
	}
	// 校验目标存在，拿作品作者ID（用于 canDelete）
	targetAuthorID, _, err := s.repo.FindTargetAuthor(req.TargetType, req.TargetID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "目标不存在", err)
	}

	// 前2条热评
	top, err := s.repo.FindTopComments(req.TargetType, req.TargetID, 2)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询评论失败", err)
	}
	// 一级评论真实总数（不排除任何评论）——total 必须等于它。
	// 历史 bug：总数≤2 时热评即全部，旧代码把热评从 total 排除 → total=0（如 note60）。
	total, err := s.repo.CountCommentPage(req.TargetType, req.TargetID, nil)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询评论失败", err)
	}
	// 分页排除热评；总数≤2 时不排除（热评即全部，excludeIDs=nil）
	excludeIDs := hotExcludeIDs(top, total)

	// 其余按发布时间倒序分页
	offset := (req.Page - 1) * req.Size
	rest, err := s.repo.FindCommentPage(req.TargetType, req.TargetID, excludeIDs, offset, req.Size)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询评论失败", err)
	}

	// 热评仅第1页展示，第2页起不再重复带出；分页"其余"每页照常排除热评，total/TotalPage 保持一致
	// Page<=1 而非 ==1：绑定层已保证 min=1，但程序化调用 Page=0 时不应丢热评
	// 总数≤2 时 excludeIDs 为空，rest 会重复带出 top → 合并后按 ID 去重
	comments := make([]entity.Comment, 0, len(top)+len(rest))
	if req.Page <= 1 {
		comments = append(comments, top...)
	}
	comments = append(comments, rest...)
	comments = dedupeComments(comments)

	items, err := s.buildCommentItems(comments, memberID, targetAuthorID)
	if err != nil {
		return nil, err
	}

	// 全量评论数（一级+二级）：实时统计、天然自愈，不依赖 ±1 维护的计数，
	// 前端"共N条评论"直接用它，避免显示成"当前加载的一级条数"
	totalComments, err := s.repo.CountAllComments(req.TargetType, req.TargetID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询评论总数失败", err)
	}

	return &responseComment.CommentListResponse{
		List:          items,
		Total:         total,
		TotalComments: totalComments,
		Page:          req.Page,
		Size:          req.Size,
		TotalPage:     int(math.Ceil(float64(total) / float64(req.Size))),
	}, nil
}

// buildCommentItems 组装一级评论列表项（用户信息、点赞态、二级热评摘要、回复数、删除权限）
func (s *commentService) buildCommentItems(comments []entity.Comment, memberID, targetAuthorID uint) ([]*responseComment.CommentItem, error) {
	items := make([]*responseComment.CommentItem, 0, len(comments))
	if len(comments) == 0 {
		return items, nil
	}

	// 收集一级评论作者 + 被回复用户
	memberIDs := make([]uint, 0, len(comments))
	commentIDs := make([]uint, 0, len(comments))
	for _, c := range comments {
		commentIDs = append(commentIDs, c.ID)
		memberIDs = append(memberIDs, c.MemberID)
		if c.ReplyToMemberID != nil {
			memberIDs = append(memberIDs, *c.ReplyToMemberID)
		}
	}

	// 每个一级评论下的二级评论总数
	replyCounts, err := s.repo.CountRepliesByParentIDs(commentIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询回复数失败", err)
	}

	// C4：热回复批量查询——一次取回本页全部一级评论下的二级，按父分组取热度前2，取代逐条 FindTopReplies 的 N+1
	repliesByParent, err := s.repo.FindRepliesByParentIDs(commentIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询热评失败", err)
	}
	tmp := make(map[uint][]entity.Comment, len(commentIDs))
	for _, r := range repliesByParent {
		if r.ReplyToID == nil {
			continue
		}
		pid := *r.ReplyToID
		grouped := tmp[pid]
		if len(grouped) < 2 {
			tmp[pid] = append(grouped, r) // 整体已按 heat DESC 排，每父前2即热度最高前2
		}
	}
	// 展示门限：仅当该父热度最高的二级评论点赞数 >10 才展示摘要
	replyMap := make(map[uint][]entity.Comment, len(commentIDs))
	allReplies := make([]entity.Comment, 0, len(commentIDs)*2)
	for parentID, topReplies := range tmp {
		if len(topReplies) > 0 && topReplies[0].LikeCount > 10 {
			replyMap[parentID] = topReplies
			allReplies = append(allReplies, topReplies...)
		}
	}

	// 收集二级评论作者 + 评论ID（用于批量查用户、点赞态）
	for _, r := range allReplies {
		commentIDs = append(commentIDs, r.ID)
		memberIDs = append(memberIDs, r.MemberID)
		if r.ReplyToMemberID != nil {
			memberIDs = append(memberIDs, *r.ReplyToMemberID)
		}
	}

	members, err := s.repo.FindMembersByIDs(dedupe(memberIDs))
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询用户信息失败", err)
	}
	likes, err := s.repo.FindCommentLikes(commentIDs, memberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询点赞状态失败", err)
	}

	// 组装
	for _, c := range comments {
		items = append(items, &responseComment.CommentItem{
			ID:              idStr(c.ID),
			User:            s.commentUser(members[c.MemberID]),
			Content:         s.maskContent(c.Content),
			LikeCount:       c.LikeCount,
			IsLiked:         likes[c.ID],
			ReplyToID:       optionalIDStr(c.ReplyToID),
			ReplyToNickname: s.replyToNickname(c.ReplyToMemberID, members),
			CreatedAt:       c.CreatedAt.Format(timeFormat),
			Replies:         s.buildReplyItems(replyMap[c.ID], members, likes, memberID, targetAuthorID),
			ReplyCount:      replyCounts[c.ID],
			CanDelete:       c.MemberID == memberID || targetAuthorID == memberID,
		})
	}
	return items, nil
}

// DeleteComment 删除评论（评论者本人或作品作者可删；一级评论连带软删除其下所有二级评论）
func (s *commentService) DeleteComment(commentID uint, memberID uint) error {
	comment, err := s.repo.FindByID(commentID)
	if err != nil {
		return errors.NewWithErr(errors.CodeNotFound, "评论不存在", err)
	}
	// 作品作者（校验目标仍存在）
	targetAuthorID, _, err := s.repo.FindTargetAuthor(comment.TargetType, comment.TargetID)
	if err != nil {
		return errors.NewWithErr(errors.CodeNotFound, "目标不存在", err)
	}
	// 权限：评论者本人 或 作品作者
	if comment.MemberID != memberID && targetAuthorID != memberID {
		return errors.New(errors.CodeForbidden, "无权删除该评论")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// C1+C3：事务第一句锁"根一级评论行"（一级=自身，二级=其 ReplyToID）。
		// 串行化并发插入的二级评论：先提交的回复 COUNT 可见；后到的回复等锁，
		// 待删除提交后 LockCommentByID（默认作用域）查不到已软删行 → 404。
		rootID := commentID
		if comment.ReplyToID != nil {
			rootID = *comment.ReplyToID
		}
		if _, err := s.repo.LockCommentByID(tx, rootID); err != nil {
			return errors.NewWithErr(errors.CodeNotFound, "评论不存在", err)
		}
		if comment.ReplyToID == nil {
			// 删除一级评论：先数其下未删二级数（用于 comment_count 级联扣减），再连带软删
			replyCount, err := s.repo.CountRepliesByParentID(tx, commentID)
			if err != nil {
				return err
			}
			if err := s.repo.DeleteCommentByID(tx, commentID); err != nil {
				return err
			}
			if err := s.repo.DeleteRepliesByParentID(tx, commentID); err != nil {
				return err
			}
			// comment_count - (1 + 级联二级数)
			return s.repo.DecrTargetCommentCount(tx, comment.TargetType, comment.TargetID, 1+replyCount)
		}
		// 删除二级评论：仅删自身，comment_count -1，并重算其归属一级评论的热度
		if err := s.repo.DeleteCommentByID(tx, commentID); err != nil {
			return err
		}
		if err := s.repo.DecrTargetCommentCount(tx, comment.TargetType, comment.TargetID, 1); err != nil {
			return err
		}
		// 删除回复不重算父级热度：热度只由 heat_score（删除不回扣）与时间驱动（设计 §2.2）
		return nil
	})
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "删除评论失败", err)
	}
	return nil
}

// ListReplies 二级评论列表：前2条热回复（不占分页计数）+ 其余按发布时间正序分页，total=其余回复数
func (s *commentService) ListReplies(commentID uint, memberID uint, page, size int) (*response.PageResponse, error) {
	// 一级评论必须存在
	parent, err := s.repo.FindByID(commentID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "评论不存在", err)
	}
	if parent.ReplyToID != nil {
		return nil, errors.New(errors.CodeInvalidParam, "只能查询一级评论的回复")
	}
	// 作品作者（用于 canDelete）
	targetAuthorID, _, err := s.repo.FindTargetAuthor(parent.TargetType, parent.TargetID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "目标不存在", err)
	}

	// 前2条热回复
	top, err := s.repo.FindTopReplies(commentID, 2)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询回复失败", err)
	}
	excludeIDs := make([]uint, 0, len(top))
	for _, c := range top {
		excludeIDs = append(excludeIDs, c.ID)
	}

	// 其余按发布时间正序分页
	offset := (page - 1) * size
	rest, err := s.repo.FindReplyPage(commentID, excludeIDs, offset, size)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询回复失败", err)
	}
	total, err := s.repo.CountReplyPage(commentID, excludeIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询回复失败", err)
	}

	replies := make([]entity.Comment, 0, len(top)+len(rest))
	replies = append(replies, top...)
	replies = append(replies, rest...)

	// 批量查用户、点赞态
	memberIDs := make([]uint, 0, len(replies))
	commentIDs := make([]uint, 0, len(replies))
	for _, r := range replies {
		commentIDs = append(commentIDs, r.ID)
		memberIDs = append(memberIDs, r.MemberID)
		if r.ReplyToMemberID != nil {
			memberIDs = append(memberIDs, *r.ReplyToMemberID)
		}
	}
	members, err := s.repo.FindMembersByIDs(dedupe(memberIDs))
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询用户信息失败", err)
	}
	likes, err := s.repo.FindCommentLikes(commentIDs, memberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询点赞状态失败", err)
	}

	items := s.buildReplyItems(replies, members, likes, memberID, targetAuthorID)
	return response.NewPageResponse(items, total, page, size), nil
}
