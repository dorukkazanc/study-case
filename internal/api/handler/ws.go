package handler

import (
	"net/http"
	"study-case/internal/hub"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct {
	hub *hub.Hub
}

func NewWSHandler(h *hub.Hub) *WSHandler {
	return &WSHandler{hub: h}
}

func (h *WSHandler) NotificationStatus(c *gin.Context) {
	id := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	ch := h.hub.Subscribe(id)
	defer h.hub.Unsubscribe(id, ch)

	for update := range ch {
		if err := conn.WriteJSON(update); err != nil {
			return
		}
	}
}
