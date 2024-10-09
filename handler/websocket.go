package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var connects sync.Map

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebsocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WebsocketHandler] upgrade websocket err. err = %v", err)
		resp := Response{Code: 201, Message: "connect err", Data: nil}
		c.JSON(http.StatusOK, resp)
		return
	}

	defer conn.Close()

	userId := c.Param("id")
	connects.Store(userId, conn)
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[WebsocketHandler] read message err. err = %v", err)
			continue
		}
		log.Printf("[WebsocketHandler] receive message. messageType = %v, message = %v", messageType, message)

		switch string(message) {
		case "ping":
			err = conn.WriteMessage(websocket.TextMessage, []byte("pong"))
		default:
			err = conn.WriteMessage(websocket.TextMessage, []byte("invalid message"))
		}
		if err != nil {
			log.Printf("[WebsocketHandler] write message err. err = %v", err)
		}
	}
}
