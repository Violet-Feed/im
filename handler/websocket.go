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
		return
	}

	defer conn.Close()

	userId := c.Param("id")
	connects.Store(userId, conn)
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[WebsocketHandler] read message err. err = %v", err)
			return
		}
		log.Printf("[WebsocketHandler] receive message. messageType = %v, message = %v", messageType, message)

		err = conn.WriteMessage(websocket.TextMessage, []byte("pong"))
		if err != nil {
			log.Printf("[WebsocketHandler] write message err. err = %v", err)
			return
		}
	}
}
