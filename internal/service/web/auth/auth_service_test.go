package auth_test

// auth 服务分支单测（mock 仓储，不连真 SMS、不依赖 config）。
// 覆盖全部在 sms.Send/sms.Receive/jwt.GenerateToken 之前返回的错误分支；
// 快乐路径（发码成功/验证码校验/注销成功）需要真短信，由 Task 8 冒烟/人工验证。

import (
	"context"
	"testing"
	"time"

	request "tuyou/internal/model/dto/request/web/auth"
	"tuyou/internal/model/entity"
	"tuyou/internal/repository"
	repoAuth "tuyou/internal/repository/web/auth"
	serviceAuth "tuyou/internal/service/web/auth"
	appErrors "tuyou/pkg/errors"
	"tuyou/pkg/sensitive"
)

// mockRedis 只覆盖用到的 Set/Del，其余接口方法由嵌入接口兜底（不会被调用）
type mockRedis struct {
	repository.RedisRepository
}

func (m *mockRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error { return nil }
func (m *mockRedis) Del(ctx context.Context, key string) error                                            { return nil }

// mockMemberRepo 覆盖测试会调用的方法，其余由嵌入接口兜底
type mockMemberRepo struct {
	repoAuth.MemberRepository
	activeByPhone    func(ctx context.Context, phone string) (*entity.Member, error)
	cancelledByPhone func(ctx context.Context, phone string) (*entity.Member, error)
	findByID         func(ctx context.Context, id uint) (*entity.Member, error)
}

func (m *mockMemberRepo) FindActiveByPhone(ctx context.Context, phone string) (*entity.Member, error)    { return m.activeByPhone(ctx, phone) }
func (m *mockMemberRepo) FindCancelledByPhone(ctx context.Context, phone string) (*entity.Member, error) { return m.cancelledByPhone(ctx, phone) }
func (m *mockMemberRepo) FindByID(ctx context.Context, id uint) (*entity.Member, error)                  { return m.findByID(ctx, id) }

func newTestService(repo *mockMemberRepo) serviceAuth.WebAuthService {
	return serviceAuth.NewAuthService(repo, &mockRedis{}, sensitive.New(nil))
}

func TestPasswordLoginRejectsCancelled(t *testing.T) {
	repo := &mockMemberRepo{
		activeByPhone: func(ctx context.Context, phone string) (*entity.Member, error) {
			return nil, appErrors.ErrUserNotFound
		},
		cancelledByPhone: func(ctx context.Context, phone string) (*entity.Member, error) {
			return &entity.Member{BaseEntity: entity.BaseEntity{ID: 1}, Phone: phone, Status: 2}, nil
		},
	}
	svc := newTestService(repo)
	_, err := svc.Login(context.Background(), &request.LoginRequest{Phone: "13800138000", Password: "123456", Type: "password"})
	if err != appErrors.ErrAccountCancelled {
		t.Fatalf("期望 ErrAccountCancelled，实际 %v", err)
	}
}

func TestPasswordLoginRejectsDisabled(t *testing.T) {
	repo := &mockMemberRepo{
		activeByPhone: func(ctx context.Context, phone string) (*entity.Member, error) {
			return &entity.Member{BaseEntity: entity.BaseEntity{ID: 1}, Phone: phone, Status: 0}, nil
		},
	}
	svc := newTestService(repo)
	_, err := svc.Login(context.Background(), &request.LoginRequest{Phone: "13800138000", Password: "123456", Type: "password"})
	if err != appErrors.ErrUserDisabled {
		t.Fatalf("期望 ErrUserDisabled，实际 %v", err)
	}
}

func TestPasswordLoginWrongPassword(t *testing.T) {
	repo := &mockMemberRepo{
		activeByPhone: func(ctx context.Context, phone string) (*entity.Member, error) {
			return &entity.Member{BaseEntity: entity.BaseEntity{ID: 1}, Phone: phone, Status: 1, Password: ""}, nil
		},
	}
	svc := newTestService(repo)
	_, err := svc.Login(context.Background(), &request.LoginRequest{Phone: "13800138000", Password: "123456", Type: "password"})
	if err != appErrors.ErrInvalidCredentials {
		t.Fatalf("期望 ErrInvalidCredentials，实际 %v", err)
	}
}

func TestPasswordLoginNotRegistered(t *testing.T) {
	repo := &mockMemberRepo{
		activeByPhone: func(ctx context.Context, phone string) (*entity.Member, error) {
			return nil, appErrors.ErrUserNotFound
		},
		cancelledByPhone: func(ctx context.Context, phone string) (*entity.Member, error) {
			return nil, appErrors.ErrUserNotFound
		},
	}
	svc := newTestService(repo)
	_, err := svc.Login(context.Background(), &request.LoginRequest{Phone: "13800138000", Password: "123456", Type: "password"})
	if err != appErrors.ErrPhoneNotRegistered {
		t.Fatalf("期望 ErrPhoneNotRegistered，实际 %v", err)
	}
}

func TestSendCodeRegisterOccupied(t *testing.T) {
	repo := &mockMemberRepo{
		activeByPhone: func(ctx context.Context, phone string) (*entity.Member, error) {
			return &entity.Member{BaseEntity: entity.BaseEntity{ID: 1}, Phone: phone, Status: 1}, nil
		},
	}
	svc := newTestService(repo)
	_, err := svc.SendCode(context.Background(), &request.SendCodeRequest{Phone: "13800138000", Scene: "register"})
	if err != appErrors.ErrPhoneAlreadyReg {
		t.Fatalf("期望 ErrPhoneAlreadyReg，实际 %v", err)
	}
}

func TestSendCodeLoginNotRegistered(t *testing.T) {
	repo := &mockMemberRepo{
		activeByPhone: func(ctx context.Context, phone string) (*entity.Member, error) {
			return nil, appErrors.ErrUserNotFound
		},
	}
	svc := newTestService(repo)
	_, err := svc.SendCode(context.Background(), &request.SendCodeRequest{Phone: "13800138000", Scene: "login"})
	if err != appErrors.ErrPhoneNotRegistered {
		t.Fatalf("期望 ErrPhoneNotRegistered，实际 %v", err)
	}
}

func TestSendCodeCancelAccountCancelled(t *testing.T) {
	repo := &mockMemberRepo{
		activeByPhone: func(ctx context.Context, phone string) (*entity.Member, error) {
			return nil, appErrors.ErrUserNotFound
		},
		cancelledByPhone: func(ctx context.Context, phone string) (*entity.Member, error) {
			return &entity.Member{BaseEntity: entity.BaseEntity{ID: 1}, Phone: phone, Status: 2}, nil
		},
	}
	svc := newTestService(repo)
	_, err := svc.SendCode(context.Background(), &request.SendCodeRequest{Phone: "13800138000", Scene: "cancel-account"})
	if err != appErrors.ErrAccountCancelled {
		t.Fatalf("期望 ErrAccountCancelled（已注销拒发），实际 %v", err)
	}
}

func TestSendCodeCancelAccountNotRegistered(t *testing.T) {
	repo := &mockMemberRepo{
		activeByPhone: func(ctx context.Context, phone string) (*entity.Member, error) {
			return nil, appErrors.ErrUserNotFound
		},
		cancelledByPhone: func(ctx context.Context, phone string) (*entity.Member, error) {
			return nil, appErrors.ErrUserNotFound
		},
	}
	svc := newTestService(repo)
	_, err := svc.SendCode(context.Background(), &request.SendCodeRequest{Phone: "13800138000", Scene: "cancel-account"})
	if err != appErrors.ErrPhoneNotRegistered {
		t.Fatalf("期望 ErrPhoneNotRegistered，实际 %v", err)
	}
}
