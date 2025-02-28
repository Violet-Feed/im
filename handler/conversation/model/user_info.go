package model

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/proto_gen/im"
	"im/util"
	"strconv"
	"time"
)

const UserInfoExpireTime = 24 * time.Hour

type ConversationUserInfo struct {
	Id             int64     `gorm:"column:id" json:"id"`
	ConShortId     int64     `gorm:"column:con_short_id" json:"con_short_id"`
	UserId         int64     `gorm:"column:user_id" json:"user_id"`
	Privilege      int32     `gorm:"column:privilege" json:"level"`
	NickName       string    `gorm:"column:nick_name" json:"nick_name"`
	BlockTimeStamp int64     `gorm:"column:block_time_stamp" json:"block_time_stamp"`
	Operator       int64     `gorm:"column:operator" json:"operator"`
	CreateTime     time.Time `gorm:"column:create_time" json:"create_time"`
	ModifyTime     time.Time `gorm:"column:modify_time" json:"modify_time"`
	Status         int32     `gorm:"column:status" json:"status"`
	Extra          string    `gorm:"column:extra" json:"extra"`
}

func (c *ConversationUserInfo) TableName() string {
	return "conversation_user_info"
}

func InsertUserInfos(ctx context.Context, conShortId int64, users []*ConversationUserInfo) error {
	err := dal.MysqlDB.Create(users).Error
	if err != nil {
		logrus.Errorf("[InsertUserInfos] mysql insert users err. err = %v", err)
		return err
	}
	var keys, values []string
	var zSetValues []redis.Z
	for _, user := range users {
		key := fmt.Sprintf("member:%v:%v", conShortId, user.UserId)
		valueByte, err := json.Marshal(user)
		if err != nil {
			logrus.Errorf("[InsertUserInfos] json marshal err. err = %v", err)
		} else {
			keys = append(keys, key)
			values = append(values, string(valueByte))
		}
		zSetValues = append(zSetValues, redis.Z{
			Member: user.UserId,
			Score:  float64(user.CreateTime.Unix()),
		})
	}
	_ = dal.RedisServer.BatchSet(ctx, keys, values, UserInfoExpireTime)
	key := fmt.Sprintf("member:%v", conShortId)
	_ = dal.RedisServer.ZAdd(ctx, key, zSetValues)
	return nil
}

func GetUserCount(ctx context.Context, conShortId int64) (int, error) {
	//TODO:本地缓存+redis？
	key := fmt.Sprintf("member:%d", conShortId)
	count, err := dal.RedisServer.ZCard(ctx, key)
	if err == nil && count > 0 {
		return int(count), nil
	}
	var users []*ConversationUserInfo
	err = dal.MysqlDB.Select("user_id", "create_time").Where("con_short_id=?", conShortId).Find(&users).Error
	if err != nil {
		logrus.Errorf("[GetUserCount] mysql get user count err. err = %v", err)
		return 0, err
	}
	if count == 0 {
		go AsyncSetUserZSetCache(ctx, key, users)
	}
	return len(users), nil
}

func GetUserInfos(ctx context.Context, conShortId int64, userIds []int64, useCache bool) (map[int64]*ConversationUserInfo, error) {
	var keys []string
	var missIds []int64
	userMap := make(map[int64]*ConversationUserInfo)
	for _, userId := range userIds {
		key := fmt.Sprintf("user:%v:%v", conShortId, userId)
		keys = append(keys, key)
	}
	if useCache {
		results, err := dal.RedisServer.MGet(ctx, keys)
		if err != nil {
			logrus.Errorf("[GetUserInfos] redis mget err. err = %v", err)
			missIds = userIds
		} else {
			for i, result := range results {
				if result != "" {
					var user *ConversationUserInfo
					if err := json.Unmarshal([]byte(result), &user); err != nil {
						userMap[user.UserId] = user
						continue
					}
				}
				missIds = append(missIds, userIds[i])
			}
			if len(missIds) == 0 {
				return userMap, nil
			}
		}
	} else {
		missIds = userIds
	}
	var users []*ConversationUserInfo
	err := dal.MysqlDB.Where("con_short_id = (?) and user_id in (?)", conShortId, missIds).Find(&users).Error
	if err != nil {
		logrus.Errorf("[GetUserInfos] mysql select err. err = %v", err)
		return nil, err
	}
	for _, user := range users {
		userMap[user.UserId] = user
	}
	if useCache {
		go AsyncSetUserCache(ctx, conShortId, users)
	}
	return userMap, nil

}

func AsyncSetUserCache(ctx context.Context, conShortId int64, users []*ConversationUserInfo) {
	var keys, values []string
	for _, user := range users {
		key := fmt.Sprintf("user:%v:%v", conShortId, user.UserId)
		value, err := json.Marshal(user)
		if err != nil {
			keys = append(keys, key)
			values = append(values, string(value))
		}
	}
	_ = dal.RedisServer.BatchSet(ctx, keys, values, UserInfoExpireTime)
}

func GetUserIdList(ctx context.Context, conShortId int64) ([]int64, error) {
	var userIds []int64
	key := fmt.Sprintf("member:%d", conShortId)
	res, err := dal.RedisServer.ZRange(ctx, key, 0, -1)
	if err == nil && len(res) > 0 {
		for _, val := range res {
			id, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				logrus.Errorf("[GetUserIdList] ParseInt err. err = %v", err)
				break
			}
			userIds = append(userIds, id)
		}
		if len(res) == len(userIds) {
			return userIds, nil
		}
		userIds = []int64{}
	}
	var users []*ConversationUserInfo
	err = dal.MysqlDB.Select("user_id", "create_time").Where("con_short_id=?", conShortId).Find(&users).Error
	if err != nil {
		logrus.Errorf("[GetUserIdList] mysql get user list err. err = %v", err)
		return nil, err
	}
	for _, user := range users {
		userIds = append(userIds, user.Id)
	}
	if len(res) == 0 {
		go AsyncSetUserZSetCache(ctx, key, users)
	}
	return userIds, nil
}

func AsyncSetUserZSetCache(ctx context.Context, key string, users []*ConversationUserInfo) {
	var values []redis.Z
	for i := 0; i < len(users); i++ {
		values = append(values, redis.Z{
			Member: users[i].UserId,
			Score:  float64(users[i].CreateTime.Unix()),
		})
	}
	_ = dal.RedisServer.ZAdd(ctx, key, values)
}

func PackUserModel(userId int64, req *im.AddConversationMembersRequest) *ConversationUserInfo {
	user := &ConversationUserInfo{
		ConShortId: req.GetConShortId(),
		UserId:     userId,
		Privilege:  2,
		Operator:   req.GetOperator(),
	}
	curTime := time.Now()
	user.CreateTime = curTime
	user.ModifyTime = curTime
	return user
}

func PackUserInfo(model *ConversationUserInfo) *im.ConversationUserInfo {
	if model == nil {
		return nil
	}
	user := &im.ConversationUserInfo{
		ConShortId:     util.Int64(model.ConShortId),
		UserId:         util.Int64(model.UserId),
		Privilege:      util.Int32(model.Privilege),
		NickName:       util.String(model.NickName),
		BlockTimeStamp: util.Int64(model.BlockTimeStamp),
		Operator:       util.Int64(model.Operator),
		CreateTime:     util.Int64(model.CreateTime.Unix()),
		ModifyTime:     util.Int64(model.ModifyTime.Unix()),
		Status:         util.Int32(model.Status),
		Extra:          util.String(model.Extra),
	}
	return user
}
