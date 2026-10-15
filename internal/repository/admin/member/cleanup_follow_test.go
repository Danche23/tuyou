package member

// 方案C 关注翻转下 admin 硬删会员的计数回归测试（需真连 MySQL，DB 门控）。
// 用法：TEST_DB="root:密码@tcp(127.0.0.1:3306)/tuyou_test?charset=utf8mb4&parseTime=True&loc=Local" \
//        go test -run TestCleanupFollowForMemberStatusFilter ./internal/repository/admin/member/
// 未设置 TEST_DB 时自动 Skip（CI 安全）。
// 复现 Critical #1：cleanupFollowForMember 的 UPDATE 子查询若不过滤 status=1，
// 会把取关保留的 status=0 行也计入，导致已取关用户被二次扣减（计数漂移）。
// 断言：
//   1. C.fans_count == 0（只有 status=1 的 C 被扣）
//   2. B.fans_count == 1（B 在取关时已扣过，不得再扣）
//   3. follow 表 A 相关行已删光

import (
	"net"
	"os"
	"testing"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"tuyou/internal/model/entity"
)

func TestCleanupFollowForMemberStatusFilter(t *testing.T) {
	dsn := os.Getenv("TEST_DB")
	if dsn == "" {
		t.Skip("TEST_DB not set; skipping DB regression test")
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

	// 清场 + 建三个测试账号（member 的 phone/password 非空，raw Exec 硬删绕开软删作用域）
	db.Exec("DELETE FROM follow")
	db.Exec("DELETE FROM member WHERE phone IN ('9000000011','9000000012','9000000013')")
	a := entity.Member{Nickname: "cleanup-test-a", Phone: "9000000011", Password: "x"}
	b := entity.Member{Nickname: "cleanup-test-b", Phone: "9000000012", Password: "x"}
	c := entity.Member{Nickname: "cleanup-test-c", Phone: "9000000013", Password: "x"}
	for _, m := range []*entity.Member{&a, &b, &c} {
		if err := db.Create(m).Error; err != nil {
			t.Fatalf("create member: %v", err)
		}
	}
	defer db.Exec("DELETE FROM member WHERE phone IN ('9000000011','9000000012','9000000013')")

	// A 关注 B 和 C（status=1 两行）
	if err := db.Create(&entity.Follow{FollowerID: a.ID, FolloweeID: b.ID, Status: 1}).Error; err != nil {
		t.Fatalf("create follow A->B: %v", err)
	}
	if err := db.Create(&entity.Follow{FollowerID: a.ID, FolloweeID: c.ID, Status: 1}).Error; err != nil {
		t.Fatalf("create follow A->C: %v", err)
	}
	// 计数与 status=1 行一致：A.follow_count=2、B.fans_count=1、C.fans_count=1
	db.Model(&entity.Member{}).Where("id = ?", a.ID).Update("follow_count", 2)
	db.Model(&entity.Member{}).Where("id = ?", b.ID).Update("fans_count", 1)
	db.Model(&entity.Member{}).Where("id = ?", c.ID).Update("fans_count", 1)

	// A 取关 B：(A,B) 行 status 置 0，A.follow_count 同步 -1
	db.Model(&entity.Follow{}).Where("follower_id = ? AND followee_id = ?", a.ID, b.ID).Update("status", 0)
	db.Model(&entity.Member{}).Where("id = ?", a.ID).Update("follow_count", 1)

	// 事务内清理 A 的关注数据
	if err := db.Transaction(func(tx *gorm.DB) error {
		return cleanupFollowForMember(tx, a.ID)
	}); err != nil {
		t.Fatalf("cleanupFollowForMember: %v", err)
	}

	// 断言 1+2：只有 status=1 的 C 被扣，已取关的 B 不得再扣
	var mb, mc entity.Member
	db.First(&mb, b.ID)
	db.First(&mc, c.ID)
	if mc.FansCount != 0 {
		t.Fatalf("C.fans_count = %d, want 0 (status=1 followee 应扣)", mc.FansCount)
	}
	if mb.FansCount != 1 {
		t.Fatalf("B.fans_count = %d, want 1 (取关已扣过，不得二次扣减)", mb.FansCount)
	}

	// 断言 3：follow 表 A 相关行已删光
	var followCount int64
	if err := db.Raw("SELECT COUNT(*) FROM follow WHERE follower_id = ? OR followee_id = ?", a.ID, a.ID).Scan(&followCount).Error; err != nil {
		t.Fatalf("count follow rows: %v", err)
	}
	if followCount != 0 {
		t.Fatalf("follow rows related to A = %d, want 0", followCount)
	}
}
