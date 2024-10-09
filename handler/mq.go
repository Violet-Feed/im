package handler

import (
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

var p rocketmq.Producer

func init() {
	var err error
	p, err = rocketmq.NewProducer(producer.WithNameServer([]string{"127.0.0.1:9876"}))
	if err != nil {
		log.Fatalf("[init] rocketmq producer create err. err = %v", err)
	}
	err = p.Start()
	if err != nil {
		log.Fatalf("[init] rocketmq producer run err. err = %v", err)
	}
}

func SendMessage(c *gin.Context) {
	message := c.Query("message")
	fmt.Println(message)
	_, err := p.SendSync(c, &primitive.Message{
		Topic: "test",
		Body:  []byte(message),
	})
	if err != nil {
		c.String(http.StatusOK, "failed")
	} else {
		c.String(http.StatusOK, "success")
	}
}
