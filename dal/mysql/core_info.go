package mysql

import (
	"github.com/sirupsen/logrus"
	"time"
)

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
}

func (c *ConversationCoreInfo) TableName() string {
	return "conversation_core_info"
}

func InsertCoreInfo(core *ConversationCoreInfo) error {
	err := DB.Create(core).Error
	if err != nil {
		logrus.Errorf("mysql insert core err. err = %v", err)
		return err
	}
	return nil
}
