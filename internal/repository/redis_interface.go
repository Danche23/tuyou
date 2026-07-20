package repository

import (
	"context"
	"time"
)

// RedisRepository Redis 仓储接口
type RedisRepository interface {
	// Set 设置缓存
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	// Get 获取缓存
	Get(ctx context.Context, key string, dest interface{}) error
	// Del 删除缓存
	Del(ctx context.Context, key string) error
	// Exists 检查是否存在
	Exists(ctx context.Context, key string) (bool, error)
}
