package model

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/proto_gen/im"
	"time"
)

const CoreInfoExpireTime = 24 * time.Hour

type ConversationCoreInfo struct {
	Id          int64     `gorm:"column:id" json:"id"`
	ConShortId  int64     `gorm:"column:con_short_id" json:"con_short_id"`
	ConId       string    `gorm:"column:con_id" json:"con_id"`
	ConType     int32     `gorm:"column:con_type" json:"con_type"`
	Name        string    `gorm:"column:name" json:"name"`
	AvatarUri   string    `gorm:"column:avatar_uri" json:"avatar_uri"`
	Description string    `gorm:"column:description" json:"description"`
	Notice      string    `gorm:"column:notice" json:"notice"`
	OwnerId     int64     `gorm:"column:owner_id" json:"owner_id"`
	CreateTime  time.Time `gorm:"column:create_time" json:"create_time"`
	ModifyTime  time.Time `gorm:"column:modify_time" json:"modify_time"`
	Status      int32     `gorm:"column:status" json:"status"`
	Extra       string    `gorm:"column:extra" json:"extra"`
	MemberCount int32     `gorm:"-" json:"member_count"`
}

func (c *ConversationCoreInfo) TableName() string {
	return "conversation_core_info"
}

func InsertCoreInfo(ctx context.Context, core *ConversationCoreInfo) error {
	err := dal.MysqlDB.Create(core).Error
	if err != nil {
		logrus.Errorf("[InsertCoreInfo] mysql insert core err. err = %v", err)
		return err
	}
	coreByte, err := json.Marshal(core)
	if err == nil {
		key := fmt.Sprintf("core:%d", core.ConShortId)
		_ = dal.RedisServer.Set(ctx, key, string(coreByte), 1*time.Minute)
	}
	return nil
}

func GetCoreInfos(ctx context.Context, conShortIds []int64) (map[int64]*ConversationCoreInfo, error) {
	var keys []string
	var missIds []int64
	coresMap := make(map[int64]*ConversationCoreInfo)
	for _, id := range conShortIds {
		key := fmt.Sprintf("core:%d", id)
		keys = append(keys, key)
	}
	results, err := dal.RedisServer.MGet(ctx, keys)
	if err != nil {
		logrus.Errorf("[GetCoreInfos] redis mget err. err = %v", err)
		missIds = conShortIds
	} else {
		for i, result := range results {
			if result != "" {
				var core *ConversationCoreInfo
				if err := json.Unmarshal([]byte(result), &core); err != nil {
					coresMap[core.ConShortId] = core
					continue
				}
			}
			missIds = append(missIds, conShortIds[i])
		}
	}
	if len(missIds) == 0 {
		return coresMap, nil
	}
	var cores []*ConversationCoreInfo
	err = dal.MysqlDB.Where("con_short_id in (?)", missIds).Find(&cores).Error
	if err != nil {
		logrus.Errorf("[GetCoreInfos] mysql select err. err = %v", err)
		return nil, err
	}
	for _, core := range cores {
		coresMap[core.ConShortId] = core
	}
	go AsyncSetCoreCache(ctx, cores)
	return coresMap, nil
}

func GetCoreInfoByConId(ctx context.Context, conId string) (*ConversationCoreInfo, error) {
	var cores []*ConversationCoreInfo
	err := dal.MysqlDB.Where("con_id = ?", conId).First(&cores).Error
	if err != nil {
		logrus.Errorf("[GetCoreInfoByConId] mysql select err. err = %v", err)
		return nil, err
	}
	if len(cores) == 0 {
		return nil, nil
	}
	return cores[0], nil
}

func AsyncSetCoreCache(ctx context.Context, cores []*ConversationCoreInfo) {
	var keys, values []string
	for _, core := range cores {
		key := fmt.Sprintf("core:%d", core.ConShortId)
		value, err := json.Marshal(core)
		if err != nil {
			keys = append(keys, key)
			values = append(values, string(value))
		}
	}
	_ = dal.RedisServer.BatchSet(ctx, keys, values, CoreInfoExpireTime)
}

func PackCoreModel(conShortId int64, req *im.CreateConversationRequest) *ConversationCoreInfo {
	core := &ConversationCoreInfo{
		ConShortId:  conShortId,
		ConId:       req.GetConId(),
		ConType:     req.GetConType(),
		Name:        req.GetName(),
		AvatarUri:   req.GetAvatarUri(),
		Description: req.GetDescription(),
		Notice:      req.GetNotice(),
		OwnerId:     req.GetOwnerId(),
		Extra:       req.GetExtra(),
	}
	//TODO:获取req里的time？
	curTime := time.Now()
	core.CreateTime = curTime
	core.ModifyTime = curTime
	return core
}

func PackCoreInfo(model *ConversationCoreInfo) *im.ConversationCoreInfo {
	if model == nil {
		return nil
	}
	core := &im.ConversationCoreInfo{
		ConShortId:  model.ConShortId,
		ConId:       model.ConId,
		ConType:     model.ConType,
		Name:        model.Name,
		AvatarUri:   model.AvatarUri,
		Description: model.Description,
		Notice:      model.Notice,
		OwnerId:     model.OwnerId,
		CreateTime:  model.CreateTime.Unix(),
		ModifyTime:  model.ModifyTime.Unix(),
		Status:      model.Status,
		Extra:       model.Extra,
	}
	return core
}
