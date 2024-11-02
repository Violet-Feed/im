package model

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/proto_gen/im"
	"im/util"
	"time"
)

const SettingInfoExpireTime = 6 * time.Hour

type ConversationSettingInfo struct {
	Id             int64     `gorm:"column:id" json:"id"`
	UserId         int64     `gorm:"column:user_id" json:"user_id"`
	ConShortId     int64     `gorm:"column:con_short_id" json:"con_short_id"`
	ConId          string    `gorm:"column:con_id" json:"con_id"`
	ConType        int32     `gorm:"column:con_type" json:"con_type"`
	MinIndex       int64     `gorm:"column:min_index" json:"min_index"`
	TopTimeStamp   int64     `gorm:"column:top_time_stamp" json:"top_time_stamp"`
	PushStatus     int32     `gorm:"column:push_status" json:"push_status"`
	ModifyTime     time.Time `gorm:"column:modify_time" json:"modify_time"`
	Extra          string    `gorm:"column:extra" json:"extra"`
	ReadIndexEnd   int64     `gorm:"-" json:"read_index_end"`
	ReadBadgeCount int64     `gorm:"-" json:"read_badge_count"`
}

func (c *ConversationSettingInfo) TableName() string {
	return "conversation_setting_info"
}

func InsertSettingInfo(ctx context.Context, setting *ConversationSettingInfo) error {
	err := dal.MysqlDB.Create(setting).Error
	if err != nil {
		logrus.Errorf("[InsertSettingInfo] mysql insert setting err. err = %v", err)
		return err
	}
	settingByte, err := json.Marshal(setting)
	if err == nil {
		key := fmt.Sprintf("setting:%d:%d", setting.ConShortId, setting.UserId)
		_ = dal.RedisServer.Set(ctx, key, string(settingByte), 1*time.Minute)
	}
	return nil
}

func GetSettingInfo(ctx context.Context, userId int64, conShortIds []int64) (map[int64]*ConversationSettingInfo, error) {
	var keys []string
	var lostIds []int64
	settingsMap := make(map[int64]*ConversationSettingInfo)
	for _, id := range conShortIds {
		key := fmt.Sprintf("setting:%d:%d", userId, id)
		keys = append(keys, key)
	}
	results, err := dal.RedisServer.MGet(ctx, keys)
	if err != nil {
		logrus.Errorf("[GetSettingInfo] redis mget err. err = %v", err)
		lostIds = conShortIds
	} else {
		for i, result := range results {
			if result != "" {
				var setting *ConversationSettingInfo
				if err := json.Unmarshal([]byte(result), &setting); err != nil {
					settingsMap[setting.ConShortId] = setting
					continue
				}
			}
			lostIds = append(lostIds, conShortIds[i])
		}
	}
	if len(lostIds) == 0 {
		return settingsMap, nil
	}
	var settings []*ConversationSettingInfo
	err = dal.MysqlDB.Where("con_short_id in (?)", lostIds).Find(settings).Error
	if err != nil {
		logrus.Errorf("[GetSettingInfo] mysql select err. err = %v", err)
		return nil, err
	}
	for _, setting := range settings {
		settingsMap[setting.ConShortId] = setting
	}
	go AsyncSetSettingCache(ctx, settings)
	return settingsMap, nil
}

func AsyncSetSettingCache(ctx context.Context, settings []*ConversationSettingInfo) {
	var keys, values []string
	for _, setting := range settings {
		key := fmt.Sprintf("setting:%d:%d", setting.UserId, setting.ConShortId)
		value, err := json.Marshal(setting)
		if err != nil {
			keys = append(keys, key)
			values = append(values, string(value))
		}
	}
	_ = dal.RedisServer.BatchSet(ctx, keys, values, SettingInfoExpireTime)
}

func PackSettingModel(userId int64, conShortId int64, req *im.CreateConversationRequest) *ConversationSettingInfo {
	setting := &ConversationSettingInfo{
		UserId:     userId,
		ConShortId: conShortId,
		ConId:      req.GetConId(),
		ConType:    req.GetConType(),
		Extra:      req.GetExtra(),
	}
	curTime := time.Now()
	setting.ModifyTime = curTime
	return setting
}

func PackSettingInfo(model *ConversationSettingInfo) *im.ConversationSettingInfo {
	if model == nil {
		return nil
	}
	setting := &im.ConversationSettingInfo{
		UserId:       util.Int64(model.UserId),
		ConShortId:   util.Int64(model.ConShortId),
		ConId:        util.String(model.ConId),
		ConType:      util.Int32(model.ConType),
		MinIndex:     util.Int64(model.MinIndex),
		TopTimeStamp: util.Int64(model.TopTimeStamp),
		PushStatus:   util.Int32(model.PushStatus),
		ModifyTime:   util.Int64(model.ModifyTime.Unix()),
		Extra:        util.String(model.Extra),
	}
	return setting
}
