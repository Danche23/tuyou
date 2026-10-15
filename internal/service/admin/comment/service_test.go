package comment

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	request "tuyou/internal/model/dto/request/admin/comment"
	"tuyou/internal/model/entity"
	repo "tuyou/internal/repository/admin/comment"
	"tuyou/pkg/errors"
)

// mockRepo 手写 mock 的 CommentRepository，按测试用例注入行为并捕获调用参数
type mockRepo struct {
	countFn              func(ctx context.Context, filters map[string]interface{}) (int64, error)
	findListFn           func(ctx context.Context, filters map[string]interface{}, sortField, sortOrder string, page, size int) ([]entity.Comment, error)
	findByIDFn           func(ctx context.Context, id uint) (*entity.Comment, error)
	findCommentsByIDsFn  func(ctx context.Context, ids []uint) ([]entity.Comment, error)
	findMembersByIDsFn   func(ctx context.Context, memberIDs []uint) (map[uint]*entity.Member, error)
	findNoteTitlesFn     func(ctx context.Context, ids []uint) (map[uint]string, error)
	findStrategyTitlesFn func(ctx context.Context, ids []uint) (map[uint]string, error)
	deleteFn             func(ctx context.Context, comment *entity.Comment) error
	batchDeleteFn        func(ctx context.Context, comments []entity.Comment) error

	// 捕获参数（用于断言）
	gotFilters   map[string]interface{}
	gotSortField string
	gotSortOrder string
}

func (m *mockRepo) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	m.gotFilters = filters
	return m.countFn(ctx, filters)
}

func (m *mockRepo) FindList(ctx context.Context, filters map[string]interface{}, sortField, sortOrder string, page, size int) ([]entity.Comment, error) {
	m.gotFilters = filters
	m.gotSortField = sortField
	m.gotSortOrder = sortOrder
	return m.findListFn(ctx, filters, sortField, sortOrder, page, size)
}

func (m *mockRepo) FindByID(ctx context.Context, id uint) (*entity.Comment, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockRepo) FindCommentsByIDs(ctx context.Context, ids []uint) ([]entity.Comment, error) {
	return m.findCommentsByIDsFn(ctx, ids)
}

func (m *mockRepo) FindMembersByIDs(ctx context.Context, memberIDs []uint) (map[uint]*entity.Member, error) {
	return m.findMembersByIDsFn(ctx, memberIDs)
}

func (m *mockRepo) FindNoteTitles(ctx context.Context, ids []uint) (map[uint]string, error) {
	return m.findNoteTitlesFn(ctx, ids)
}

func (m *mockRepo) FindStrategyTitles(ctx context.Context, ids []uint) (map[uint]string, error) {
	return m.findStrategyTitlesFn(ctx, ids)
}

func (m *mockRepo) Delete(ctx context.Context, comment *entity.Comment) error {
	return m.deleteFn(ctx, comment)
}

func (m *mockRepo) BatchDelete(ctx context.Context, comments []entity.Comment) error {
	return m.batchDeleteFn(ctx, comments)
}

var _ repo.CommentRepository = (*mockRepo)(nil)

// newMockRepo 构造带默认行为的 mock（Count=0、列表空、成员/标题空 map）
func newMockRepo() *mockRepo {
	return &mockRepo{
		countFn:              func(ctx context.Context, filters map[string]interface{}) (int64, error) { return 0, nil },
		findListFn:           func(ctx context.Context, filters map[string]interface{}, sortField, sortOrder string, page, size int) ([]entity.Comment, error) { return []entity.Comment{}, nil },
		findByIDFn:           func(ctx context.Context, id uint) (*entity.Comment, error) { return nil, nil },
		findCommentsByIDsFn:  func(ctx context.Context, ids []uint) ([]entity.Comment, error) { return []entity.Comment{}, nil },
		findMembersByIDsFn:   func(ctx context.Context, memberIDs []uint) (map[uint]*entity.Member, error) { return map[uint]*entity.Member{}, nil },
		findNoteTitlesFn:     func(ctx context.Context, ids []uint) (map[uint]string, error) { return map[uint]string{}, nil },
		findStrategyTitlesFn: func(ctx context.Context, ids []uint) (map[uint]string, error) { return map[uint]string{}, nil },
		deleteFn:             func(ctx context.Context, comment *entity.Comment) error { return nil },
		batchDeleteFn:        func(ctx context.Context, comments []entity.Comment) error { return nil },
	}
}

func ptr(v uint) *uint { return &v }

func newSvc(m *mockRepo) CommentService { return NewCommentService(m, nil) }

// 校验错误码
func codeOf(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		t.Fatal("期望报错，实际返回 nil")
	}
	be, ok := err.(*errors.BizError)
	if !ok {
		t.Fatalf("错误类型不是 *BizError: %T", err)
	}
	return be.Code
}

func TestListComments_BuildItems(t *testing.T) {
	m := newMockRepo()
	m.countFn = func(ctx context.Context, filters map[string]interface{}) (int64, error) { return 2, nil }
	m.findListFn = func(ctx context.Context, filters map[string]interface{}, sortField, sortOrder string, page, size int) ([]entity.Comment, error) {
		deletedAt := gorm.DeletedAt{Valid: true, Time: time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)}
		return []entity.Comment{
			// 一级、软删、笔记目标
			{BaseEntity: entity.BaseEntity{ID: 1, CreatedAt: time.Date(2026, 7, 20, 14, 30, 0, 0, time.UTC), DeletedAt: deletedAt},
				TargetType: 1, TargetID: 9, MemberID: 3, Content: "一级评论", LikeCount: 28, Heat: 86},
			// 二级、正常、攻略目标，被回复的是用户3
			{BaseEntity: entity.BaseEntity{ID: 2, CreatedAt: time.Date(2026, 7, 21, 10, 0, 0, 0, time.UTC)},
				TargetType: 2, TargetID: 5, MemberID: 4, Content: "二级评论", ReplyToID: ptr(1), ReplyToMemberID: ptr(3), LikeCount: 5, Heat: 12},
		}, nil
	}
	m.findMembersByIDsFn = func(ctx context.Context, memberIDs []uint) (map[uint]*entity.Member, error) {
		return map[uint]*entity.Member{
			3: {BaseEntity: entity.BaseEntity{ID: 3}, Nickname: "张三", Avatar: "http://avatar/3"},
			4: {BaseEntity: entity.BaseEntity{ID: 4}, Nickname: "李四"},
		}, nil
	}
	m.findNoteTitlesFn = func(ctx context.Context, ids []uint) (map[uint]string, error) { return map[uint]string{9: "稻城亚丁徒步攻略笔记"}, nil }
	m.findStrategyTitlesFn = func(ctx context.Context, ids []uint) (map[uint]string, error) { return map[uint]string{5: "武功山两日游攻略"}, nil }

	resp, err := newSvc(m).ListComments(context.Background(), &request.ListCommentReq{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("ListComments 报错: %v", err)
	}
	if resp.Total != 2 || len(resp.List) != 2 {
		t.Fatalf("期望 total=2 list=2，实际 total=%d list=%d", resp.Total, len(resp.List))
	}
	if resp.TotalPage != 1 {
		t.Fatalf("期望 total_page=1，实际 %d", resp.TotalPage)
	}

	// 第一条：一级 + 软删状态 + 笔记标题
	first := resp.List[0]
	if first.ID != "1" {
		t.Fatalf("id 期望纯数字\"1\"，实际 %q", first.ID)
	}
	if first.Level != "一级" || first.TargetType != "笔记" {
		t.Fatalf("一级评论 level=%q target_type=%q", first.Level, first.TargetType)
	}
	if first.Status != "已删除" || first.DeleteTime == "" {
		t.Fatalf("软删评论 status=%q delete_time=%q", first.Status, first.DeleteTime)
	}
	if first.TargetTitle != "稻城亚丁徒步攻略笔记" || first.Nickname != "张三" || first.Avatar != "http://avatar/3" {
		t.Fatalf("组装错误: title=%q nickname=%q avatar=%q", first.TargetTitle, first.Nickname, first.Avatar)
	}

	// 第二条：二级 + 正常 + 被回复昵称
	second := resp.List[1]
	if second.Level != "二级" || second.TargetType != "攻略" {
		t.Fatalf("二级评论 level=%q target_type=%q", second.Level, second.TargetType)
	}
	if second.Status != "正常" || second.DeleteTime != "" {
		t.Fatalf("正常评论 status=%q delete_time=%q", second.Status, second.DeleteTime)
	}
	if second.ReplyToNickname != "张三" {
		t.Fatalf("被回复昵称期望\"张三\"，实际 %q", second.ReplyToNickname)
	}
	if second.TargetTitle != "武功山两日游攻略" {
		t.Fatalf("攻略标题期望\"武功山两日游攻略\"，实际 %q", second.TargetTitle)
	}
}

func TestListComments_InvalidParams(t *testing.T) {
	cases := []struct {
		name string
		req  request.ListCommentReq
		code int
	}{
		{"非法目标类型", request.ListCommentReq{Page: 1, Size: 10, TargetType: "图片"}, errors.CodeInvalidParam},
		{"非法层级", request.ListCommentReq{Page: 1, Size: 10, Level: "三级"}, errors.CodeInvalidParam},
		{"非法排序方式", request.ListCommentReq{Page: 1, Size: 10, SortOrder: "随机"}, errors.CodeInvalidParam},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := newMockRepo()
			_, err := newSvc(m).ListComments(context.Background(), &c.req)
			if got := codeOf(t, err); got != c.code {
				t.Fatalf("期望错误码 %d，实际 %d", c.code, got)
			}
		})
	}
}

func TestListComments_SortCapture(t *testing.T) {
	m := newMockRepo()
	svc := newSvc(m)

	// 默认：不传 sort → 空字段 + ASC（仓储默认 id 升序）
	if _, err := svc.ListComments(context.Background(), &request.ListCommentReq{Page: 1, Size: 10}); err != nil {
		t.Fatalf("ListComments 报错: %v", err)
	}
	if m.gotSortField != "" || m.gotSortOrder != "ASC" {
		t.Fatalf("默认排序期望 (\"\",ASC)，实际 (%q,%q)", m.gotSortField, m.gotSortOrder)
	}

	// likes + 降序 → 映射为 like_count + DESC
	if _, err := svc.ListComments(context.Background(), &request.ListCommentReq{Page: 1, Size: 10, SortField: "likes", SortOrder: "降序"}); err != nil {
		t.Fatalf("ListComments 报错: %v", err)
	}
	if m.gotSortField != "like_count" || m.gotSortOrder != "DESC" {
		t.Fatalf("likes/降序 期望 (like_count,DESC)，实际 (%q,%q)", m.gotSortField, m.gotSortOrder)
	}
}

func TestListComments_FilterCapture(t *testing.T) {
	m := newMockRepo()
	svc := newSvc(m)
	if _, err := svc.ListComments(context.Background(), &request.ListCommentReq{
		Page: 1, Size: 10, TargetType: "笔记", Nickname: "张", Title: "稻城",
		BeginTime: "2026-07-01", FinishTime: "2026-07-31", Level: "二级",
	}); err != nil {
		t.Fatalf("ListComments 报错: %v", err)
	}
	if m.gotFilters["target_type"] != 1 {
		t.Fatalf("target_type 期望 1，实际 %v", m.gotFilters["target_type"])
	}
	if m.gotFilters["nickname"] != "张" || m.gotFilters["title"] != "稻城" {
		t.Fatalf("nickname/title 筛选错误: %v", m.gotFilters)
	}
	if m.gotFilters["begin_time"] != "2026-07-01" || m.gotFilters["finish_time"] != "2026-07-31 23:59:59" {
		t.Fatalf("时间筛选错误: %v", m.gotFilters)
	}
	if m.gotFilters["level"] != "二级" {
		t.Fatalf("level 期望\"二级\"，实际 %v", m.gotFilters["level"])
	}
}

func TestListComments_Pagination(t *testing.T) {
	m := newMockRepo()
	m.countFn = func(ctx context.Context, filters map[string]interface{}) (int64, error) { return 25, nil }
	resp, err := newSvc(m).ListComments(context.Background(), &request.ListCommentReq{Page: 3, Size: 10})
	if err != nil {
		t.Fatalf("ListComments 报错: %v", err)
	}
	if resp.Total != 25 || resp.Page != 3 || resp.Size != 10 || resp.TotalPage != 3 {
		t.Fatalf("分页字段错误: total=%d page=%d size=%d total_page=%d", resp.Total, resp.Page, resp.Size, resp.TotalPage)
	}
}

func TestGetCommentDetail_NotFound(t *testing.T) {
	m := newMockRepo() // findByIDFn 默认返回 nil,nil
	_, err := newSvc(m).GetCommentDetail(context.Background(), 999)
	if got := codeOf(t, err); got != errors.CodeNotFound {
		t.Fatalf("期望错误码 %d，实际 %d", errors.CodeNotFound, got)
	}
}

func TestDeleteComment_NotFound(t *testing.T) {
	m := newMockRepo()
	err := newSvc(m).DeleteComment(context.Background(), 999)
	if got := codeOf(t, err); got != errors.CodeNotFound {
		t.Fatalf("期望错误码 %d，实际 %d", errors.CodeNotFound, got)
	}
}

func TestBatchDeleteComments_InvalidID(t *testing.T) {
	m := newMockRepo()
	err := newSvc(m).BatchDeleteComments(context.Background(), &request.BatchDeleteReq{IDs: []string{"1", "abc"}})
	if got := codeOf(t, err); got != errors.CodeInvalidParam {
		t.Fatalf("期望错误码 %d，实际 %d", errors.CodeInvalidParam, got)
	}
}

// sensitiveStubRepo 嵌入接口：其余方法未实现，一旦被调用即 panic——证明 buildItems 只走这 3 个查询
type sensitiveStubRepo struct {
	repo.CommentRepository
}

func (sensitiveStubRepo) FindMembersByIDs(ctx context.Context, ids []uint) (map[uint]*entity.Member, error) {
	return map[uint]*entity.Member{27: {BaseEntity: entity.BaseEntity{ID: 27}, Nickname: "测试甲"}}, nil
}
func (sensitiveStubRepo) FindNoteTitles(ctx context.Context, ids []uint) (map[uint]string, error) {
	return map[uint]string{9: "笔记标题"}, nil
}
func (sensitiveStubRepo) FindStrategyTitles(ctx context.Context, ids []uint) (map[uint]string, error) {
	return map[uint]string{}, nil
}

func TestBuildItems_SensitiveFields(t *testing.T) {
	svc := commentService{commentRepo: sensitiveStubRepo{}}
	items, err := svc.buildItems(context.Background(), []entity.Comment{{
		BaseEntity:  entity.BaseEntity{ID: 1},
		TargetType:  1, TargetID: 9, MemberID: 27,
		Content: "赌博网站", ReplyToMemberID: nil,
		IsSensitive: 1, HitWords: "赌博",
	}})
	if err != nil {
		t.Fatalf("buildItems err: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("应组装 1 条，got %d", len(items))
	}
	it := items[0]
	if it.Content != "赌博网站" {
		t.Fatalf("后台应返回原文，got %q", it.Content)
	}
	if it.IsSensitive != 1 {
		t.Fatalf("应带 is_sensitive，got %d", it.IsSensitive)
	}
	if len(it.HitWords) != 1 || it.HitWords[0] != "赌博" {
		t.Fatalf("hit_words 应为数组 [赌博]，got %v", it.HitWords)
	}
}

// TestSplitHitWords 拆分函数单测：多词、空串、含空段三种情况
func TestSplitHitWords(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"赌博;色情", []string{"赌博", "色情"}},
		{"赌博", []string{"赌博"}},
		{"", []string{}},
		{"赌博;;色情", []string{"赌博", "色情"}},
	}
	for _, c := range cases {
		got := splitHitWords(c.in)
		if len(got) != len(c.want) {
			t.Fatalf("splitHitWords(%q) = %v，期望 %v", c.in, got, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Fatalf("splitHitWords(%q) = %v，期望 %v", c.in, got, c.want)
			}
		}
	}
}
