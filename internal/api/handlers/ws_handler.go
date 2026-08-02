package handlers

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	wsInfra "github.com/misalima/edunex-backend/internal/infra/websocket"
	"go.uber.org/zap"
)

type WsHandler struct {
	hub        *wsInfra.Hub
	tickets    *wsInfra.TicketStore
	upgrader   websocket.Upgrader
}

func NewWsHandler(hub *wsInfra.Hub, tickets *wsInfra.TicketStore) *WsHandler {
	return &WsHandler{
		hub:     hub,
		tickets: tickets,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Permitir todas as origens para facilidade no dev/local.
				// Next.js (client) fará a requisição a partir de portas locais distintas.
				return true
			},
		},
	}
}

// IssueTicket godoc
// @Summary Issue a short-lived WebSocket ticket
// @Description Exchanges a Bearer JWT for a single-use WS ticket (TTL ~30s). Use the ticket as ?ticket= on the WebSocket URL — never put the JWT in the query string.
// @Tags WebSocket
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /ws/ticket [post]
func (h *WsHandler) IssueTicket(c echo.Context) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	ticket, err := h.tickets.Issue(userID)
	if err != nil {
		logger.Log.Error("failed to issue WS ticket", zap.Error(err), zap.String("user_id", userID.String()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to issue ticket"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"ticket": ticket,
	})
}

// ConnectWs Upgrade HTTP -> WebSocket after consuming a one-time ticket.
func (h *WsHandler) ConnectWs(c echo.Context) error {
	ticket := c.QueryParam("ticket")
	userID, ok := h.tickets.Consume(ticket)
	if !ok {
		logger.Log.Warn("WS connection attempt blocked: invalid or expired ticket")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired ticket"})
	}

	wsConn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logger.Log.Error("failed to upgrade to websocket connection", zap.Error(err))
		return err
	}

	client := wsInfra.NewClient(h.hub, wsConn, userID)
	if !h.hub.Register(client) {
		logger.Log.Warn("WS hub is stopped, rejecting connection", zap.String("user_id", userID.String()))
		wsConn.Close()
		return nil
	}

	go client.WritePump()
	go client.ReadPump()

	return nil
}
