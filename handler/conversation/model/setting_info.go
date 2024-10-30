package model

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"im/dal"
	"time"
)

type ConversationSettingInfo struct {
	Id             int64     `gorm:"column:id" json:"id"`
	UserId         int64     `gorm:"column:user_id" json:"user_id"`
	ConShortId     int64     `gorm:"column:con_short_id" json:"con_short_id"`
	ConId          string    `gorm:"column:con_id" json:"con_id"`
	ConType        int32     `gorm:"column:con_type" json:"con_type"`
	MinIndex       int64     `gorm:"column:min_index" json:"min_index"`
	TopTimeStamp   int64     `gorm:"column:top_time_stamp" json:"set_top_time"`
	MuteStatus     int32     `gorm:"column:mute_status" json:"push_status"`
	ModifyTime     time.Time `gorm:"column:modify_time" json:"modify_time"`
	Extra          string    `gorm:"column:extra" json:"extra"`
	ReadIndex      int64     `gorm:"-" json:"read_index"`
	ReadBadgeCount int32     `gorm:"-" json:"read_badge_count"`
}

func (c *ConversationSettingInfo) TableName() string {
	return "conversation_setting_info"
}

func InsertSettingInfo(ctx context.Context, setting *ConversationSettingInfo) error {
	err := dal.MysqlDB.Create(setting).Error
	if err != nil {
		logrus.Errorf("mysql insert setting err. err = %v", err)
		return err
	}
	settingByte, err := json.Marshal(setting)
	if err == nil {
		key := fmt.Sprintf("setting:%d:%d", setting.ConShortId, setting.UserId)
		_ = dal.RedisServer.Set(ctx, key, string(settingByte), 1*time.Minute)
	}
	return nil
}
