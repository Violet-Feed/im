package model

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/proto_gen/im"
	"im/util"
	"strconv"
	"time"
)

const SettingInfoExpireTime = 6 * time.Hour

type ConversationSettingInfo struct {
	Id             int64     `gorm:"column:id" json:"id"`
	UserId         int64     `gorm:"column:user_id" json:"user_id"`
	ConShortId     int64     `gorm:"column:con_short_id" json:"con_short_id"`
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

func InsertSettingInfos(ctx context.Context, settings []*ConversationSettingInfo) error {
	logrus.Infof("[InsertSettingInfos] settings = %v", settings)
	err := dal.MysqlDB.Create(&settings).Error
	if err != nil {
		logrus.Errorf("[InsertSettingInfo] mysql insert setting err. err = %v", err)
		return err
	}
	keys, values := make([]string, 0), make([]string, 0)
	for _, setting := range settings {
		key := fmt.Sprintf("setting:%d:%d", setting.ConShortId, setting.UserId)
		value, err := json.Marshal(setting)
		if err != nil {
			keys = append(keys, key)
			values = append(values, string(value))
		}
	}
	_ = dal.RedisServer.BatchSet(ctx, keys, values, 1*time.Minute)
	return nil
}

func GetSettingInfo(ctx context.Context, userId int64, conShortIds []int64) (map[int64]*ConversationSettingInfo, error) {
	var keys []string
	var missIds []int64
	settingsMap := make(map[int64]*ConversationSettingInfo)
	for _, id := range conShortIds {
		key := fmt.Sprintf("setting:%d:%d", userId, id)
		keys = append(keys, key)
	}
	results, err := dal.RedisServer.MGet(ctx, keys)
	if err != nil {
		logrus.Errorf("[GetSettingInfo] redis mget err. err = %v", err)
		missIds = conShortIds
	} else {
		for i, result := range results {
			if result != "" {
				var setting *ConversationSettingInfo
				if err := json.Unmarshal([]byte(result), &setting); err != nil {
					settingsMap[setting.ConShortId] = setting
					continue
				}
			}
			missIds = append(missIds, conShortIds[i])
		}
	}
	if len(missIds) == 0 {
		return settingsMap, nil
	}
	var settings []*ConversationSettingInfo
	err = dal.MysqlDB.Where("user_id = (?) and con_short_id in (?) ", userId, missIds).Find(&settings).Error
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
		ConType:    req.GetConType(),
		ModifyTime: time.Now(),
		Extra:      req.GetExtra(),
	}
	return setting
}

func PackSettingInfo(model *ConversationSettingInfo) *im.ConversationSettingInfo {
	if model == nil {
		return nil
	}
	setting := &im.ConversationSettingInfo{
		UserId:       util.Int64(model.UserId),
		ConShortId:   util.Int64(model.ConShortId),
		ConType:      util.Int32(model.ConType),
		MinIndex:     util.Int64(model.MinIndex),
		TopTimeStamp: util.Int64(model.TopTimeStamp),
		PushStatus:   util.Int32(model.PushStatus),
		ModifyTime:   util.Int64(model.ModifyTime.Unix()),
		Extra:        util.String(model.Extra),
	}
	return setting
}

func SetReadIndexStart(ctx context.Context, conShortId int64, userIds []int64, index int64) error {
	var values map[string]string
	for _, userId := range userIds {
		key := fmt.Sprintf("read_start:%d:%d", userId, conShortId)
		values[key] = strconv.FormatInt(index, 10)
	}
	err := dal.KvrocksServer.MSet(ctx, values)
	if err != nil {
		logrus.Errorf("[SetReadIndexStart] kvrocks mset err. err = %v", err)
		return err
	}
	return nil
}

func GetReadIndexStart(ctx context.Context, conShortIds []int64, userId int64) (map[int64]int64, error) {
	var keys []string
	for _, conShortId := range conShortIds {
		key := fmt.Sprintf("read_start:%d:%d", userId, conShortId)
		keys = append(keys, key)
	}
	results, err := dal.KvrocksServer.MGet(ctx, keys)
	if err != nil {
		logrus.Errorf("[GetReadIndexStart] kvrocks mget err. err = %v", err)
		return nil, err
	}
	indexMap := make(map[int64]int64)
	for i, conShortId := range conShortIds {
		if results[i] != "" {
			readIndex, _ := strconv.ParseInt(results[i], 10, 64)
			indexMap[conShortId] = readIndex
		} else {
			indexMap[conShortId] = 0
		}
	}
	return indexMap, nil
}

func SetReadIndexEnd(ctx context.Context, conShortId int64, userIds []int64, index int64) error {
	var values map[string]string
	for _, userId := range userIds {
		key := fmt.Sprintf("read_end:%d:%d", userId, conShortId)
		values[key] = strconv.FormatInt(index, 10)
	}
	err := dal.KvrocksServer.MSet(ctx, values)
	if err != nil {
		logrus.Errorf("[SetReadIndexEnd] kvrocks mset err. err = %v", err)
		return err
	}
	return nil
}

func GetReadIndexEnd(ctx context.Context, conShortIds []int64, userId int64) (map[int64]int64, error) {
	var keys []string
	for _, conShortId := range conShortIds {
		key := fmt.Sprintf("read_end:%d:%d", userId, conShortId)
		keys = append(keys, key)
	}
	results, err := dal.KvrocksServer.MGet(ctx, keys)
	if err != nil {
		logrus.Errorf("[GetReadIndexEnd] kvrocks mget err. err = %v", err)
		return nil, err
	}
	indexMap := make(map[int64]int64)
	for i, conShortId := range conShortIds {
		if results[i] != "" {
			readIndex, _ := strconv.ParseInt(results[i], 10, 64)
			indexMap[conShortId] = readIndex
		} else {
			indexMap[conShortId] = 0
		}
	}
	return indexMap, nil
}

func SetReadBadge(ctx context.Context, conShortId int64, userIds []int64, count int64) error {
	var values map[string]string
	for _, userId := range userIds {
		key := fmt.Sprintf("read_badge:%d:%d", userId, conShortId)
		values[key] = strconv.FormatInt(count, 10)
	}
	err := dal.KvrocksServer.MSet(ctx, values)
	if err != nil {
		logrus.Errorf("[SetReadBadge] kvrocks mset err. err = %v", err)
		return err
	}
	return nil
}

func GetReadBadge(ctx context.Context, conShortIds []int64, userId int64) (map[int64]int64, error) {
	var keys []string
	for _, conShortId := range conShortIds {
		key := fmt.Sprintf("read_badge:%d:%d", userId, conShortId)
		keys = append(keys, key)
	}
	results, err := dal.KvrocksServer.MGet(ctx, keys)
	if err != nil {
		logrus.Errorf("[GetReadBadge] kvrocks mget err. err = %v", err)
		return nil, err
	}
	countMap := make(map[int64]int64)
	for i, conShortId := range conShortIds {
		if results[i] != "" {
			readIndex, _ := strconv.ParseInt(results[i], 10, 64)
			countMap[conShortId] = readIndex
		} else {
			countMap[conShortId] = 0
		}
	}
	return countMap, nil
}
