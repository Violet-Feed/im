package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/util"
	"net/http"
	"sync"
)

var Connections sync.Map

type ConnInfo struct {
	UserId   int64  `json:"user_id"`
	DeviceId int64  `json:"device_id"`
	Platform string `json:"platform"`
}

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
		logrus.Errorf("[WebsocketHandler] upgrade websocket err. err = %v", err)
		c.JSON(http.StatusOK, StateCode_Internal_ERROR)
		return
	}
	defer conn.Close()

	connId := util.ConnIdGenerator.Generate().String()
	Connections.Store(connId, conn)
	defer Connections.Delete(connId)

	userIdStr, _ := c.Get("userId")
	userId := userIdStr.(int64)
	key := fmt.Sprintf("conn:%d", userId)
	connInfo, _ := json.Marshal(ConnInfo{UserId: userId})
	err = dal.RedisServer.HSet(c, key, connId, connInfo)
	if err != nil {
		logrus.Errorf("[WebsocketHandler] redis hset err. err = %v", err)
		c.JSON(http.StatusOK, StateCode_Internal_ERROR)
		return
	}
	defer func() {
		err = dal.RedisServer.HDel(c, key, connId)
		if err != nil {
			logrus.Errorf("[WebsocketHandler] redis hdel err. err = %v", err)
		}
	}()
	defer logrus.Warnf("[WebsocketHandler] defer %v", connId)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			logrus.Errorf("[WebsocketHandler] read message err. err = %v", err)
		}
		logrus.Infof("[WebsocketHandler] receive message. messageType = %v, message = %v", messageType, message)

		switch string(message) {
		case "ping":
			err = conn.WriteMessage(websocket.TextMessage, []byte("pong"))
		default:
			err = conn.WriteMessage(websocket.TextMessage, []byte("invalid message"))
		}
		if err != nil {
			logrus.Errorf("[WebsocketHandler] write message err. err = %v", err)
		}
	}
}
