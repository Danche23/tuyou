package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRepository Redis 仓储接口
type RedisRepository interface {
	// Set 设置缓存
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	// Get 获取缓存
	Get(ctx context.Context, key string, dest interface{}) error

	// XAdd 向 Stream 发送一条消息。Redis 不可用时返回 redis.Nil
	XAdd(ctx context.Context, stream string, values map[string]interface{}) error
	// XGroupCreate 创建消费者组（带 MKSTREAM，自动创建 Stream）。Redis 不可用时返回 redis.Nil
	XGroupCreate(ctx context.Context, stream, group, start string) error
	// XReadGroup 从消费者组读取消息（BLOCK 模式）
	XReadGroup(ctx context.Context, group, consumer string, streams []string, count int64, block time.Duration) ([]redis.XStream, error)
	// XAck 确认消息已处理
	XAck(ctx context.Context, stream, group string, ids ...string) error
	// XTrimMaxLen 截断 Stream 到指定最大长度（非精确修剪）
	XTrimMaxLen(ctx context.Context, stream string, maxLen int64) error
}
