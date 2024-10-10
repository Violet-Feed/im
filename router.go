package main

import (
	"github.com/gin-gonic/gin"
	"im/handler"
	"im/util"
	"net/http"
	"strings"
)

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.FullPath() == "/login" {
			c.Next()
		}
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Missing token"})
			return
		}
		parts := strings.Split(token, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Invalid token format"})
			return
		}
		userId, err := util.ParseUserToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Invalid token"})
			return
		}
		c.Set("userId", userId)
		c.Next()
	}
}

func Router(r *gin.Engine) *gin.Engine {
	r.Use(authMiddleware())
	r.GET("/demo", handler.GetMessage)
	r.GET("/ws/:id", handler.WebsocketHandler)
	r.GET("/test", handler.TestWs)
	kv := r.Group("/kv")
	{
		kv.GET("/set", handler.Set)
		kv.GET("/get", handler.Get)
	}
	r.GET("/mq", handler.SendMessage)

	message := r.Group("/message")
	{
		message.POST("/send", handler.Send)
	}
	return r
}
