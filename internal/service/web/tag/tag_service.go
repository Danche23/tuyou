package tag

import (
	"context"
	"strconv"

	response "tuyou/internal/model/dto/response/web/tag"
	repo "tuyou/internal/repository/web/tag"
	"tuyou/pkg/errors"
)

type tagService struct {
	tagRepo repo.TagRepository
}

// NewTagService 创建标签服务实例
func NewTagService(tagRepo repo.TagRepository) TagService {
	return &tagService{tagRepo: tagRepo}
}

// ListTags 获取启用标签列表
func (s *tagService) ListTags(ctx context.Context) ([]response.TagItem, error) {
	tags, err := s.tagRepo.ListEnabledTags(ctx)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取标签列表失败", err)
	}
	list := make([]response.TagItem, 0, len(tags))
	for _, t := range tags {
		list = append(list, response.TagItem{ID: strconv.FormatUint(uint64(t.ID), 10), TagName: t.TagName})
	}
	return list, nil
}
