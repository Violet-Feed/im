package mysql

import "time"

type ConversationUserInfo struct {
	Id         int64     `gorm:"column:id" json:"id"`
	ConShortId int64     `gorm:"column:con_short_id" json:"con_short_id"`
	UserId     int64     `gorm:"column:user_id" json:"user_id"`
	Level      int32     `gorm:"column:level" json:"level"`
	NickName   string    `gorm:"column:nick_name" json:"nick_name"`
	BlockTime  time.Time `gorm:"column:blocked" json:"blocked"`
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
	ModifyTime time.Time `gorm:"column:modify_time" json:"modify_time"`
	Status     int32     `gorm:"column:status" json:"status"`
	Extra      string    `gorm:"column:extra" json:"extra"`
}

func (c *ConversationUserInfo) TableName() string {
	return "conversation_user_info"
}
