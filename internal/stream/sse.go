package stream

import (
	"encoding/json"
	"sync"
)

// SSEClient 单个 SSE 连接，含独立缓存队列
type SSEClient struct {
	Events chan []byte //缓存事件队列，满了就丢
}

// SSEHub 会员id ->SSE连接集合
type SSEHub struct {
	mu    sync.Mutex
	conns map[uint]map[*SSEClient]struct{}
}

// NewSSEHub 创建 hub
func NewSSEHub() *SSEHub {
	return &SSEHub{
		conns: make(map[uint]map[*SSEClient]struct{}),
	}
}

// Subscribe 注册一个会员连接，返回客户端对象。一个对象可多端（多个连接）
func (h *SSEHub) Subscribe(memberID uint) *SSEClient {
	c := &SSEClient{Events: make(chan []byte, 32)}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[memberID] == nil {
		h.conns[memberID] = make(map[*SSEClient]struct{})
	}
	h.conns[memberID][c] = struct{}{}
	return c
}

func (h *SSEHub) Unsubscribe(memberID uint, c *SSEClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.conns[memberID]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.conns, memberID)
		}
	}
}

func formatSSE(event string, data any) []byte {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	return []byte("event:" + event + "\ndata:" + string(payload) + "\n\n")
}

// Publish 推送给指定会员
// 非阻塞：对方离线/缓存满直接丢弃，不影响调用方
func (h *SSEHub) Publish(memberID uint, event string, data any) {
	msg := formatSSE(event, data)
	if msg == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.conns[memberID] {
		select {
		case c.Events <- msg:
		default:
		}
	}
}

// PublishAll 推送给所有在线会员
func (h *SSEHub) PublishAll(event string, data any) {
	msg := formatSSE(event, data)
	if msg == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, set := range h.conns {
		for c := range set {
			select {
			case c.Events <- msg:
			default:
			}
		}
	}
}
