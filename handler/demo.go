package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"im/dal"
	"net/http"
)

func GetMessage(c *gin.Context) {
	param := c.Query("param")
	fmt.Println(param)
	str, err := dal.DemoServer.GetMessage(c, param)
	if err != nil {
		c.String(http.StatusOK, "err")
	} else {
		c.String(http.StatusOK, str)
	}
}

func TestWs(c *gin.Context) {
	userId := c.Query("id")
	message := c.Query("message")
	connInter, isExist := connects.Load(userId)
	fmt.Println(connInter, isExist)
	if conn, ok := connInter.(*websocket.Conn); ok && isExist {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			c.String(http.StatusOK, "err")
		} else {
			c.String(http.StatusOK, "ok")
		}
	} else {
		c.String(http.StatusOK, "not exist")
	}
}
