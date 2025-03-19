package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"im/biz"
	"im/dal"
	"im/proto_gen/im"
	"im/util"
	"net/http"
	"time"
)

type ConnInfo struct {
	UserId   int64  `json:"user_id"`
	DeviceId string `json:"device_id"`
	Platform string `json:"platform"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// TODO:改成select语句
func WebsocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("[WebsocketHandler] upgrade websocket err. err = %v", err)
		c.JSON(http.StatusOK, im.StatusCode_Server_Error)
		return
	}
	defer conn.Close()

	connId := util.ConnIdGenerator.Generate().String()
	biz.Connections.Store(connId, conn)
	defer biz.Connections.Delete(connId)

	userIdStr, _ := c.Get("userId")
	deviceId, _ := c.Get("deviceId")
	platform, _ := c.Get("platform")
	userId := userIdStr.(int64)
	key := fmt.Sprintf("conn:%d", userId)
	connInfo, _ := json.Marshal(ConnInfo{UserId: userId, DeviceId: deviceId.(string), Platform: platform.(string)})
	err = dal.RedisServer.HSet(c, key, connId, connInfo)
	if err != nil {
		logrus.Errorf("[WebsocketHandler] redis hset err. err = %v", err)
		c.JSON(http.StatusOK, im.StatusCode_Server_Error)
		return
	}
	defer func() {
		err = dal.RedisServer.HDel(c, key, connId)
		if err != nil {
			logrus.Errorf("[WebsocketHandler] redis hdel err. err = %v", err)
		}
	}()
	schedule := time.NewTicker(1 * time.Minute)
	defer schedule.Stop()
	go func() {
		for range schedule.C {
			exist, err := dal.RedisServer.HExists(c, key, connId)
			if err == nil && !exist {
				_ = dal.RedisServer.HSet(c, key, connId, connInfo)
				err = conn.WriteMessage(websocket.TextMessage, []byte("need init"))
				if err != nil {
					logrus.Warnf("[WebsocketHandler] write message err. err = %v", err)
					conn.Close()
				}
			}
		}
	}()
	heartBeat := time.NewTicker(5 * time.Second)
	defer heartBeat.Stop()
	var alive int32
	go func() {
		for range heartBeat.C {
			if alive == -3 {
				logrus.Warnf("[WebsocketHandler] heartbeat timeout.")
				conn.Close()
			}
			alive--
		}
	}()
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			logrus.Warnf("[WebsocketHandler] read message err. err = %v", err)
			return
		}
		logrus.Infof("[WebsocketHandler] receive message. messageType = %v, message = %v", messageType, string(message))
		switch string(message) {
		case "ping":
			alive = 0
		default:
			err = conn.WriteMessage(websocket.TextMessage, []byte("invalid message"))
		}
		if err != nil {
			logrus.Warnf("[WebsocketHandler] write message err. err = %v", err)
			return
		}
	}
}
