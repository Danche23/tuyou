package tag

import (
	"context"
	"math"
	"strconv"
	"time"

	request "tuyou/internal/model/dto/request/admin/tag"
	response "tuyou/internal/model/dto/response/admin/tag"
	"tuyou/internal/model/entity"
	repo "tuyou/internal/repository/admin/tag"
	"tuyou/pkg/errors"
)

type tagService struct {
	tagRepo repo.TagRepository
}

// NewTagService 创建后台标签服务实例
func NewTagService(tagRepo repo.TagRepository) TagService {
	return &tagService{tagRepo: tagRepo}
}

// ListTags 分页查询标签列表
func (s *tagService) ListTags(ctx context.Context, req *request.ListTagReq) (*response.ListTagResp, error) {
	// ========== 1. 拼装筛选条件 ==========
	filters := make(map[string]interface{})

	if req.TagName != "" {
		filters["tag_name"] = req.TagName
	}
	if req.Status != "" {
		switch req.Status {
		case "启用":
			filters["status"] = 1
		case "禁用":
			filters["status"] = 0
		default:
			return nil, errors.New(errors.CodeInvalidParam, "状态参数不合法")
		}
	}
	if req.BeginTime != "" {
		filters["begin_time"] = req.BeginTime
	}
	if req.FinishTime != "" {
		filters["finish_time"] = req.FinishTime + " 23:59:59"
	}

	// ========== 2. 查询总数 + 本页标签 ==========
	total, err := s.tagRepo.Count(ctx, filters)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取标签列表失败", err)
	}

	tags, err := s.tagRepo.FindList(ctx, filters, req.Page, req.Size)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取标签列表失败", err)
	}

	// ========== 3. 组装列表项 ==========
	list := make([]response.TagListItem, 0, len(tags))
	for _, t := range tags {
		status := "禁用"
		if t.Status == 1 {
			status = "启用"
		}

		isHot := "否"
		if t.IsHot == 1 {
			isHot = "是"
		}

		deleteTime := ""
		if t.DeletedAt.Valid {
			deleteTime = t.DeletedAt.Time.Format("2006-01-02 15:04:05")
		}

		list = append(list, response.TagListItem{
			ID:         strconv.Itoa(int(t.ID)),
			TagName:    t.TagName,
			Status:     status,
			IsHot:      isHot,
			CreatedAt:  t.CreatedAt.Format("2006-01-02 15:04:05"),
			DeleteTime: deleteTime,
		})
	}

	return &response.ListTagResp{
		List:      list,
		Total:     total,
		Page:      req.Page,
		Size:      req.Size,
		TotalPage: int(math.Ceil(float64(total) / float64(req.Size))),
	}, nil
}

// CreateTag 新增标签（标签名称唯一，重名返回错误）
func (s *tagService) CreateTag(ctx context.Context, req *request.CreateTagReq) error {
	// ========== 1. 状态解析 ==========
	var status int
	switch req.Status {
	case "启用":
		status = 1
	case "禁用":
		status = 0
	default:
		return errors.New(errors.CodeInvalidParam, "状态参数不合法")
	}

	// ========== 2. 重名校验（Unscoped 包含禁用/软删标签，名字仍被占用） ==========
	exists, err := s.tagRepo.ExistsByName(ctx, req.TagName)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "新增标签失败", err)
	}
	if exists {
		return errors.New(errors.CodeResourceAlreadyExists, "标签名称已存在")
	}

	// ========== 3. 插入 ==========
	tag := &entity.Tag{
		TagName: req.TagName,
		Status:  status,
	}
	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "新增标签失败", err)
	}
	return nil
}

// UpdateTag 编辑标签（status/is_hot 必填，标签名称不可修改）
func (s *tagService) UpdateTag(ctx context.Context, id uint, req *request.UpdateTagReq) error {
	// ========== 1. 校验标签存在 ==========
	tag, err := s.tagRepo.FindByID(ctx, id)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "获取标签失败", err)
	}
	if tag == nil {
		return errors.New(errors.CodeNotFound, "标签不存在")
	}

	// ========== 2. 解析必传字段（switch 校验，同笔记模式） ==========
	var status int
	switch req.Status {
	case "启用":
		status = 1
	case "禁用":
		status = 0
	default:
		return errors.New(errors.CodeInvalidParam, "状态参数不合法")
	}

	var isHot int
	switch req.IsHot {
	case "是":
		isHot = 1
	case "否":
		isHot = 0
	default:
		return errors.New(errors.CodeInvalidParam, "是否热门参数不合法")
	}

	// ========== 3. 组装更新字段 ==========
	updates := map[string]interface{}{
		"status": status,
		"is_hot": isHot,
	}
	// 状态变化时同步 delete_time（禁用写删除时间，恢复启用清除，同笔记下架/恢复逻辑）
	if tag.Status != status {
		if status == 0 {
			updates["delete_time"] = time.Now()
		} else {
			updates["delete_time"] = nil
		}
	}

	if err := s.tagRepo.Update(ctx, id, updates); err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "编辑标签失败", err)
	}
	return nil
}

// DeleteTag 删除标签（硬删主行，连带清理 note_tag/strategy_tag 关联）
func (s *tagService) DeleteTag(ctx context.Context, id uint) error {
	// ========== 1. 校验标签存在 ==========
	tag, err := s.tagRepo.FindByID(ctx, id)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "获取标签失败", err)
	}
	if tag == nil {
		return errors.New(errors.CodeNotFound, "标签不存在")
	}

	// ========== 2. 硬删除 ==========
	if err := s.tagRepo.Delete(ctx, id); err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "删除标签失败", err)
	}
	return nil
}

// BatchDeleteTags 批量删除标签（硬删主行，连带清理关联；不存在的 ID 自动跳过）
func (s *tagService) BatchDeleteTags(ctx context.Context, req *request.BatchDeleteTagReq) error {
	// ========== 1. 解析 ID 列表 ==========
	ids := make([]uint, 0, len(req.IDs))
	for _, s := range req.IDs {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return errors.New(errors.CodeInvalidParam, "ID参数不合法")
		}
		ids = append(ids, uint(id))
	}

	// ========== 2. 批量硬删除（单事务） ==========
	if err := s.tagRepo.BatchDelete(ctx, ids); err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "批量删除标签失败", err)
	}
	return nil
}
