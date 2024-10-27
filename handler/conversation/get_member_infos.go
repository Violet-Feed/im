package conversation

import "context"

func GetMemberInfos(ctx context.Context) {
	//convId,userIds
	//获取core信息,群主
	//获取userInfoList，查询redis，600，key:convId:userId,mget，未命中查mysql，写入redis一天
}
