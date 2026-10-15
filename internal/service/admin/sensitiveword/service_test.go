package sensitiveword

import (
	"context"
	"testing"

	request "tuyou/internal/model/dto/request/admin/sensitiveword"
	repo "tuyou/internal/repository/admin/sensitiveword"
	"tuyou/internal/model/entity"
	"tuyou/pkg/errors"
	"tuyou/pkg/sensitive"
)

// stubRepo 内存版：ListActiveWords 结果可控，用来断言增删后 trie 是否热更
type stubRepo struct {
	repo.SensitiveWordRepository
	activeWords []string
	exists      bool
	created     *entity.SensitiveWord
	deleted     bool
}

func (s *stubRepo) ExistsByName(ctx context.Context, word string) (bool, error) { return s.exists, nil }
func (s *stubRepo) Create(ctx context.Context, w *entity.SensitiveWord) error {
	s.created = w
	s.activeWords = append(s.activeWords, w.Word)
	return nil
}
func (s *stubRepo) Delete(ctx context.Context, id uint) error {
	s.deleted = true
	s.activeWords = nil // 模拟删空
	return nil
}
func (s *stubRepo) ListActiveWords(ctx context.Context) ([]string, error) { return s.activeWords, nil }
func (s *stubRepo) FindByID(ctx context.Context, id uint) (*entity.SensitiveWord, error) {
	// 计划补丁：服务 Delete 先按 id 查存在性，stub 需覆盖该方法，否则嵌入的 nil 接口会 panic
	if len(s.activeWords) > 0 {
		return &entity.SensitiveWord{BaseEntity: entity.BaseEntity{ID: id}, Word: s.activeWords[0], Status: 1}, nil
	}
	return nil, nil
}

func TestCreate_ReloadsTrie(t *testing.T) {
	matcher := sensitive.New([]string{"旧词"})
	stub := &stubRepo{activeWords: []string{"旧词"}}
	svc := NewSensitiveWordService(stub, matcher)

	if err := svc.Create(context.Background(), &request.CreateReq{Word: "新词"}); err != nil {
		t.Fatalf("Create err: %v", err)
	}
	if _, hit := matcher.Validate("这里新词出现"); !hit {
		t.Fatal("新增后 trie 应热更命中 新词")
	}
}

func TestCreate_DuplicateRejected(t *testing.T) {
	stub := &stubRepo{exists: true, activeWords: []string{"重复"}}
	svc := NewSensitiveWordService(stub, sensitive.New(nil))
	err := svc.Create(context.Background(), &request.CreateReq{Word: "重复"})
	be, ok := err.(*errors.BizError)
	if !ok || be.Code != errors.CodeResourceAlreadyExists {
		t.Fatalf("重名应返 CodeResourceAlreadyExists，got %v", err)
	}
}

func TestDelete_ReloadsTrie(t *testing.T) {
	matcher := sensitive.New([]string{"旧词"})
	stub := &stubRepo{activeWords: []string{"旧词"}}
	svc := NewSensitiveWordService(stub, matcher)

	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete err: %v", err)
	}
	if _, hit := matcher.Validate("旧词内容"); hit {
		t.Fatal("删除后 trie 应热更移除 旧词")
	}
}
