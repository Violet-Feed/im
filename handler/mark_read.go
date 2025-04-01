package handler

import (
	"context"
	"im/proto_gen/common"
	"im/proto_gen/im"
)

func MarkRead(ctx context.Context, req *im.MarkReadRequest) (resp *im.MarkReadResponse) {
	resp = &im.MarkReadResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	return
	//设置总、各会话未读数(im_counter_manager_rust)->设置已读index和count(im_conversation_api,redis hash,abase xset)->
	//发送命令消息->更新最近会话(recent)->lastindex,保存,写命令链,push->发送已读消息(im_lastuser_api)
}
