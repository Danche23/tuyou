package sensitiveword

import (
	"context"
	"math"
	"strconv"
	"strings"

	request "tuyou/internal/model/dto/request/admin/sensitiveword"
	response "tuyou/internal/model/dto/response/admin/sensitiveword"
	"tuyou/internal/model/entity"
	repo "tuyou/internal/repository/admin/sensitiveword"
	"tuyou/pkg/errors"
	"tuyou/pkg/sensitive"
)

const timeFormat = "2006-01-02 15:04:05"

type sensitiveWordService struct {
	wordRepo repo.SensitiveWordRepository
	matcher  *sensitive.SensitiveMatcher
}

func NewSensitiveWordService(wordRepo repo.SensitiveWordRepository, matcher *sensitive.SensitiveMatcher) SensitiveWordService {
	return &sensitiveWordService{wordRepo: wordRepo, matcher: matcher}
}

// reload 增删后热更 trie：重新从库取启用词（DB 异常时本次不热更，下次增删再试）
func (s *sensitiveWordService) reload(ctx context.Context) {
	if s.matcher == nil {
		return
	}
	words, err := s.wordRepo.ListActiveWords(ctx)
	if err != nil {
		return
	}
	s.matcher.Reload(words)
}

func (s *sensitiveWordService) List(ctx context.Context, req *request.ListReq) (*response.ListResp, error) {
	filters := make(map[string]interface{})
	if req.Keyword != "" {
		filters["keyword"] = req.Keyword
	}
	total, err := s.wordRepo.Count(ctx, filters)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取敏感词列表失败", err)
	}
	words, err := s.wordRepo.FindList(ctx, filters, req.Page, req.Size)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取敏感词列表失败", err)
	}
	list := make([]response.Item, 0, len(words))
	for _, w := range words {
		list = append(list, response.Item{
			ID:         strconv.FormatUint(uint64(w.ID), 10),
			Word:       w.Word,
			CreateTime: w.CreatedAt.Format(timeFormat),
		})
	}
	return &response.ListResp{
		List: list, Total: total, Page: req.Page, Size: req.Size,
		TotalPage: int(math.Ceil(float64(total) / float64(req.Size))),
	}, nil
}

func (s *sensitiveWordService) Create(ctx context.Context, req *request.CreateReq) error {
	word := strings.TrimSpace(req.Word)
	if word == "" {
		return errors.New(errors.CodeInvalidParam, "敏感词不能为空")
	}
	exists, err := s.wordRepo.ExistsByName(ctx, word)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "新增敏感词失败", err)
	}
	if exists {
		return errors.New(errors.CodeResourceAlreadyExists, "敏感词已存在")
	}
	if err := s.wordRepo.Create(ctx, &entity.SensitiveWord{Word: word, Status: 1}); err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "新增敏感词失败", err)
	}
	s.reload(ctx)
	return nil
}

func (s *sensitiveWordService) Delete(ctx context.Context, id uint) error {
	w, err := s.wordRepo.FindByID(ctx, id)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "获取敏感词失败", err)
	}
	if w == nil {
		return errors.New(errors.CodeNotFound, "敏感词不存在")
	}
	if err := s.wordRepo.Delete(ctx, id); err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "删除敏感词失败", err)
	}
	s.reload(ctx)
	return nil
}
