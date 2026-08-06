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

// Redis Stream 配置
const (
	StreamKey     = "stream:content" // Stream 名称
	ConsumerGroup = "cg:content"     // 消费者组名
	ConsumerName  = "consumer:1"     // 消费者实例名（同一组内唯一）

	BatchMax      = 50              // 单批最大消息数，达到立即 flush
	FlushInterval = 5 * time.Second // 无新消息时定时 flush 间隔
	StreamMaxLen  = 10000           // Stream 最大长度，超出自动截断
)

// validTables 消息 type 字段对应的数据库表名白名单
var validTables = map[string]bool{
	"note":     true,
	"strategy": true,
}

// Consumer 异步消费者，从 Redis Stream 拉取消息 → 分发处理 → 批量写入 MySQL。
// 非线程安全，每个实例单 goroutine 消费。
type Consumer struct {
	redisRepo repository.RedisRepository
	db        *gorm.DB

	ctx    context.Context    // 控制消费者生命周期
	cancel context.CancelFunc // 触发优雅关闭
	done   chan struct{}      // 关闭完成信号
}

// NewConsumer 创建消费者实例
func NewConsumer(redisRepo repository.RedisRepository, db *gorm.DB) *Consumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &Consumer{
		redisRepo: redisRepo,
		db:        db,
		ctx:       ctx,
		cancel:    cancel,
		done:      make(chan struct{}),
	}
}

// Start 创建消费者组并启动消费循环。Redis 不可用时消费者不启动。
func (c *Consumer) Start() {
	if c.redisRepo == nil {
		logger.Warn("Consumer: Redis 未初始化，不启动")
		close(c.done)
		return
	}
	// 创建消费者组（已存在则忽略 BUSYGROUP 错误）
	_ = c.redisRepo.XGroupCreate(c.ctx, StreamKey, ConsumerGroup, "0")
	go c.consume()
	logger.Info("Consumer 已启动",
		zap.String("stream", StreamKey),
		zap.String("group", ConsumerGroup),
	)
}

// Stop 优雅停止消费者，等待当前批次处理完成。
func (c *Consumer) Stop() {
	c.cancel()
	<-c.done
	logger.Info("Consumer 已停止")
}

// consume 是消费者的主循环：拉取消息 → 聚合 → 条件触发批量写 DB。
// panic 时自动重启，保证消费者始终运行。
func (c *Consumer) consume() {
	defer close(c.done)
	defer c.restartOnPanic()

	// 聚合计数器：counts["note"][noteID] = 累计浏览量
	counts := make(map[string]map[uint]int64)
	var msgIDs []string // 本批次所有消息 ID，XAck 用
	lastFlush := time.Now()

	for {
		// 收到关闭信号时，刷完最后一批再退出
		select {
		case <-c.ctx.Done():
			c.flush(counts, msgIDs)
			return
		default:
		}

		// 从 Stream 拉消息并合并到聚合计数器
		c.pullAndMerge(counts, &msgIDs)

		// 批量条件：消息数达标 或 定时到期且有积压
		n := len(msgIDs)
		if n >= BatchMax || (n > 0 && time.Since(lastFlush) >= FlushInterval) {
			c.flush(counts, msgIDs)
			// 重置批次
			counts = make(map[string]map[uint]int64)
			msgIDs = nil
			lastFlush = time.Now()
		}
	}
}

// pullAndMerge 从 Stream 拉取一批消息，解析后合并到聚合计数器中。
// BLOCK 模式下超时返回 redis.Nil，此时直接返回等待下次循环。
func (c *Consumer) pullAndMerge(counts map[string]map[uint]int64, msgIDs *[]string) {
	streams, err := c.redisRepo.XReadGroup(c.ctx,
		ConsumerGroup, ConsumerName,
		[]string{StreamKey}, int64(BatchMax), FlushInterval,
	)
	if err != nil {
		// BLOCK 超时 → 没有新消息，正常跳过
		// 其他错误 → 打日志后等一秒重试
		if err != redis.Nil && c.ctx.Err() == nil {
			logger.Warn("Consumer: XReadGroup 失败", zap.Error(err))
			time.Sleep(time.Second)
		}
		return
	}

	for _, s := range streams {
		for _, m := range s.Messages {
			table, id, ok := parseMsg(m)
			if !ok {
				// 消息格式异常，XAck 跳过避免阻塞队列
				_ = c.redisRepo.XAck(c.ctx, StreamKey, ConsumerGroup, m.ID)
				continue
			}
			// 聚合：同一 table 下相同 id 的浏览量累加
			if counts[table] == nil {
				counts[table] = make(map[uint]int64)
			}
			counts[table][id]++
			*msgIDs = append(*msgIDs, m.ID)
		}
	}
}

// flush 将聚合后的计数批量写入 MySQL，成功则 XAck 确认消息消费完成。
// 任何一条 UPDATE 失败都放弃本批次，消息留在 Stream 中等待下次重试。
func (c *Consumer) flush(counts map[string]map[uint]int64, msgIDs []string) {
	if len(msgIDs) == 0 {
		return
	}

	// 批量 UPDATE view_count
	for table, ids := range counts {
		for id, n := range ids {
			sql := fmt.Sprintf("UPDATE %s SET view_count = view_count + ? WHERE id = ?", table)
			if err := c.db.Exec(sql, n, id).Error; err != nil {
				logger.Error("Consumer: UPDATE 失败，本批次不 XAck",
					zap.String("table", table),
					zap.Uint("id", id),
					zap.Int64("count", n),
					zap.Error(err),
				)
				return // 不 XAck，等重试
			}
		}
	}

	// 全部写入成功 → XAck 确认消费
	_ = c.redisRepo.XAck(c.ctx, StreamKey, ConsumerGroup, msgIDs...)
	// 截断防止 Stream 无限增长
	_ = c.redisRepo.XTrimMaxLen(c.ctx, StreamKey, StreamMaxLen)
}

// restartOnPanic 捕获 panic 并重启消费循环。
func (c *Consumer) restartOnPanic() {
	if r := recover(); r != nil {
		logger.Error("Consumer panic，重启", zap.Any("panic", r))
		c.ctx, c.cancel = context.WithCancel(context.Background())
		c.done = make(chan struct{})
		go c.consume()
	}
}

// parseMsg 解析 Redis Stream 消息，校验 type 和 id 字段合法性。
// 返回表名、记录 ID 和是否解析成功。
func parseMsg(m redis.XMessage) (table string, id uint, ok bool) {
	// 校验 type 字段（必须在白名单中）
	t, _ := m.Values["type"].(string)
	if !validTables[t] {
		return "", 0, false
	}

	// 校验 id 字段（必须是合法 uint）
	idStr, _ := m.Values["id"].(string)
	n, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return "", 0, false
	}

	return t, uint(n), true
}
