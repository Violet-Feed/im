package push

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/handler"
	"im/proto_gen/im"
)

func Push(ctx context.Context, req *im.PushRequest) (resp *im.PushResponse, err error) {
	resp = &im.PushResponse{}
	message, _ := json.Marshal(req)
	userId := req.GetReceiverId()
	key := fmt.Sprintf("conn:%d", userId)
	conns, err := dal.RedisServer.HGetAll(ctx, key)
	if err != nil {
		logrus.Errorf("[Push] redid HGetAll err. err = %v", err)
		return nil, err
	}
	for connId, connInfo := range conns {
		logrus.Infof("[Push] get connections. connId = %v, connInfo = %v", connId, connInfo)
		connInter, _ := handler.Connections.Load(connId)
		if connInter == nil {
			continue
		}
		if conn, ok := connInter.(*websocket.Conn); ok {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				logrus.Warnf("[Push] WriteMessage err. err = %v", err)
			}
		}
	}
	return resp, nil
}
