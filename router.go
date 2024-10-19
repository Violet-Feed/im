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
		if c.FullPath() == "/ws/create" {
			c.Set("userId", int64(1844310578969968640))
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
	r.GET("/rpc", handler.GetMessage)
	r.GET("/ws", handler.WebsocketHandler)

	message := r.Group("/message")
	{
		message.POST("/send", handler.Send)
		message.POST("/modify")
		message.POST("/recall")
		message.POST("/delete")
		message.POST("/forward")
		message.POST("/pin")
		message.POST("/mark_read")
		message.POST("/get_by_init")
		message.POST("/get_by_conv")
	}
	conversation := r.Group("/conversation")
	{
		conversation.POST("/create")
		conversation.POST("/delete")
		conversation.POST("/pin")
		conversation.POST("/get_info")
		conversation.POST("/modify_info")
		conversation.POST("/disband")
		conversation.POST("/join")
		conversation.POST("/exit")
		conversation.POST("/share")
	}
	return r
}
