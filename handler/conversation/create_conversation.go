package conversation

import (
	"context"
	"im/proto_gen/im"
	"im/util"
)

func CreateConversation(ctx context.Context, req *im.CreateConversationRequest, resp *im.CreateConversationResponse) error {
	if req.GetConvType() == int32(im.ConversationType_ConversationType_One_Chat) {
		//TODO：鉴权
	}
	shortId := util.ConvIdGenerator.Generate().Int64()
	resp.ConvShortId = util.Int64(shortId)
	//幂等创建Identity->写redis->创建/更新core->写redis->发送命令消息，单聊更新setting，群聊添加成员、审核开关
	//TODO：入库
	return nil
}
