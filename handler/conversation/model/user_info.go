package model

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"im/dal"
	"time"
)

type ConversationUserInfo struct {
	Id         int64     `gorm:"column:id" json:"id"`
	ConShortId int64     `gorm:"column:con_short_id" json:"con_short_id"`
	UserId     int64     `gorm:"column:user_id" json:"user_id"`
	Privilege  int32     `gorm:"column:privilege" json:"level"`
	NickName   string    `gorm:"column:nick_name" json:"nick_name"`
	BlockTime  time.Time `gorm:"column:blocked" json:"blocked"`
	Operator   int64     `gorm:"column:operator" json:"operator"`
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
	ModifyTime time.Time `gorm:"column:modify_time" json:"modify_time"`
	Status     int32     `gorm:"column:status" json:"status"`
	Extra      string    `gorm:"column:extra" json:"extra"`
}

func (c *ConversationUserInfo) TableName() string {
	return "conversation_user_info"
}

func GetUserCount(ctx context.Context, conShortId int64) (int, error) {
	//本地缓存+redis？
	key := fmt.Sprintf("member:%d", conShortId)
	count, err := dal.RedisServer.ZCard(ctx, key)
	if err == nil && count > 0 {
		return int(count), nil
	}
	var res []*ConversationUserInfo
	err = dal.MysqlDB.Select("user_id", "create_time").Where("con_short_id=?", conShortId).Find(&res).Error
	if err != nil {
		logrus.Errorf("mysql get user count err. err = %v", err)
		return 0, err
	}
	go func() {
		err := dal.RedisServer.Del(ctx, key)
		if err != nil {
			return
		}
		var values []redis.Z
		for i := 0; i < len(res); i++ {
			values = append(values, redis.Z{
				Member: res[i].UserId,
				Score:  float64(res[i].CreateTime.Unix()),
			})
		}
		err = dal.RedisServer.ZAdd(ctx, key, values)
	}()
	return len(res), nil
}
