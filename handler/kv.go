package handler

import (
	"github.com/gin-gonic/gin"
	"im/service"
	"net/http"
)

func Set(c *gin.Context) {
	err := service.KvrocksServer.Set(c, "a", "1")
	if err != nil {
		c.String(http.StatusOK, err.Error())
	} else {
		c.String(http.StatusOK, "success")
	}
}

func Get(c *gin.Context) {
	resp, err := service.KvrocksServer.Get(c, "a")
	if err != nil {
		c.String(http.StatusOK, err.Error())
	} else {
		c.String(http.StatusOK, resp)
	}
}
