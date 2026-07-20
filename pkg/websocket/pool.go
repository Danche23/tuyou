package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// WriteWait 写入超时时间
	WriteWait = 10 * time.Second
	// PongWait 等待 Pong 的时间
	PongWait = 60 * time.Second
	// PingPeriod 发送 Ping 的周期 (必须小于 PongWait)
	PingPeriod = (PongWait * 9) / 10
	// MaxMessageSize 最大消息大小
	MaxMessageSize = 512
)

// SessionMetadata 会话元数据
type SessionMetadata struct {
	UserID      uint
	SessionType string
	CreatedAt   int64
}

// Client 包装了 WebSocket 连接和发送缓冲区
type Client struct {
	Pool     *ConnectionPool
	Conn     *websocket.Conn
	Send     chan []byte
	Metadata *SessionMetadata
	once     sync.Once
}

// ConnectionPool WebSocket连接池（优化版：读写分离、心跳检测）
type ConnectionPool struct {
	// userID -> { client -> struct{} }
	userClients map[uint]map[*Client]struct{}
	mu          sync.RWMutex
}

// NewConnectionPool 创建连接池
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		userClients: make(map[uint]map[*Client]struct{}),
	}
}

// Add 创建并添加一个新客户端
func (p *ConnectionPool) Add(userID uint, conn *websocket.Conn, metadata *SessionMetadata) *Client {
	client := &Client{
		Pool:     p,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Metadata: metadata,
	}

	p.mu.Lock()
	if p.userClients[userID] == nil {
		p.userClients[userID] = make(map[*Client]struct{})
	}
	p.userClients[userID][client] = struct{}{}
	p.mu.Unlock()

	// 启动写协程
	go client.WritePump()

	return client
}

// Close 关闭客户端连接
func (p *Client) Close() {
	p.once.Do(func() {
		p.Pool.mu.Lock()
		defer p.Pool.mu.Unlock()

		userID := p.Metadata.UserID
		if clients, ok := p.Pool.userClients[userID]; ok {
			delete(clients, p)
			if len(clients) == 0 {
				delete(p.Pool.userClients, userID)
			}
		}
		close(p.Send)
		_ = p.Conn.Close()
	})
}

// WritePump 负责将消息从通道写入网络（每个连接一个协程）
func (c *Client) WritePump() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				// 通道已关闭
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 将队列中剩余的消息合并发送
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendToUser 发送消息给指定用户的所有客户端
func (p *ConnectionPool) SendToUser(userID uint, data []byte) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if clients, ok := p.userClients[userID]; ok {
		for client := range clients {
			select {
			case client.Send <- data:
			default:
				// 如果发送缓冲区满了，主动关闭这个慢连接，防止阻塞整个系统
				go client.Close()
			}
		}
	}
}

// Broadcast 广播消息给所有用户
func (p *ConnectionPool) Broadcast(data []byte) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, clients := range p.userClients {
		for client := range clients {
			select {
			case client.Send <- data:
			default:
				go client.Close()
			}
		}
	}
}

// GetConnectionCount 获取总连接数
func (p *ConnectionPool) GetConnectionCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	count := 0
	for _, clients := range p.userClients {
		count += len(clients)
	}
	return count
}

// CloseAll 关闭所有连接
func (p *ConnectionPool) CloseAll() {
	p.mu.RLock()
	// 收集所有 client 避免在循环中持有写锁导致 Close() 死锁
	var allClients []*Client
	for _, clients := range p.userClients {
		for client := range clients {
			allClients = append(allClients, client)
		}
	}
	p.mu.RUnlock()

	for _, client := range allClients {
		client.Close()
	}
}
