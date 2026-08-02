package websocket

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"go.uber.org/zap"
)

const (
	// Tempo permitido para escrever uma mensagem no peer
	writeWait = 10 * time.Second

	// Tempo permitido para ler o próximo pong do peer
	pongWait = 60 * time.Second

	// Período de envio de pings ao peer (deve ser menor que pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Tamanho máximo da mensagem recebida do peer (não esperamos comandos, apenas batimentos)
	maxMessageSize = 512
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID uuid.UUID
	send   chan []byte

	mu     sync.Mutex
	closed bool
}

func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		userID: userID,
		send:   make(chan []byte, 256),
	}
}

// trySend enfileira uma mensagem sem panic se o canal já foi fechado.
// Retorna false se o client já foi fechado ou se o buffer está cheio.
func (c *Client) trySend(message []byte) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return false
	}
	select {
	case c.send <- message:
		return true
	default:
		return false
	}
}

// closeSend fecha o canal send no máximo uma vez.
func (c *Client) closeSend() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	close(c.send)
}

// ReadPump lê dados da conexão. Só serve para detectar desconexão e manter o batimento
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		logger.Log.Error("WS failed to set read deadline", zap.Error(err), zap.String("user_id", c.userID.String()))
		return
	}
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			return err
		}
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Log.Error("WS connection error in read pump", zap.Error(err))
			}
			break
		}
	}
}

// WritePump consome as mensagens pendentes no canal send e envia via WebSocket
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logger.Log.Error("WS failed to set write deadline", zap.Error(err), zap.String("user_id", c.userID.String()))
				return
			}
			if !ok {
				// O Hub fechou o canal send
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					logger.Log.Debug("WS failed to write close message", zap.Error(err), zap.String("user_id", c.userID.String()))
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logger.Log.Error("WS failed to get next writer", zap.Error(err), zap.String("user_id", c.userID.String()))
				return
			}
			if _, err := w.Write(message); err != nil {
				logger.Log.Error("WS failed to write message", zap.Error(err), zap.String("user_id", c.userID.String()))
				_ = w.Close()
				return
			}

			// Consome mensagens extras enfileiradas no mesmo tick
			n := len(c.send)
			for i := 0; i < n; i++ {
				if _, err := w.Write([]byte{'\n'}); err != nil {
					logger.Log.Error("WS failed to write message separator", zap.Error(err), zap.String("user_id", c.userID.String()))
					_ = w.Close()
					return
				}
				if _, err := w.Write(<-c.send); err != nil {
					logger.Log.Error("WS failed to write queued message", zap.Error(err), zap.String("user_id", c.userID.String()))
					_ = w.Close()
					return
				}
			}

			if err := w.Close(); err != nil {
				logger.Log.Error("WS failed to close writer", zap.Error(err), zap.String("user_id", c.userID.String()))
				return
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logger.Log.Error("WS failed to set write deadline for ping", zap.Error(err), zap.String("user_id", c.userID.String()))
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Log.Error("WS failed to write ping", zap.Error(err), zap.String("user_id", c.userID.String()))
				return
			}
		}
	}
}
