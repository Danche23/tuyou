package member_test

// admin 会员服务的三态/守卫/恢复单测（mock 仓库，不需要 DB）。
// 覆盖：status 筛选传参、非法 status、UpdateMember 对已注销会员守卫、Restore 三个分支。

import (
	"context"
	"strings"
	"testing"

	"tuyou/internal/model/entity"
	request "tuyou/internal/model/dto/request/admin/member"
	response "tuyou/internal/model/dto/response/admin/member"
	repoMember "tuyou/internal/repository/admin/member"
	serviceMember "tuyou/internal/service/admin/member"
	appErrors "tuyou/pkg/errors"
)

// mockAdminRepo 覆盖测试会调用的方法，其余由嵌入接口兜底
type mockAdminRepo struct {
	repoMember.MemberRepository
	count                func(ctx context.Context, filters map[string]interface{}) (int64, error)
	list                 func(ctx context.Context, filters map[string]interface{}, page, size int) ([]response.MemberListItem, error)
	getDetail            func(ctx context.Context, id uint) (*response.MemberDetail, error)
	update               func(ctx context.Context, id uint, updates map[string]interface{}) error
	findActiveByPhoneExc func(ctx context.Context, phone string, exceptID uint) (*entity.Member, error)
}

func (m *mockAdminRepo) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	return m.count(ctx, filters)
}
func (m *mockAdminRepo) List(ctx context.Context, filters map[string]interface{}, page, size int) ([]response.MemberListItem, error) {
	return m.list(ctx, filters, page, size)
}
func (m *mockAdminRepo) GetDetailByID(ctx context.Context, id uint) (*response.MemberDetail, error) {
	return m.getDetail(ctx, id)
}
func (m *mockAdminRepo) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return m.update(ctx, id, updates)
}
func (m *mockAdminRepo) FindActiveByPhoneExcept(ctx context.Context, phone string, exceptID uint) (*entity.Member, error) {
	return m.findActiveByPhoneExc(ctx, phone, exceptID)
}

func newTestAdminService(repo *mockAdminRepo) serviceMember.MemberService {
	return serviceMember.NewMemberService(repo, nil)
}

func TestListMembersStatusFilter(t *testing.T) {
	var gotFilters map[string]interface{}
	repo := &mockAdminRepo{
		count: func(ctx context.Context, filters map[string]interface{}) (int64, error) {
			gotFilters = filters
			return 0, nil
		},
		list: func(ctx context.Context, filters map[string]interface{}, page, size int) ([]response.MemberListItem, error) {
			return []response.MemberListItem{}, nil
		},
	}
	role := 1
	svc := newTestAdminService(repo)
	_, err := svc.ListMembers(context.Background(), &role, &request.ListMemberReq{Status: "2", Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	if gotFilters["status"] != 2 {
		t.Fatalf("期望 status 筛选=2，实际 %v", gotFilters["status"])
	}
}

func TestListMembersInvalidStatus(t *testing.T) {
	repo := &mockAdminRepo{}
	role := 1
	svc := newTestAdminService(repo)
	_, err := svc.ListMembers(context.Background(), &role, &request.ListMemberReq{Status: "abc", Page: 1, Size: 10})
	if err != appErrors.ErrInvalidParam {
		t.Fatalf("期望 ErrInvalidParam，实际 %v", err)
	}
}

func TestUpdateMemberGuardCancelled(t *testing.T) {
	statusEnable := "启用"
	repo := &mockAdminRepo{
		getDetail: func(ctx context.Context, id uint) (*response.MemberDetail, error) {
			return &response.MemberDetail{ID: "1", Status: "已注销"}, nil
		},
	}
	svc := newTestAdminService(repo)
	err := svc.UpdateMember(context.Background(), 1, &request.UpdateMemberReq{Status: &statusEnable})
	if err == nil || !strings.Contains(err.Error(), "已注销会员请使用恢复接口") {
		t.Fatalf("期望'已注销会员请使用恢复接口'，实际 %v", err)
	}
}

func TestRestoreNotCancelled(t *testing.T) {
	repo := &mockAdminRepo{
		getDetail: func(ctx context.Context, id uint) (*response.MemberDetail, error) {
			return &response.MemberDetail{ID: "1", Status: "启用"}, nil
		},
	}
	svc := newTestAdminService(repo)
	err := svc.Restore(context.Background(), 1)
	if err == nil || !strings.Contains(err.Error(), "该会员未注销") {
		t.Fatalf("期望'该会员未注销，无需恢复'，实际 %v", err)
	}
}

func TestRestorePhoneOccupied(t *testing.T) {
	repo := &mockAdminRepo{
		getDetail: func(ctx context.Context, id uint) (*response.MemberDetail, error) {
			return &response.MemberDetail{ID: "1", Phone: "13800138000", Status: "已注销"}, nil
		},
		findActiveByPhoneExc: func(ctx context.Context, phone string, exceptID uint) (*entity.Member, error) {
			return &entity.Member{BaseEntity: entity.BaseEntity{ID: 2}, Phone: phone, Status: 1}, nil
		},
	}
	svc := newTestAdminService(repo)
	err := svc.Restore(context.Background(), 1)
	if err == nil || !strings.Contains(err.Error(), "手机号已被占用") {
		t.Fatalf("期望'手机号已被占用'，实际 %v", err)
	}
}

func TestRestoreSuccess(t *testing.T) {
	var gotID uint
	var gotUpdates map[string]interface{}
	repo := &mockAdminRepo{
		getDetail: func(ctx context.Context, id uint) (*response.MemberDetail, error) {
			return &response.MemberDetail{ID: "1", Phone: "13800138000", Status: "已注销"}, nil
		},
		findActiveByPhoneExc: func(ctx context.Context, phone string, exceptID uint) (*entity.Member, error) {
			return nil, appErrors.ErrUserNotFound // 无在用账号占用，允许恢复
		},
		update: func(ctx context.Context, id uint, updates map[string]interface{}) error {
			gotID = id
			gotUpdates = updates
			return nil
		},
	}
	svc := newTestAdminService(repo)
	if err := svc.Restore(context.Background(), 1); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if gotID != 1 || gotUpdates["status"] != 1 {
		t.Fatalf("期望 Update(1, {status:1})，实际 Update(%d, %v)", gotID, gotUpdates)
	}
}
