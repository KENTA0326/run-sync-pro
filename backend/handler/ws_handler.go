package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/domain"
	"github.com/KENTA0326/run-sync-pro/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type wsHub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]domain.UserID
}

var trainingLogWS = &wsHub{clients: make(map[*websocket.Conn]domain.UserID)}

// TrainingLogsWebSocket GET /api/v1/ws/training-logs （レガシー: GET /auth/ws/training-logs）
// 接続後 JSON メッセージ { "type": "ping" } → { "type": "pong" }
func (h *Handlers) TrainingLogsWebSocket(c *gin.Context) {
	userID, ok := domain.UserIDFromRequest(c.Request)
	if !ok {
		response.WriteError(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "認証が必要です")
		return
	}

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	trainingLogWS.mu.Lock()
	trainingLogWS.clients[conn] = userID
	trainingLogWS.mu.Unlock()
	defer func() {
		trainingLogWS.mu.Lock()
		delete(trainingLogWS.clients, conn)
		trainingLogWS.mu.Unlock()
	}()

	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	for {
		var msg map[string]string
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		switch msg["type"] {
		case "ping":
			_ = conn.WriteJSON(map[string]string{"type": "pong"})
		default:
			_ = conn.WriteJSON(map[string]string{"type": "error", "message": "unknown message type"})
		}
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	}
}

// BroadcastTrainingLogEvent は取込完了等を WebSocket クライアントへ通知する（サンプル）。
func BroadcastTrainingLogEvent(userID domain.UserID, eventType string, payload map[string]any) {
	trainingLogWS.mu.Lock()
	defer trainingLogWS.mu.Unlock()
	msg := map[string]any{"type": eventType}
	for k, v := range payload {
		msg[k] = v
	}
	for conn, uid := range trainingLogWS.clients {
		if uid != userID {
			continue
		}
		_ = conn.WriteJSON(msg)
	}
}
