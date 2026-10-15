package personal_test

// 方案C 关注翻转的真实并发测试（需真连 MySQL，DB 门控）。
// 用法：TEST_DB="root:密码@tcp(127.0.0.1:3306)/tuyou_test?charset=utf8mb4&parseTime=True&loc=Local" \
//        go test -run TestToggleFollowConcurrency ./internal/service/web/personal/
// 未设置 TEST_DB 时自动 Skip（CI 安全）。
// 断言三个不变量：
//   1. 100 个并发请求全程无错误（不撞唯一索引、不 500）
//   2. follow 表中 A→B 只有一行
//   3. member.follow_count / fans_count 与 follow 表 status=1 实际行数一致（计数不漂移）

import (
	"net"
	"os"
	"sync"
	"testing"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"tuyou/internal/model/entity"
	repoMsg "tuyou/internal/repository/web/message"
	repoPersonal "tuyou/internal/repository/web/personal"
	serviceMsg "tuyou/internal/service/web/message"
	servicePersonal "tuyou/internal/service/web/personal"
	"tuyou/internal/stream"
	"tuyou/pkg/sensitive"
)

func TestToggleFollowConcurrency(t *testing.T) {
	dsn := os.Getenv("TEST_DB")
	if dsn == "" {
		t.Skip("TEST_DB not set; skipping DB concurrency test")
	}
	parsed, err := mysqlDriver.ParseDSN(dsn)
	if err != nil || parsed.DBName != "tuyou_test" || parsed.Net != "tcp" {
		t.Fatal("本测试会删除 follow 数据；TEST_DB 必须指向本机独立的 tuyou_test 数据库")
	}
	host, _, err := net.SplitHostPort(parsed.Addr)
	if err != nil || (host != "127.0.0.1" && host != "localhost" && host != "::1") {
		t.Fatal("TEST_DB 只允许连接本机 MySQL")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	if err := db.AutoMigrate(&entity.Member{}, &entity.Follow{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// 清场 + 建两个测试账号（member 的 phone/password 非空，raw Exec 硬删绕开软删作用域）
	db.Exec("DELETE FROM follow")
	db.Exec("DELETE FROM member WHERE phone IN ('9000000001','9000000002')")
	a := entity.Member{Nickname: "follow-test-a", Phone: "9000000001", Password: "x"}
	b := entity.Member{Nickname: "follow-test-b", Phone: "9000000002", Password: "x"}
	if err := db.Create(&a).Error; err != nil {
		t.Fatalf("create member A: %v", err)
	}
	if err := db.Create(&b).Error; err != nil {
		t.Fatalf("create member B: %v", err)
	}
	defer db.Exec("DELETE FROM member WHERE phone IN ('9000000001','9000000002')")

	// 构造真实 service（notifier 用真仓储 + hub；msg 表未迁移，通知走 best-effort 失败静默）
	msgRepo := repoMsg.NewMessageRepository(db)
	notifier := serviceMsg.NewNotifier(msgRepo, db, stream.NewSSEHub())
	repo := repoPersonal.NewPersonalRepository(db)
	svc := servicePersonal.NewPersonalService(db, repo, nil, nil, nil, sensitive.New(nil), notifier)

	// 100 个 goroutine 并发翻转 A→B
	const n = 100
	var wg sync.WaitGroup
	errCh := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := svc.ToggleFollow(a.ID, b.ID); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("ToggleFollow failed: %v", err) // 不变量1：全程无 500
	}

	// 不变量2：A→B 有且仅有一行
	var rowCount int64
	if err := db.Model(&entity.Follow{}).
		Where("follower_id = ? AND followee_id = ?", a.ID, b.ID).
		Count(&rowCount).Error; err != nil {
		t.Fatalf("count follow rows: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("follow rows = %d, want 1", rowCount)
	}

	// 不变量3：计数与实际 status=1 行数一致
	var statusOne int64
	if err := db.Model(&entity.Follow{}).
		Where("follower_id = ? AND followee_id = ? AND status = 1", a.ID, b.ID).
		Count(&statusOne).Error; err != nil {
		t.Fatalf("count status=1: %v", err)
	}
	var ma, mb entity.Member
	db.First(&ma, a.ID)
	db.First(&mb, b.ID)
	if ma.FollowCount != statusOne {
		t.Fatalf("A.follow_count = %d, want %d (status=1 rows)", ma.FollowCount, statusOne)
	}
	if mb.FansCount != statusOne {
		t.Fatalf("B.fans_count = %d, want %d (status=1 rows)", mb.FansCount, statusOne)
	}
}
