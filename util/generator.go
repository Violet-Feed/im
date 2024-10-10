package util

import (
	"github.com/bwmarrin/snowflake"
	"github.com/sirupsen/logrus"
)

var MessageIdGenerator *snowflake.Node

func init() {
	var err error
	MessageIdGenerator, err = snowflake.NewNode(0)
	if err != nil {
		logrus.Fatalf("[main] snowflake NewNode err. err = %v", err)
	}
}
