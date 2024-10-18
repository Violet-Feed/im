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
	//TODO：入库
	return nil
}
