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

// XAdd 向 Stream 发送一条消息。Redis 不可用时返回 redis.Nil 以触发调用方降级
func (r *redisRepository) XAdd(ctx context.Context, stream string, values map[string]interface{}) error {
	if r.redis == nil {
		return redis.Nil
	}
	return r.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Err()
}

// XGroupCreate 创建消费者组（MKSTREAM，Stream 不存在时自动创建）。
// Redis 不可用时返回 redis.Nil
func (r *redisRepository) XGroupCreate(ctx context.Context, stream, group, start string) error {
	if r.redis == nil {
		return redis.Nil
	}
	return r.redis.XGroupCreateMkStream(ctx, stream, group, start).Err()
}

// XReadGroup 从消费者组读取消息（BLOCK 模式）
func (r *redisRepository) XReadGroup(ctx context.Context, group, consumer string, streams []string, count int64, block time.Duration) ([]redis.XStream, error) {
	if r.redis == nil {
		return nil, redis.Nil
	}
	return r.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  append(streams, ">"),
		Count:    count,
		Block:    block,
	}).Result()
}

// XAck 确认消息已处理
func (r *redisRepository) XAck(ctx context.Context, stream, group string, ids ...string) error {
	if r.redis == nil {
		return nil
	}
	return r.redis.XAck(ctx, stream, group, ids...).Err()
}

// XTrimMaxLen 截断 Stream 到指定最大长度（非精确修剪）
func (r *redisRepository) XTrimMaxLen(ctx context.Context, stream string, maxLen int64) error {
	if r.redis == nil {
		return nil
	}
	return r.redis.XTrimMaxLen(ctx, stream, maxLen).Err()
}
