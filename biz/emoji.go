package biz

import (
	"context"
	"im/biz/model"
	"im/proto_gen/common"
	"im/proto_gen/im"
	"im/util"
)

func AddEmoji(ctx context.Context, req *im.AddEmojiRequest) (*im.AddEmojiResponse, error) {
	resp := &im.AddEmojiResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	emojiId := util.EmojiIdGenerator.Generate().Int64()
	if err := model.InsertEmojiInfo(ctx, model.PackEmojiModel(emojiId, req)); err != nil {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, err
	}
	resp.EmojiId = emojiId
	return resp, nil
}

func GetEmojiList(ctx context.Context, req *im.GetEmojiListRequest) (*im.GetEmojiListResponse, error) {
	resp := &im.GetEmojiListResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	emojiList, err := model.GetEmojiList(ctx, req.GetUserId(), req.GetPage())
	if err != nil {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, err
	}
	var emojis []*im.EmojiInfo
	for _, emoji := range emojiList {
		emojis = append(emojis, model.PackEmojiInfo(emoji))
	}
	resp.Emojis = emojis
	return resp, nil
}
