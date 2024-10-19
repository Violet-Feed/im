package util

import (
	"github.com/bwmarrin/snowflake"
	"github.com/sirupsen/logrus"
)

var (
	MsgIdGenerator  *snowflake.Node
	ConvIdGenerator *snowflake.Node
	ConnIdGenerator *snowflake.Node
)

func init() {
	var err error
	MsgIdGenerator, err = snowflake.NewNode(0)
	if err != nil {
		logrus.Fatalf("[main] MsgIdGenerator NewNode err. err = %v", err)
	}
	ConvIdGenerator, err = snowflake.NewNode(0)
	if err != nil {
		logrus.Fatalf("[main] ConvIdGenerator NewNode err. err = %v", err)
	}
	ConnIdGenerator, err = snowflake.NewNode(0)
	if err != nil {
		logrus.Fatalf("[main] ConnIdGenerator NewNode err. err = %v", err)
	}
}
