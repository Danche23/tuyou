package tag

import (
	"context"
	"math"
	"strconv"

	request "tuyou/internal/model/dto/request/admin/tag"
	response "tuyou/internal/model/dto/response/admin/tag"
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
