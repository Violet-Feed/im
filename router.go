package main

import (
	"github.com/gin-gonic/gin"
	"im/handler"
)

func Router(r *gin.Engine) *gin.Engine {
	r.GET("/message", handler.GetMessage)
	r.GET("/ws/:id", handler.WebsocketHandler)
	r.GET("/test", handler.TestWs)
	kv := r.Group("/kv")
	{
		kv.GET("/set", handler.Set)
		kv.GET("/get", handler.Get)
	}
	r.GET("/mq", handler.SendMessage)

	return r
}
