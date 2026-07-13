// Package ws 【第9步】WebSocket HTTP 入口：鉴权、升级连接、心跳保活。
package ws

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// TokenParser 从 JWT 字符串解析出 userId / driverId（token 里字段名均为 userId）
type TokenParser func(token string) (int64, error)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 开发环境；生产应校验 Origin
	},
}

// ServeUserWS 乘客订单 WebSocket：GET /ws/user?token=xxx
func ServeUserWS(hub *Hub, parse TokenParser) http.HandlerFunc {
	return serveWS(hub, parse, true)
}

// ServeDriverWS 司机订单 WebSocket：GET /ws/driver?token=xxx
func ServeDriverWS(hub *Hub, parse TokenParser) http.HandlerFunc {
	return serveWS(hub, parse, false)
}

func serveWS(hub *Hub, parse TokenParser, isUser bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			token = r.Header.Get("token")
		}
		if token == "" {
			http.Error(w, "token 为空", http.StatusUnauthorized)
			return
		}

		uid, err := parse(token)
		if err != nil || uid <= 0 {
			http.Error(w, "token 无效", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("[ws] upgrade failed: %v", err)
			return
		}

		if isUser {
			hub.RegisterUser(uid, conn)
			defer func() {
				hub.UnregisterUser(uid, conn)
				_ = conn.Close()
			}()
		} else {
			hub.RegisterDriver(uid, conn)
			defer func() {
				hub.UnregisterDriver(uid, conn)
				_ = conn.Close()
			}()
		}

		conn.SetReadLimit(4096)
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})
		go writePing(conn)

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}
}

func writePing(conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
	}
}

// ParseUserIDFromClaims 解析 JWT claims 中的 userId
func ParseUserIDFromClaims(v interface{}) (int64, error) {
	switch x := v.(type) {
	case string:
		return strconv.ParseInt(x, 10, 64)
	case float64:
		return int64(x), nil
	case int64:
		return x, nil
	case int:
		return int64(x), nil
	default:
		return 0, strconv.ErrSyntax
	}
}
