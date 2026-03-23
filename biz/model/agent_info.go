package model

import (
	"context"
	"im/dal"
	"im/proto_gen/im"
	"time"

	"github.com/sirupsen/logrus"
)

type ConversationAgentInfo struct {
	Id         int64     `gorm:"column:id" json:"id"`
	ConShortId int64     `gorm:"column:con_short_id" json:"con_short_id"`
	AgentId    int64     `gorm:"column:agent_id" json:"agent_id"`
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
	ModifyTime time.Time `gorm:"column:modify_time" json:"modify_time"`
	Status     int32     `gorm:"column:status" json:"status"`
	Extra      string    `gorm:"column:extra" json:"extra"`
}

func (c *ConversationAgentInfo) TableName() string {
	return "conversation_agent_info"
}

func InsertAgentInfos(ctx context.Context, agents []*ConversationAgentInfo) error {
	tx := dal.MysqlDB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logrus.Errorf("[InsertAgentInfos] panic recovered: %v", r)
		}
	}()
	for i := range agents {
		if err := tx.Create(agents[i]).Error; err != nil {
			tx.Rollback()
			logrus.Errorf("[InsertAgentInfos] mysql insert agent err. err = %v", err)
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		logrus.Errorf("[InsertAgentInfos] mysql commit err. err = %v", err)
		return err
	}
	return nil
}

func GetAgentInfos(ctx context.Context, conShortId int64) ([]*ConversationAgentInfo, error) {
	var agentInfos []*ConversationAgentInfo
	err := dal.MysqlDB.Where("con_short_id = ?", conShortId).Find(&agentInfos).Error
	if err != nil {
		logrus.Errorf("[GetAgentInfos] mysql get agent infos err. err = %v", err)
		return nil, err
	}
	return agentInfos, nil
}

func GetAgentInfosByIds(ctx context.Context, conShortId int64, agentIds []int64) (map[int64]ConversationAgentInfo, error) {
	var agentInfo []ConversationAgentInfo
	err := dal.MysqlDB.Where("con_short_id = (?) and agent_id in (?)", conShortId, agentIds).Find(&agentInfo).Error
	if err != nil {
		logrus.Errorf("[GetAgentInfosByIds] mysql get agent infos err. err = %v", err)
		return nil, err
	}
	agentInfoMap := make(map[int64]ConversationAgentInfo)
	for _, a := range agentInfo {
		agentInfoMap[a.AgentId] = a
	}
	return agentInfoMap, nil
}

func PackAgentModel(agentId int64, req *im.AddConversationAgentsRequest) *ConversationAgentInfo {
	agent := &ConversationAgentInfo{
		ConShortId: req.ConShortId,
		AgentId:    agentId,
	}
	now := time.Now()
	agent.CreateTime = now
	agent.ModifyTime = now
	return agent
}

func PackAgentInfo(model *ConversationAgentInfo) *im.ConversationAgentInfo {
	return &im.ConversationAgentInfo{
		ConShortId: model.ConShortId,
		AgentId:    model.AgentId,
		CreateTime: model.CreateTime.Unix(),
		ModifyTime: model.ModifyTime.Unix(),
		Status:     model.Status,
		Extra:      model.Extra,
	}
}
