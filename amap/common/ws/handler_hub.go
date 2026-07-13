// Package ws 【第9步·WebSocket】apiGateway 上的长连接，把 MQ 通知实时推到 App。
//
// 在本项目中的作用：
//   - Hub：维护「用户 id / 司机 id → WebSocket 连接」映射
//   - handler：/ws/user、/ws/driver 升级 HTTP 为 WS，JWT 鉴权
//   - consumer：消费 RabbitMQ 通知队列，按 event 推给对应在线用户
//
// 典型 event：driver_accepted（乘客）、order_cancelled（乘客）、order_completed（乘客）、new_order_nearby（司机）
package ws

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub 连接管理中心（单 Gateway 实例 MVP）
type Hub struct {
	mu          sync.RWMutex
	userConns   map[int64]*clientConn
	driverConns map[int64]*clientConn
}

type clientConn struct {
	id   int64
	conn *websocket.Conn
	mu   sync.Mutex
}

// NewHub 创建 Hub
func NewHub() *Hub {
	return &Hub{
		userConns:   make(map[int64]*clientConn),
		driverConns: make(map[int64]*clientConn),
	}
}

// RegisterUser 注册乘客连接
func (h *Hub) RegisterUser(userId int64, conn *websocket.Conn) {
	h.register(&h.userConns, userId, conn, "user")
}

// UnregisterUser 移除乘客连接
func (h *Hub) UnregisterUser(userId int64, conn *websocket.Conn) {
	h.unregister(&h.userConns, userId, conn, "user")
}

// RegisterDriver 注册司机连接
func (h *Hub) RegisterDriver(driverId int64, conn *websocket.Conn) {
	h.register(&h.driverConns, driverId, conn, "driver")
}

// UnregisterDriver 移除司机连接
func (h *Hub) UnregisterDriver(driverId int64, conn *websocket.Conn) {
	h.unregister(&h.driverConns, driverId, conn, "driver")
}

func (h *Hub) register(m *map[int64]*clientConn, id int64, conn *websocket.Conn, role string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := (*m)[id]; ok {
		_ = old.conn.Close()
	}
	(*m)[id] = &clientConn{id: id, conn: conn}
	log.Printf("[ws-hub] %s %d connected, total=%d", role, id, len(*m))
}

func (h *Hub) unregister(m *map[int64]*clientConn, id int64, conn *websocket.Conn, role string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	cur, ok := (*m)[id]
	if !ok || cur.conn != conn {
		return
	}
	delete(*m, id)
	log.Printf("[ws-hub] %s %d disconnected, total=%d", role, id, len(*m))
}

// PushUser 向乘客推送 JSON
func (h *Hub) PushUser(userId int64, payload []byte) bool {
	return h.push(&h.userConns, userId, payload, "user")
}

// PushDriver 向司机推送 JSON
func (h *Hub) PushDriver(driverId int64, payload []byte) bool {
	return h.push(&h.driverConns, driverId, payload, "driver")
}

func (h *Hub) push(m *map[int64]*clientConn, id int64, payload []byte, role string) bool {
	h.mu.RLock()
	c, ok := (*m)[id]
	h.mu.RUnlock()
	if !ok {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
		log.Printf("[ws-hub] push %s %d failed: %v", role, id, err)
		return false
	}
	return true
}
