package conversation

import (
	"context"
	"github.com/sirupsen/logrus"
	"im/handler/conversation/model"
	"im/proto_gen/im"
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
		settingInfos = append(settingInfos, model.PackSettingInfo(settingModel[id]))
	}
	//TODO:获取readIndex，readBadge
	resp.SettingInfos = settingInfos
	return resp, nil
	//userId,convIds
	//redis mget key；convId:userId,mysql,5小时
	//补齐，获取core，并发：判断是否成员，设置默认setting
	//获取readIndex、readBadge，abase mget，key；convId:userId
}
