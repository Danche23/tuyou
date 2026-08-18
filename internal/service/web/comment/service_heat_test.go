package comment

import (
	"context"
	"testing"
	"time"

	"tuyou/internal/repository"
)

// mockHeatRedis 只实现配额相关方法，其余方法走嵌入接口（被调用即 panic，证明测试没走其他 Redis 路径）
type mockHeatRedis struct {
	repository.RedisRepository
	quota   int64
	expires int
}

func (m *mockHeatRedis) IncrBy(_ context.Context, _ string, increment int64) (int64, error) {
	m.quota += increment
	return m.quota, nil
}

func (m *mockHeatRedis) Expire(_ context.Context, _ string, _ time.Duration) error {
	m.expires++
	return nil
}

func TestConsumeHeatQuota_UnderLimit(t *testing.T) {
	r := &mockHeatRedis{}
	svc := &commentService{redis: r}
	if !svc.consumeHeatQuota(1, heatLikeScore) {
		t.Fatal("配额内应返回 true")
	}
}

func TestConsumeHeatQuota_OverLimitRollback(t *testing.T) {
	r := &mockHeatRedis{quota: heatQuotaDaily} // 已用满
	svc := &commentService{redis: r}
	if svc.consumeHeatQuota(1, heatLikeScore) {
		t.Fatal("超限应返回 false")
	}
	if r.quota != heatQuotaDaily {
		t.Fatalf("超限应回滚扣减，期望配额不变=%d，实际 %d", heatQuotaDaily, r.quota)
	}
}

func TestConsumeHeatQuota_RedisDownFailOpen(t *testing.T) {
	svc := &commentService{redis: nil}
	if !svc.consumeHeatQuota(1, heatLikeScore) {
		t.Fatal("Redis 不可用应 fail-open 返回 true")
	}
}

func TestConsumeHeatQuota_FirstCreateSetsExpiry(t *testing.T) {
	r := &mockHeatRedis{}
	svc := &commentService{redis: r}
	svc.consumeHeatQuota(1, heatReplyScore)
	if r.expires != 1 {
		t.Fatalf("首次创建 key 应设置过期，实际 %d 次", r.expires)
	}
}

func TestShouldScoreReply(t *testing.T) {
	if shouldScoreReply(heatReplyPerComment, true) {
		t.Fatal("已计分回复达3条时不应再计分")
	}
	if !shouldScoreReply(2, true) {
		t.Fatal("未达上限且配额充足时应计分")
	}
	if shouldScoreReply(2, false) {
		t.Fatal("配额不足时不应计分")
	}
}
