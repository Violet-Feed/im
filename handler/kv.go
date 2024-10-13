package handler

import (
	"github.com/gin-gonic/gin"
	"im/dal"
	"net/http"
	"strconv"
)

func Set(c *gin.Context) {
	value := c.Query("val")
	err := dal.KvrocksServer.Set(c, "a", value)
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

func Cas(c *gin.Context) {
	oldValue := c.Query("old")
	newValue := c.Query("new")
	resp, err := dal.KvrocksServer.Cas(c, "a", oldValue, newValue)
	if err != nil {
		c.String(http.StatusOK, err.Error())
	} else {
		c.String(http.StatusOK, strconv.FormatInt(resp, 10))
	}
}
