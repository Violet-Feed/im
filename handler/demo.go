package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
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
