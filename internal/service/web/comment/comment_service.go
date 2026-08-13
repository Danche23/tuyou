package comment

import (
	"strconv"
	"strings"

	requestComment "tuyou/internal/model/dto/request/web/comment"
	responseComment "tuyou/internal/model/dto/response/web/comment"
	"tuyou/internal/model/entity"
	repoComment "tuyou/internal/repository/web/comment"
	"tuyou/pkg/errors"
	"tuyou/pkg/response"

	"gorm.io/gorm"
)

type commentService struct {
	db   *gorm.DB
	repo repoComment.CommentRepository
}

func NewCommentService(db *gorm.DB, repo repoComment.CommentRepository) CommentService {
	return &commentService{db: db, repo: repo}
}

// timeFormat 时间格式化样式（与项目其他模块一致）
const timeFormat = "2006-01-02 15:04:05"

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

// commentUser 组装评论用户信息
func commentUser(m *entity.Member) *responseComment.CommentUser {
	if m == nil {
		return &responseComment.CommentUser{}
	}
	return &responseComment.CommentUser{
		ID:       idStr(m.ID),
		Nickname: m.Nickname,
		Avatar:   m.Avatar,
	}
}

// replyToNickname 取被回复用户的昵称（nil→""）
func replyToNickname(memberID *uint, members map[uint]*entity.Member) string {
	if memberID == nil {
		return ""
	}
	if m := members[*memberID]; m != nil {
		return m.Nickname
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

// buildReplyItems 组装二级评论列表项
func buildReplyItems(replies []entity.Comment, members map[uint]*entity.Member, likes map[uint]bool, memberID, targetAuthorID uint) []*responseComment.CommentReplyItem {
	items := make([]*responseComment.CommentReplyItem, 0, len(replies))
	for _, r := range replies {
		items = append(items, &responseComment.CommentReplyItem{
			ID:              idStr(r.ID),
			User:            commentUser(members[r.MemberID]),
			Content:         r.Content,
			LikeCount:       r.LikeCount,
			IsLiked:         likes[r.ID],
			ReplyToID:       optionalIDStr(r.ReplyToID),
			ReplyToNickname: replyToNickname(r.ReplyToMemberID, members),
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
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 先锁评论行：挡掉 commentID=0 的非法请求，并串行化并发点赞——
		// 不锁行时"删点赞记录 + 插点赞记录"在并发下会撞唯一索引 idx_comment_member 导致 500。
		comment, err := s.repo.LockCommentByID(tx, commentID)
		if err != nil {
			return errors.NewWithErr(errors.CodeNotFound, "评论不存在", err)
		}

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
		// like_count 变化后重算该评论热度
		if err := s.repo.RecomputeHeat(tx, commentID); err != nil {
			return err
		}

		// 最新点赞数 = 锁内读到的值 ± 1（持有行锁，无并发扰动，值准确）
		likeCount = comment.LikeCount
		if isLiked {
			likeCount++
		} else {
			likeCount--
		}
		return nil
	})
	if err != nil {
		return nil, err
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
	// 校验目标存在（笔记/攻略），并取作品作者 memberID
	if _, err := s.repo.FindTargetAuthor(req.TargetType, req.TargetID); err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "目标不存在", err)
	}

	// 二级评论：对任何评论的回复都统一存二级，reply_to_id 指向归属的一级评论
	var replyToID, replyToMemberID *uint
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
		// 被回复的用户：缺省取被回复评论的作者
		//（回复一级→一级作者；回复二级→该二级作者，与"回复二级必传"缺省行为一致）
		if req.ReplyToMemberID == nil {
			mid := replied.MemberID
			replyToMemberID = &mid
		} else {
			replyToMemberID = req.ReplyToMemberID
		}
	}

	comment := &entity.Comment{
		TargetType:      req.TargetType,
		TargetID:        req.TargetID,
		MemberID:        memberID,
		Content:         req.Content,
		ReplyToID:       replyToID,
		ReplyToMemberID: replyToMemberID,
		LikeCount:       0,
		Heat:            0,
		Status:          1,
	}
	// 事务：写评论 + 目标 comment_count +1 + 重算新评论热度；
	// 若是回复，父一级评论的 reply_count 变了，同事务重算父评论热度，避免热评榜/回复预览滞后到小时任务
	err := s.db.Transaction(func(tx *gorm.DB) error {
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
			replyToNicknameStr = m.Nickname
		}
	}
	return &responseComment.CommentItem{
		ID:              idStr(comment.ID),
		User:            commentUser(members[memberID]),
		Content:         comment.Content,
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

// ListComments 评论列表：前2条热评（不占分页计数）+ 其余按发布时间倒序分页，total=其余评论总数
func (s *commentService) ListComments(memberID uint, req *requestComment.ListCommentRequest) (*response.PageResponse, error) {
	if req.TargetType != 1 && req.TargetType != 2 {
		return nil, errors.New(errors.CodeInvalidParam, "目标类型错误")
	}
	// 校验目标存在，拿作品作者ID（用于 canDelete）
	targetAuthorID, err := s.repo.FindTargetAuthor(req.TargetType, req.TargetID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "目标不存在", err)
	}

	// 前2条热评
	top, err := s.repo.FindTopComments(req.TargetType, req.TargetID, 2)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询评论失败", err)
	}
	excludeIDs := make([]uint, 0, len(top))
	for _, c := range top {
		excludeIDs = append(excludeIDs, c.ID)
	}

	// 其余按发布时间倒序分页
	offset := (req.Page - 1) * req.Size
	rest, err := s.repo.FindCommentPage(req.TargetType, req.TargetID, excludeIDs, offset, req.Size)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询评论失败", err)
	}
	total, err := s.repo.CountCommentPage(req.TargetType, req.TargetID, excludeIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询评论失败", err)
	}

	// 热评 + 本页其余 = list
	comments := make([]entity.Comment, 0, len(top)+len(rest))
	comments = append(comments, top...)
	comments = append(comments, rest...)

	items, err := s.buildCommentItems(comments, memberID, targetAuthorID)
	if err != nil {
		return nil, err
	}
	return response.NewPageResponse(items, total, req.Page, req.Size), nil
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

	// 每个一级评论取热度最高前2条二级评论：仅当热度最高的二级评论点赞数 >10 才展示
	replyMap := make(map[uint][]entity.Comment, len(commentIDs))
	allReplies := make([]entity.Comment, 0, len(commentIDs)*2)
	for _, c := range comments {
		if replyCounts[c.ID] == 0 {
			continue
		}
		topReplies, err := s.repo.FindTopReplies(c.ID, 2)
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeInternalError, "查询热评失败", err)
		}
		if len(topReplies) > 0 && topReplies[0].LikeCount > 10 {
			replyMap[c.ID] = topReplies
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
			User:            commentUser(members[c.MemberID]),
			Content:         c.Content,
			LikeCount:       c.LikeCount,
			IsLiked:         likes[c.ID],
			ReplyToID:       optionalIDStr(c.ReplyToID),
			ReplyToNickname: replyToNickname(c.ReplyToMemberID, members),
			CreatedAt:       c.CreatedAt.Format(timeFormat),
			Replies:         buildReplyItems(replyMap[c.ID], members, likes, memberID, targetAuthorID),
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
	targetAuthorID, err := s.repo.FindTargetAuthor(comment.TargetType, comment.TargetID)
	if err != nil {
		return errors.NewWithErr(errors.CodeNotFound, "目标不存在", err)
	}
	// 权限：评论者本人 或 作品作者
	if comment.MemberID != memberID && targetAuthorID != memberID {
		return errors.New(errors.CodeForbidden, "无权删除该评论")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
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
		return s.repo.RecomputeHeat(tx, *comment.ReplyToID)
	})
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "删除评论失败", err)
	}
	return nil
}

// ListReplies 二级评论列表：前2条热回复（不占分页计数）+ 其余按发布时间倒序分页，total=其余回复数
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
	targetAuthorID, err := s.repo.FindTargetAuthor(parent.TargetType, parent.TargetID)
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

	// 其余按发布时间倒序分页
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

	items := buildReplyItems(replies, members, likes, memberID, targetAuthorID)
	return response.NewPageResponse(items, total, page, size), nil
}
