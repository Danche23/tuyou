package auth_test

// auth 仓库的"放开 phone 唯一索引"语义测试（真连 MySQL，TEST_DB 门控）。
// 依赖 Task 1 已执行：member.phone 唯一索引已删除，否则同号两条插不进去。
// 用法：
//   TEST_DB="root:123456@tcp(192.168.10.4:3308)/tuyou?charset=utf8mb4&parseTime=True&loc=Local" \
//   go test -run TestRepo -v ./internal/repository/web/auth/

import (
	"context"
	"os"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"tuyou/internal/model/entity"
	repoAuth "tuyou/internal/repository/web/auth"
	appErrors "tuyou/pkg/errors"
)

func TestRepoFindActiveByPhone(t *testing.T) {
	dsn := os.Getenv("TEST_DB")
	if dsn == "" {
		t.Skip("TEST_DB not set; skipping DB test")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	if err := db.AutoMigrate(&entity.Member{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if db.Migrator().HasIndex(&entity.Member{}, "idx_member_phone") {
		t.Skip("member.phone 唯一索引仍存在：请先执行 Task 1 的迁移（DROP INDEX idx_member_phone）")
	}

	phone := "9000000011"
	db.Exec("DELETE FROM member WHERE phone = ?", phone)
	defer db.Exec("DELETE FROM member WHERE phone = ?", phone)

	// 已注销 + 正常两账号共用同一手机号（Option A：注销后手机号可被重新注册）
	cancelled := entity.Member{Nickname: "cancel-test", Phone: phone, Password: "x", Status: 2}
	active := entity.Member{Nickname: "active-test", Phone: phone, Password: "x", Status: 1}
	if err := db.Create(&cancelled).Error; err != nil {
		t.Fatalf("create cancelled: %v", err)
	}
	if err := db.Create(&active).Error; err != nil {
		t.Fatalf("create active: %v", err)
	}

	repo := repoAuth.NewMemberRepository(db)
	ctx := context.Background()

	// FindActiveByPhone 只返回在用的（status<>2）账号
	m, err := repo.FindActiveByPhone(ctx, phone)
	if err != nil {
		t.Fatalf("FindActiveByPhone: %v", err)
	}
	if m.ID != active.ID {
		t.Fatalf("FindActiveByPhone 应返回正常账号(%d)，实际 %d", active.ID, m.ID)
	}

	// FindCancelledByPhone 只返回已注销账号
	c, err := repo.FindCancelledByPhone(ctx, phone)
	if err != nil {
		t.Fatalf("FindCancelledByPhone: %v", err)
	}
	if c.ID != cancelled.ID {
		t.Fatalf("FindCancelledByPhone 应返回已注销账号(%d)，实际 %d", cancelled.ID, c.ID)
	}
}

func TestRepoUpdateStatus(t *testing.T) {
	dsn := os.Getenv("TEST_DB")
	if dsn == "" {
		t.Skip("TEST_DB not set; skipping DB test")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	if err := db.AutoMigrate(&entity.Member{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	phone := "9000000012"
	db.Exec("DELETE FROM member WHERE phone = ?", phone)
	defer db.Exec("DELETE FROM member WHERE phone = ?", phone)

	m := entity.Member{Nickname: "status-test", Phone: phone, Password: "x", Status: 1}
	if err := db.Create(&m).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	repo := repoAuth.NewMemberRepository(db)
	ctx := context.Background()

	if err := repo.UpdateStatus(ctx, m.ID, 2); err != nil {
		t.Fatalf("UpdateStatus(2): %v", err)
	}
	got, err := repo.FindByID(ctx, m.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Status != 2 {
		t.Fatalf("期望 status=2，实际 %d", got.Status)
	}
	// 已注销后，FindActiveByPhone 应查不到
	if _, err := repo.FindActiveByPhone(ctx, phone); err != appErrors.ErrUserNotFound {
		t.Fatalf("期望 ErrUserNotFound，实际 %v", err)
	}
}
