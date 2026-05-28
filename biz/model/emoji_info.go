package model

import (
	"context"
	"im/dal"
	"im/proto_gen/im"
	"time"

	"github.com/sirupsen/logrus"
)

type EmojiInfo struct {
	Id         int64     `gorm:"column:id" json:"id"`
	EmojiId    int64     `gorm:"column:emoji_id" json:"emoji_id"`
	EmojiName  string    `gorm:"column:emoji_name" json:"name"`
	EmojiUri   string    `gorm:"column:emoji_uri" json:"avatar_uri"`
	OwnerId    int64     `gorm:"column:owner_id" json:"owner_id"`
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
	ModifyTime time.Time `gorm:"column:modify_time" json:"modify_time"`
	Status     int32     `gorm:"column:status" json:"status"`
	Extra      string    `gorm:"column:extra" json:"extra"`
}

func (c *EmojiInfo) TableName() string {
	return "emoji"
}

func InsertEmojiInfo(ctx context.Context, emoji *EmojiInfo) error {
	if err := dal.MysqlDB.Create(emoji).Error; err != nil {
		logrus.Errorf("[InsertEmojiInfo] mysql insert emoji err. err = %v", err)
		return err
	}
	return nil
}

func GetEmojiList(ctx context.Context, ownerId int64, page int32) ([]*EmojiInfo, error) {
	var emojiList []*EmojiInfo
	const pageSize = 15
	offset := (page - 1) * pageSize
	if err := dal.MysqlDB.Where("status  = ?", 0).Order("modify_time DESC").Limit(pageSize).Offset(offset).Find(&emojiList).Error; err != nil {
		logrus.Errorf("[GetEmojiList] mysql get emoji list err. err = %v", err)
		return nil, err
	}
	return emojiList, nil
}

func PackEmojiModel(emojiId int64, req *im.AddEmojiRequest) *EmojiInfo {
	emoji := &EmojiInfo{
		EmojiId:   emojiId,
		EmojiName: req.GetEmojiName(),
		EmojiUri:  req.GetEmojiUri(),
		OwnerId:   req.GetUserId(),
	}
	curTime := time.Now()
	emoji.CreateTime = curTime
	emoji.ModifyTime = curTime
	return emoji
}

func PackEmojiInfo(model *EmojiInfo) *im.EmojiInfo {
	if model == nil {
		return nil
	}
	emoji := &im.EmojiInfo{
		EmojiId:    model.EmojiId,
		EmojiName:  model.EmojiName,
		EmojiUri:   model.EmojiUri,
		OwnerId:    model.OwnerId,
		CreateTime: model.CreateTime.Unix(),
		ModifyTime: model.ModifyTime.Unix(),
		Status:     model.Status,
		Extra:      model.Extra,
	}
	return emoji
}
