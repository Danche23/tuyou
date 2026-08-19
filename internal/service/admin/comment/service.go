package comment

// 后台评论管理模块 —— 服务层实现（管理端业务逻辑）
//
// 本文件做的事可以概括为"翻译 + 组装"：
//  1. 翻译：把 postman 里定的中文枚举翻译成仓储能用的值——
//     target_type（笔记/攻略→1/2）、level（一级/二级）、sort_order（升序/降序→ASC/DESC）、
//     sort_field（likes/heat→like_count/heat，白名单防 SQL 注入）。
//  2. 组装（buildItems）：评论行只存了 member_id / target_id / reply_to_id，
//     要展示给运营看，就得跨表补出：昵称、头像、目标标题、被回复人昵称、层级、状态。
//     这里全部用"批量查一次"（FindMembersByIDs / FindNoteTitles / FindStrategyTitles），
//     避免对每条评论各查一次（防 N+1）。
//  3. 删除语义：后台删 = 物理硬删，直接调仓储的事务方法，
//     comment_count 的扣减逻辑在仓储层（deleteCommentInTx），service 只负责查出来、传进去。
//
// 错误约定：所有失败返回 *errors.BizError，带业务错误码——
//   参数不合法 CodeInvalidParam(2001)、不存在 CodeNotFound(404)、内部错误 CodeInternalError(500)。
import (
	"context"
	"math"
	"strconv"
	"strings"

	"gorm.io/gorm"

	request "tuyou/internal/model/dto/request/admin/comment"
	response "tuyou/internal/model/dto/response/admin/comment"
	"tuyou/internal/model/entity"
	repo "tuyou/internal/repository/admin/comment"
	"tuyou/pkg/errors"
)

// timeFormat 时间格式化样式（与项目其他模块一致）
const timeFormat = "2006-01-02 15:04:05"

type commentService struct {
	commentRepo repo.CommentRepository
	db          *gorm.DB
}

// NewCommentService 创建后台评论服务实例
func NewCommentService(commentRepo repo.CommentRepository, db *gorm.DB) CommentService {
	return &commentService{commentRepo: commentRepo, db: db}
}

// 目标类型 中文 → 数字（1=笔记，2=攻略）
var targetTypeStrToInt = map[string]int{"笔记": 1, "攻略": 2}

// 目标类型 数字 → 中文
var targetTypeIntToStr = map[int]string{1: "笔记", 2: "攻略"}

// 排序字段中文 → 数据库列名（白名单，防注入）
var commentSortFieldMap = map[string]string{
	"likes": "like_count",
	"heat":  "heat",
}

// ListComments 分页查询评论列表
func (s *commentService) ListComments(ctx context.Context, req *request.ListCommentReq) (*response.ListCommentResp, error) {
	// ========== 1. 拼装筛选条件 ==========
	filters := make(map[string]interface{})

	if req.TargetType != "" {
		t, ok := targetTypeStrToInt[req.TargetType]
		if !ok {
			return nil, errors.New(errors.CodeInvalidParam, "目标类型不合法")
		}
		filters["target_type"] = t
	}
	if req.Nickname != "" {
		filters["nickname"] = req.Nickname
	}
	if req.Title != "" {
		filters["title"] = req.Title
	}
	if req.BeginTime != "" {
		filters["begin_time"] = req.BeginTime
	}
	if req.FinishTime != "" {
		filters["finish_time"] = req.FinishTime + " 23:59:59"
	}
	if req.Level != "" {
		if req.Level != "一级" && req.Level != "二级" {
			return nil, errors.New(errors.CodeInvalidParam, "评论层级不合法")
		}
		filters["level"] = req.Level
	}

	// ========== 2. 排序解析 ==========
	sortField := commentSortFieldMap[req.SortField] // 不在白名单 → 空串（仓储默认 id 升序）
	sortOrder := "ASC"
	if req.SortOrder != "" {
		switch req.SortOrder {
		case "升序":
			sortOrder = "ASC"
		case "降序":
			sortOrder = "DESC"
		default:
			return nil, errors.New(errors.CodeInvalidParam, "排序方式不合法")
		}
	}

	// ========== 3. 查询总数 + 本页评论 ==========
	total, err := s.commentRepo.Count(ctx, filters)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取评论列表失败", err)
	}

	comments, err := s.commentRepo.FindList(ctx, filters, sortField, sortOrder, req.Page, req.Size)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取评论列表失败", err)
	}

	items, err := s.buildItems(ctx, comments)
	if err != nil {
		return nil, err
	}

	return &response.ListCommentResp{
		List:      items,
		Total:     total,
		Page:      req.Page,
		Size:      req.Size,
		TotalPage: int(math.Ceil(float64(total) / float64(req.Size))),
	}, nil
}

// splitHitWords 把存储用的分号分隔字符串拆成响应数组。
// 空/未命中返回空数组而非 null，保证 JSON 出参稳定为 []（前端可直接 .join 渲染）。
func splitHitWords(hitWords string) []string {
	if hitWords == "" {
		return []string{}
	}
	parts := strings.Split(hitWords, ";")
	words := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			words = append(words, p)
		}
	}
	return words
}

// buildItems 组装列表项（批量查用户、目标标题、被回复用户昵称，避免 N+1）
func (s *commentService) buildItems(ctx context.Context, comments []entity.Comment) ([]response.CommentItem, error) {
	items := make([]response.CommentItem, 0, len(comments))
	if len(comments) == 0 {
		return items, nil
	}

	// 收集评论者 + 被回复用户（去重）、目标 ID（按类型分表）
	memberIDs := make([]uint, 0, len(comments)*2)
	memberIDSet := make(map[uint]bool)
	noteIDs := make([]uint, 0)
	strategyIDs := make([]uint, 0)
	for _, c := range comments {
		if !memberIDSet[c.MemberID] {
			memberIDSet[c.MemberID] = true
			memberIDs = append(memberIDs, c.MemberID)
		}
		if c.ReplyToMemberID != nil && !memberIDSet[*c.ReplyToMemberID] {
			memberIDSet[*c.ReplyToMemberID] = true
			memberIDs = append(memberIDs, *c.ReplyToMemberID)
		}
		if c.TargetType == 1 {
			noteIDs = append(noteIDs, c.TargetID)
		} else {
			strategyIDs = append(strategyIDs, c.TargetID)
		}
	}

	members, err := s.commentRepo.FindMembersByIDs(ctx, memberIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询用户信息失败", err)
	}
	noteTitles, err := s.commentRepo.FindNoteTitles(ctx, noteIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询笔记标题失败", err)
	}
	strategyTitles, err := s.commentRepo.FindStrategyTitles(ctx, strategyIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询攻略标题失败", err)
	}

	for _, c := range comments {
		targetTitle := ""
		if c.TargetType == 1 {
			targetTitle = noteTitles[c.TargetID]
		} else {
			targetTitle = strategyTitles[c.TargetID]
		}

		level := "一级"
		if c.ReplyToID != nil {
			level = "二级"
		}

		replyToNickname := ""
		if c.ReplyToMemberID != nil {
			if m := members[*c.ReplyToMemberID]; m != nil {
				replyToNickname = m.Nickname
			}
		}

		// 状态为派生值：delete_time 有值 → 已删除；空 → 正常
		status := "正常"
		deleteTime := ""
		if c.DeletedAt.Valid {
			status = "已删除"
			deleteTime = c.DeletedAt.Time.Format(timeFormat)
		}

		nickname := ""
		avatar := ""
		if m := members[c.MemberID]; m != nil {
			nickname = m.Nickname
			avatar = m.Avatar
		}

		items = append(items, response.CommentItem{
			ID:              strconv.FormatUint(uint64(c.ID), 10),
			Nickname:        nickname,
			Avatar:          avatar,
			TargetType:      targetTypeIntToStr[c.TargetType],
			TargetTitle:     targetTitle,
			Content:         c.Content,
			Level:           level,
			ReplyToNickname: replyToNickname,
			LikeCount:       c.LikeCount,
			Heat:            c.Heat,
			CreateTime:      c.CreatedAt.Format(timeFormat),
			Status:          status,
			IsSensitive:     c.IsSensitive,
			HitWords:        splitHitWords(c.HitWords),
			DeleteTime:      deleteTime,
		})
	}
	return items, nil
}

// GetCommentDetail 获取评论详情
func (s *commentService) GetCommentDetail(ctx context.Context, id uint) (*response.CommentItem, error) {
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取评论详情失败", err)
	}
	if comment == nil {
		return nil, errors.New(errors.CodeNotFound, "评论不存在")
	}
	items, err := s.buildItems(ctx, []entity.Comment{*comment})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New(errors.CodeNotFound, "评论不存在")
	}
	return &items[0], nil
}

// DeleteComment 彻底删除评论（一级连带删除其下所有二级）
func (s *commentService) DeleteComment(ctx context.Context, id uint) error {
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "获取评论失败", err)
	}
	if comment == nil {
		return errors.New(errors.CodeNotFound, "评论不存在")
	}
	if err := s.commentRepo.Delete(ctx, comment); err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "删除评论失败", err)
	}
	return nil
}

// BatchDeleteComments 批量彻底删除评论（不存在的 ID 自动跳过）
func (s *commentService) BatchDeleteComments(ctx context.Context, req *request.BatchDeleteReq) error {
	ids, err := parseCommentIDs(req.IDs)
	if err != nil {
		return errors.New(errors.CodeInvalidParam, "ID参数不合法")
	}
	// 一次查出存在的评论（IN 子句自动去重，重复/不存在的 ID 自动跳过）
	comments, err := s.commentRepo.FindCommentsByIDs(ctx, ids)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "获取评论失败", err)
	}
	if len(comments) == 0 {
		return nil
	}
	if err := s.commentRepo.BatchDelete(ctx, comments); err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "批量删除评论失败", err)
	}
	return nil
}

// parseCommentIDs 解析纯数字评论 ID 列表
func parseCommentIDs(ids []string) ([]uint, error) {
	result := make([]uint, 0, len(ids))
	for _, s := range ids {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		result = append(result, uint(id))
	}
	return result, nil
}
