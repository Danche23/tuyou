package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisRepository struct {
	redis *redis.Client
}

// NewRedisRepository 创建 Redis 仓库
func NewRedisRepository(redisClient *redis.Client) RedisRepository {
	return &redisRepository{
		redis: redisClient,
	}
}

func (r *redisRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if r.redis == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.redis.Set(ctx, key, data, expiration).Err()
}

func (r *redisRepository) Get(ctx context.Context, key string, dest interface{}) error {
	if r.redis == nil {
		return redis.Nil
	}
	data, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (r *redisRepository) Del(ctx context.Context, key string) error {
	if r.redis == nil {
		return nil
	}
	return r.redis.Del(ctx, key).Err()
}

func (r *redisRepository) Exists(ctx context.Context, key string) (bool, error) {
	if r.redis == nil {
		return false, nil
	}
	n, err := r.redis.Exists(ctx, key).Result()
	return n > 0, err
}
