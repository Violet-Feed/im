package handler

import (
	"github.com/gin-gonic/gin"
	"im/dal"
	"net/http"
)

func Set(c *gin.Context) {
	err := dal.KvrocksServer.Set(c, "a", "1")
	if err != nil {
		c.String(http.StatusOK, err.Error())
	} else {
		c.String(http.StatusOK, "success")
	}
}

func Get(c *gin.Context) {
	resp, err := dal.KvrocksServer.Get(c, "a")
	if err != nil {
		c.String(http.StatusOK, err.Error())
	} else {
		c.String(http.StatusOK, resp)
	}
}
