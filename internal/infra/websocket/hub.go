package websocket

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"go.uber.org/zap"
)

type Hub struct {
	// Conexões registradas indexadas por userID
	clients    map[uuid.UUID]map[*Client]bool
	clientsMu  sync.RWMutex
	register   chan *Client
	unregister chan *Client
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewHub() *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Register enfileira o client no event loop. Retorna false se o hub já foi parado.
func (h *Hub) Register(c *Client) bool {
	select {
	case <-h.ctx.Done():
		return false
	case h.register <- c:
		return true
	}
}

func (h *Hub) Unregister(c *Client) {
	select {
	case <-h.ctx.Done():
		c.closeSend()
	case h.unregister <- c:
	}
}

func (h *Hub) Start() {
	go func() {
		for {
			select {
			case <-h.ctx.Done():
				return
			case client := <-h.register:
				h.clientsMu.Lock()
				if _, ok := h.clients[client.userID]; !ok {
					h.clients[client.userID] = make(map[*Client]bool)
				}
				h.clients[client.userID][client] = true
				h.clientsMu.Unlock()
				logger.Log.Debug("WS Client registered",
					zap.String("user_id", client.userID.String()),
					zap.String("remote_addr", client.conn.RemoteAddr().String()))

			case client := <-h.unregister:
				h.clientsMu.Lock()
				if userClients, ok := h.clients[client.userID]; ok {
					if _, exists := userClients[client]; exists {
						delete(userClients, client)
						client.closeSend()
						logger.Log.Debug("WS Client unregistered",
							zap.String("user_id", client.userID.String()),
							zap.String("remote_addr", client.conn.RemoteAddr().String()))
					}
					if len(userClients) == 0 {
						delete(h.clients, client.userID)
					}
				}
				h.clientsMu.Unlock()
			}
		}
	}()
}

func (h *Hub) Stop() {
	h.cancel()
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	for userID, userClients := range h.clients {
		for client := range userClients {
			client.closeSend()
			client.conn.Close()
			delete(userClients, client)
		}
		delete(h.clients, userID)
	}
}

// BroadcastToUser envia dados JSON para todas as conexões ativas de um userID
func (h *Hub) BroadcastToUser(userID uuid.UUID, message interface{}) {
	payload, err := json.Marshal(message)
	if err != nil {
		logger.Log.Error("failed to marshal WS message", zap.Error(err))
		return
	}

	h.clientsMu.RLock()
	userClients, ok := h.clients[userID]
	if !ok || len(userClients) == 0 {
		h.clientsMu.RUnlock()
		return
	}

	// Cria uma cópia dos clients para evitar segurar o lock de leitura durante o envio lento
	clientsCopy := make([]*Client, 0, len(userClients))
	for client := range userClients {
		clientsCopy = append(clientsCopy, client)
	}
	h.clientsMu.RUnlock()

	for _, client := range clientsCopy {
		if client.trySend(payload) {
			continue
		}
		// Fechado ou buffer cheio -> desconecta sem panic em canal fechado
		logger.Log.Warn("WS Client send dropped, unregistering client",
			zap.String("user_id", userID.String()),
			zap.String("remote_addr", client.conn.RemoteAddr().String()))
		h.Unregister(client)
		client.conn.Close()
	}
}
