package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"awebo/app/exceptions"
)

type realtimeHandler struct {
	*Handler
}

func newRealtimeHandler(h *Handler) *realtimeHandler {
	return &realtimeHandler{h}
}

func (h *realtimeHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/ws", h.wsAuth(), h.connect)
}

// upgrader's CheckOrigin mirrors corsMiddleware's wide-open stance
// (Access-Control-Allow-Origin: *) — the dev setup has the frontend and
// backend on different origins (:3000/:8080), and prod sits behind a
// same-origin nginx proxy either way, so there's no origin worth rejecting.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// connect upgrades an already-authenticated request (wsAuth ran first) to
// a WebSocket and registers it in the hub. Clients don't send anything
// meaningful over this socket in v1 — messages go through the REST
// endpoints — but the read loop below is still needed to detect
// disconnects/pings and release the connection.
func (h *realtimeHandler) connect(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		respondError(c, exceptions.Unauthorized(""))
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("realtime: upgrade failed:", err)
		return
	}

	client := h.hub.Register(userID, conn)
	defer h.hub.Unregister(userID, client)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
