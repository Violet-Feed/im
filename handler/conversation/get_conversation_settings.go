package conversation

import (
	"context"
	"github.com/sirupsen/logrus"
	"im/handler/conversation/model"
	"im/proto_gen/im"
	"im/util"
)

func GetConversationSettings(ctx context.Context, req *im.GetConversationSettingsRequest) (resp *im.GetConversationSettingsResponse, err error) {
	resp = &im.GetConversationSettingsResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	settingModel, err := model.GetSettingInfo(ctx, req.GetUserId(), req.GetConShortIds())
	if err != nil {
		logrus.Errorf("[GetConversationSettings] GetSettingInfo err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	var settingInfos []*im.ConversationSettingInfo
	for _, id := range req.GetConShortIds() {
		if settingModel[id] == nil {
			//TODO:补齐
			settingModel[id], err = model.FixSettingModel(ctx, req.GetUserId(), id)
			if err != nil {
				logrus.Errorf("[GetConversationSettings] FixSettingModel err. err = %v", err)
			} else {
				err = model.InsertSettingInfo(ctx, settingModel[id])
				if err != nil {
					logrus.Errorf("[GetConversationSettings] InsertSettingInfo err. err = %v", err)
				}
			}
		}
		settingInfos = append(settingInfos, model.PackSettingInfo(settingModel[id]))
	}
	//TODO:获取readIndex，readBadge
	readIndexStart, err := model.GetReadIndexStart(ctx, req.GetConShortIds(), req.GetUserId())
	if err != nil {
		logrus.Errorf("[GetConversationSettings] GetReadIndexStart err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	readIndexEnd, err := model.GetReadIndexEnd(ctx, req.GetConShortIds(), req.GetUserId())
	if err != nil {
		logrus.Errorf("[GetConversationSettings] GetReadIndexEnd err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	readBadge, err := model.GetReadBadge(ctx, req.GetConShortIds(), req.GetUserId())
	if err != nil {
		logrus.Errorf("[GetConversationSettings] GetReadBadge err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	for i, conShortId := range req.GetConShortIds() {
		settingInfos[i].MinIndex = util.Int64(readIndexStart[conShortId])
		settingInfos[i].ReadIndexEnd = util.Int64(readIndexEnd[conShortId])
		settingInfos[i].ReadBadgeCount = util.Int64(readBadge[conShortId])
	}
	resp.SettingInfos = settingInfos
	return resp, nil
	//userId,convIds
	//redis mget key；convId:userId,mysql,5小时
	//补齐，获取core，并发：判断是否成员，设置默认setting
	//获取readIndex、readBadge，abase mget，key；convId:userId
}
