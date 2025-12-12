package util

import (
	"github.com/bwmarrin/snowflake"
	"github.com/sirupsen/logrus"
)

var (
	MsgIdGenerator    *snowflake.Node
	ConIdGenerator    *snowflake.Node
	NoticeIdGenerator *snowflake.Node
)

func init() {
	var err error
	MsgIdGenerator, err = snowflake.NewNode(0)
	if err != nil {
		logrus.Fatalf("[main] MsgIdGenerator NewNode err. err = %v", err)
	}
	ConIdGenerator, err = snowflake.NewNode(0)
	if err != nil {
		logrus.Fatalf("[main] ConIdGenerator NewNode err. err = %v", err)
	}
	NoticeIdGenerator, err = snowflake.NewNode(0)
	if err != nil {
		logrus.Fatalf("[main] NoticeIdGenerator NewNode err. err = %v", err)
	}
}
