package stream

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"tuyou/internal/repository"
	"tuyou/pkg/logger"
)

// ======================== 常量 ========================

const (
	ViewStreamKey     = "stream:view_counts"
	ViewConsumerGroup = "cg:view_counts"
	ViewConsumerName  = "consumer:view_counts"
	ViewBatchMaxSize  = 50
	ViewBatchInterval = 5 * time.Second
	ViewStreamMaxLen  = 10000
)

// ======================== 消息结构 ========================

type viewMsg struct {
	msgType string // "note" | "strategy"
	id      uint
	msgID   string // Redis 消息 ID，XAck 用
}

type batchKey struct {
	msgType string
	id      uint
}

// ======================== ViewConsumer ========================

// ViewConsumer 浏览量异步消费者：从 Redis Stream 读取 → 聚合 → 批量写入 MySQL
type ViewConsumer struct {
	redisRepo repository.RedisRepository
	db        *gorm.DB
	ctx       context.Context
	cancel    context.CancelFunc
	done      chan struct{}
}

// NewViewConsumer 创建浏览量消费者
func NewViewConsumer(redisRepo repository.RedisRepository, db *gorm.DB) *ViewConsumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &ViewConsumer{
		redisRepo: redisRepo,
		db:        db,
		ctx:       ctx,
		cancel:    cancel,
		done:      make(chan struct{}),
	}
}

// Start 创建消费者组并启动消费循环
func (c *ViewConsumer) Start() {
	if c.redisRepo == nil {
		logger.Warn("ViewConsumer: Redis 未初始化，消费者不启动")
		close(c.done)
		return
	}

	// 创建消费者组（group 已存在则忽略 BUSYGROUP 错误）
	if err := c.redisRepo.XGroupCreate(c.ctx, ViewStreamKey, ViewConsumerGroup, "0"); err != nil {
		logger.Warn("ViewConsumer: 创建消费者组", zap.Error(err))
	}

	go c.loop()
	logger.Info("ViewConsumer: 消费者已启动",
		zap.String("stream", ViewStreamKey),
		zap.String("group", ViewConsumerGroup),
	)
}

// Stop 优雅停止消费者
func (c *ViewConsumer) Stop() {
	logger.Info("ViewConsumer: 正在停止...")
	c.cancel()
	<-c.done
	logger.Info("ViewConsumer: 已停止")
}

// loop 消费主循环，带 panic 恢复
func (c *ViewConsumer) loop() {
	defer close(c.done)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("ViewConsumer: panic，重启中", zap.Any("panic", r))
			c.ctx, c.cancel = context.WithCancel(context.Background())
			c.done = make(chan struct{})
			go c.loop()
		}
	}()

	c.run()
}

// run 消费主循环：读消息 → 聚合 → 批量 UPDATE → XAck → XTrim
func (c *ViewConsumer) run() {
	batch := make(map[batchKey]int64)
	var msgs []viewMsg
	lastFlush := time.Now()

	for {
		select {
		case <-c.ctx.Done():
			if len(batch) > 0 {
				c.flush(batch, msgs)
			}
			return
		default:
		}

		// XReadGroup BLOCK 读取
		streams, err := c.redisRepo.XReadGroup(c.ctx,
			ViewConsumerGroup, ViewConsumerName,
			[]string{ViewStreamKey}, int64(ViewBatchMaxSize), ViewBatchInterval,
		)

		if err != nil {
			if err == redis.Nil {
				// BLOCK 超时，无新消息
			} else if c.ctx.Err() != nil {
				continue
			} else {
				logger.Warn("ViewConsumer: XReadGroup 错误", zap.Error(err))
				time.Sleep(1 * time.Second)
			}
		}

		// 解析消息
		for _, stream := range streams {
			for _, msg := range stream.Messages {
				msgType, ok1 := msg.Values["type"].(string)
				idStr, ok2 := msg.Values["id"].(string)
				if !ok1 || !ok2 {
					logger.Error("ViewConsumer: 消息格式异常，已跳过并 XAck",
						zap.String("msgID", msg.ID),
						zap.Any("values", msg.Values),
					)
					_ = c.redisRepo.XAck(c.ctx, ViewStreamKey, ViewConsumerGroup, msg.ID)
					continue
				}
				id, err := strconv.ParseUint(idStr, 10, 64)
				if err != nil {
					logger.Error("ViewConsumer: 消息 id 解析失败，已跳过并 XAck",
						zap.String("msgID", msg.ID), zap.String("id", idStr), zap.Error(err),
					)
					_ = c.redisRepo.XAck(c.ctx, ViewStreamKey, ViewConsumerGroup, msg.ID)
					continue
				}

				key := batchKey{msgType: msgType, id: uint(id)}
				batch[key]++
				msgs = append(msgs, viewMsg{msgType: msgType, id: uint(id), msgID: msg.ID})
			}
		}

		// 批处理触发条件
		shouldFlush := len(msgs) >= ViewBatchMaxSize ||
			(len(msgs) > 0 && time.Since(lastFlush) >= ViewBatchInterval)

		if shouldFlush && len(batch) > 0 {
			c.flush(batch, msgs)
			batch = make(map[batchKey]int64)
			msgs = nil
			lastFlush = time.Now()
		}
	}
}

// flush 批量写 DB → XAck → XTrim
func (c *ViewConsumer) flush(batch map[batchKey]int64, msgs []viewMsg) {
	// 批量 UPDATE
	for key, count := range batch {
		tableName := key.msgType // "note" / "strategy"
		sql := fmt.Sprintf("UPDATE %s SET view_count = view_count + ? WHERE id = ?", tableName)
		if err := c.db.Exec(sql, count, key.id).Error; err != nil {
			logger.Error("ViewConsumer: 批量 UPDATE 失败，不 XAck 等待重试",
				zap.String("table", tableName),
				zap.Uint("id", key.id),
				zap.Int64("count", count),
				zap.Error(err),
			)
			return // 不 XAck，消息保留在 Stream 中，下次重新消费
		}
	}

	// 全部 UPDATE 成功后 XAck
	msgIDs := make([]string, len(msgs))
	for i, m := range msgs {
		msgIDs[i] = m.msgID
	}
	if err := c.redisRepo.XAck(c.ctx, ViewStreamKey, ViewConsumerGroup, msgIDs...); err != nil {
		logger.Warn("ViewConsumer: XAck 失败", zap.Error(err))
	}

	// XTrim 截断，防止 Stream 无限增长
	if err := c.redisRepo.XTrimMaxLen(c.ctx, ViewStreamKey, ViewStreamMaxLen); err != nil {
		logger.Warn("ViewConsumer: XTrim 失败", zap.Error(err))
	}
}
