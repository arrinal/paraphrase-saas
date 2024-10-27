package api

import (
	"log"
	"net/http"

	"github.com/arrinal/paraphrase-saas/internal/websocket"
	"github.com/gin-gonic/gin"
	gorillaWs "github.com/gorilla/websocket"
)

var upgrader = gorillaWs.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, implement proper origin checking
	},
}

func HandleWebSocket(hub *websocket.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}

		client := &websocket.Client{
			Hub:    hub,
			Conn:   conn,
			Send:   make(chan []byte, 256),
			UserID: userID.(uint),
		}

		client.Hub.Register <- client // Use capitalized Register

		// Start goroutines for reading and writing
		go client.Write()
		go client.Read()
	}
}
