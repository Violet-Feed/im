package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/sirupsen/logrus"
	"im/biz/constant"
	"im/dal"
	"im/dal/mq"
	"im/proto_gen/common"
	"im/proto_gen/im"
	"im/util"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	ConvLimit = 50
	CmdLimit  = 200
	MsgLimit  = 5
)

func checkMessageSendRequest(req *im.SendMessageRequest) bool {
	if req.GetConId() == "" {
		return false
	}
	if req.GetConType() < 1 || req.GetConType() > 5 {
		return false
	}
	if req.GetMsgType() < 1 || req.GetMsgType() > 5 && req.GetMsgType() != 101 && req.GetMsgType() != 102 {
		return false
	}
	if req.GetMsgContent() == "" {
		return false
	}
	return true
}

func SendMessage(ctx context.Context, req *im.SendMessageRequest) (resp *im.SendMessageResponse, err error) {
	resp = &im.SendMessageResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	if !checkMessageSendRequest(req) {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Param_Error, StatusMessage: "invalid param"}
		return resp, errors.New("param error")
	}
	messageId := util.MsgIdGenerator.Generate().Int64()
	//是否群成员
	var isMember int32
	if req.GetConType() == int32(im.ConversationType_One_Chat) || req.GetConType() == int32(im.ConversationType_AI_Chat) {
		isMember = IsSingleMember(ctx, req.GetConId(), req.GetUserId())
	} else if req.GetConType() == int32(im.ConversationType_Group_Chat) {
		if req.GetUserId() == int64(common.SpecialUser_Conversation) {
			isMember = 1
		} else {
			status, err := IsGroupsMember(ctx, []int64{req.GetConShortId()}, req.GetUserId())
			if err != nil {
				logrus.Errorf("[SendMessage] IsConversationMembers err. err = %v", err)
				resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
				return resp, err
			}
			isMember = status[req.GetConShortId()]
		}
	}
	if isMember != 1 {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Not_Found_Error, StatusMessage: "conversation not member"}
		return resp, errors.New("not conversation member")
	}
	if req.GetConShortId() == 0 && (req.GetConType() == int32(im.ConversationType_One_Chat) || req.GetConType() == int32(im.ConversationType_AI_Chat)) { //创建会话
		parts := strings.Split(req.GetConId(), ":")
		minId, _ := strconv.ParseInt(parts[0], 10, 64)
		maxId, _ := strconv.ParseInt(parts[1], 10, 64)
		createConversationRequest := &im.CreateConversationRequest{
			ConId:   req.GetConId(),
			ConType: req.GetConType(),
			Members: []int64{minId, maxId},
		}
		createConversationResponse, err := CreateConversation(ctx, createConversationRequest)
		if err != nil {
			logrus.Errorf("[SendMessage] CreateConversation err. err = %v", err)
			resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
			return resp, err
		}
		req.ConShortId = createConversationResponse.GetConCoreInfo().GetConShortId()
	}
	//TODO：消息频率控制
	createTime := time.Now().Unix()
	messageBody := &im.MessageBody{
		UserId:      req.GetUserId(),
		ConId:       req.GetConId(),
		ConShortId:  req.GetConShortId(),
		ConType:     req.GetConType(),
		ClientMsgId: req.GetClientMsgId(),
		MsgId:       messageId,
		MsgType:     req.GetMsgType(),
		MsgContent:  req.GetMsgContent(),
		CreateTime:  createTime,
		Extra:       "",
	}
	if req.GetConType() == int32(im.ConversationType_AI_Chat) && req.GetUserId() != int64(common.SpecialUser_AI) {
		err := CallModel(ctx, messageBody)
		if err != nil {
			logrus.Errorf("[SendMessage] CallModel err. err = %v", err)
			resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
			return resp, err
		}
	}
	messageEvent := &im.MessageEvent{
		MsgBody: messageBody,
	}
	logrus.Infof("[SendMessage] messageEvent=%v", messageEvent)
	err = mq.SendToMq(ctx, constant.IM_CONV_TOPIC, strconv.FormatInt(req.GetConShortId(), 10), messageEvent)
	if err != nil {
		logrus.Errorf("[SendMessage] SendToMq err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	resp.MsgId = messageId
	return resp, nil
}

func GetMessages(ctx context.Context, conShortId int64, msgIds []int64) ([]*im.MessageBody, error) {
	messageBodies := make([]*im.MessageBody, 0)
	if len(msgIds) == 0 {
		return messageBodies, nil
	}
	keys := make([]string, 0)
	for _, id := range msgIds {
		keys = append(keys, fmt.Sprintf("msg:%d:%d", conShortId, id))
	}
	messages, err := dal.KvrocksServer.MGet(ctx, keys)
	if err != nil {
		logrus.Errorf("[GetMessages] kvrocks mget err. err = %v", err)
		return nil, err
	}
	for _, message := range messages {
		var messageBody im.MessageBody
		_ = json.Unmarshal([]byte(message), &messageBody)
		messageBodies = append(messageBodies, &messageBody)
	}
	return messageBodies, nil
}

func GetCommands(ctx context.Context, userId int64, msgIds []int64) ([]*im.MessageBody, error) {
	messageBodies := make([]*im.MessageBody, 0)
	if len(msgIds) == 0 {
		return messageBodies, nil
	}
	keys := make([]string, 0)
	for _, id := range msgIds {
		keys = append(keys, fmt.Sprintf("cmd:%d:%d", userId, id))
	}
	messages, err := dal.KvrocksServer.MGet(ctx, keys)
	if err != nil {
		logrus.Errorf("[GetCommands] kvrocks mget err. err = %v", err)
		return nil, err
	}
	for _, message := range messages {
		var messageBody im.MessageBody
		_ = json.Unmarshal([]byte(message), &messageBody)
		messageBodies = append(messageBodies, &messageBody)
	}
	return messageBodies, nil
}

func StoreMessage(ctx context.Context, msgBody *im.MessageBody) error {
	var key string
	if msgBody.MsgType > 100 {
		key = fmt.Sprintf("cmd:%d:%d", msgBody.GetUserId(), msgBody.GetMsgId())
	} else {
		key = fmt.Sprintf("msg:%d:%d", msgBody.GetConShortId(), msgBody.GetMsgId())
	}
	messageBody, err := json.Marshal(msgBody)
	if err != nil {
		logrus.Errorf("[StoreMessage] marshal messageBody err. err = %v", err)
		return err
	}
	err = dal.KvrocksServer.Set(ctx, key, string(messageBody))
	if err != nil {
		logrus.Errorf("[StoreMessage] kvrocks set err. err = %v", err)
		return err
	}
	return nil
}

func GetMessageByUser(ctx context.Context, req *im.GetMessageByUserRequest) (resp *im.GetMessageByUserResponse, err error) {
	resp = &im.GetMessageByUserResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	limit := req.GetLimit()
	if limit > ConvLimit {
		limit = ConvLimit
	}
	//拉取用户会话链
	conShortIds, userConIndexes, err := PullUserConIndex(ctx, req.GetUserId(), req.GetUserConIndex(), limit+1)
	if err != nil {
		logrus.Errorf("[GetMessageByUser] PullUserConIndex err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	if int64(len(conShortIds)) == limit+1 {
		resp.HasMore = true
		conShortIds = conShortIds[:len(conShortIds)-1]
		userConIndexes = userConIndexes[:len(userConIndexes)-1]
	} else {
		resp.HasMore = false
	}
	if len(userConIndexes) == 0 {
		resp.UserConIndex = 0
	} else {
		resp.UserConIndex = userConIndexes[len(userConIndexes)-1]
	}
	var globalErr error
	msgBodiesMapChan, coresMapChan, settingsMapChan := make(chan map[int64][]*im.MessageBody), make(chan map[int64]*im.ConversationCoreInfo), make(chan map[int64]*im.ConversationSettingInfo)
	statusMapChan, badgesMapChan := make(chan map[int64]int32), make(chan map[int64]int64)
	//拉取会话链
	go func() {
		msgBodiesChan := make(chan []*im.MessageBody, len(conShortIds))
		for _, convShortId := range conShortIds {
			go func(convShortId int64) {
				msgIds, conIndexes, err := PullConversationIndex(ctx, convShortId, math.MaxInt64, MsgLimit)
				if err != nil {
					logrus.Errorf("[GetMessageByUser] PullConversationIndex err. err = %v", err)
					msgBodiesChan <- nil
					return
				}
				//TODO:过滤撤回消息
				msgBodies, err := GetMessages(ctx, convShortId, msgIds)
				if err != nil {
					logrus.Errorf("[GetMessageByUser] GetMessage err. err = %v", err)
					msgBodiesChan <- nil
					return
				}
				for i, msgBody := range msgBodies {
					msgBody.ConIndex = conIndexes[i]
				}
				msgBodiesChan <- msgBodies
			}(convShortId)
		}
		msgBodiesMap := make(map[int64][]*im.MessageBody)
		for i := 0; i < len(conShortIds); i++ {
			msgBodies := <-msgBodiesChan
			if msgBodies != nil {
				msgBodiesMap[msgBodies[0].GetConShortId()] = msgBodies
			}
		}
		msgBodiesMapChan <- msgBodiesMap
	}()
	//获取会话core
	go func() {
		cores, err := GetConversationCores(ctx, conShortIds, true)
		if err != nil {
			logrus.Errorf("[GetMessageByUser] GetConversationCores err. err = %v", err)
			coresMapChan <- nil
			statusMapChan <- nil
			globalErr = err
			return
		}
		coresMap := make(map[int64]*im.ConversationCoreInfo)
		for _, core := range cores {
			coresMap[core.GetConShortId()] = core
		}
		coresMapChan <- coresMap
		//获取是否是会话member
		go func() {
			statusMap := make(map[int64]int32)
			groupIds := make([]int64, 0)
			for _, conShortId := range conShortIds {
				core := coresMap[conShortId]
				if core.GetConType() == int32(im.ConversationType_One_Chat) || core.GetConType() == int32(im.ConversationType_AI_Chat) {
					statusMap[conShortId] = IsSingleMember(ctx, core.GetConId(), req.GetUserId())
				} else if core.GetConType() == int32(im.ConversationType_Group_Chat) {
					groupIds = append(groupIds, conShortId)
				}
			}
			status, err := IsGroupsMember(ctx, groupIds, req.GetUserId())
			if err != nil {
				logrus.Errorf("[GetMessageByUser] IsConversationMembers err. err = %v", err)
				statusMapChan <- statusMap
				return
			}
			for k, v := range status {
				statusMap[k] = v
			}
			statusMapChan <- statusMap
		}()
		//TODO:获取最近member，id昵称权限block
		go func() {
			_, err := GetConversationMemberInfos(ctx, 0, []int64{0})
			if err != nil {
				logrus.Errorf("[GetMessageByUser] GetConversationMemberInfos err. err = %v", err)
				return
			}
		}()
	}()
	//获取会话setting
	go func() {
		settings, err := GetConversationSettings(ctx, req.GetUserId(), conShortIds)
		if err != nil {
			logrus.Errorf("[GetMessageByUser] GetConversationSettings err. err = %v", err)
			settingsMapChan <- nil
			globalErr = err
			return
		}
		settingsMap := make(map[int64]*im.ConversationSettingInfo)
		for _, setting := range settings {
			settingsMap[setting.GetConShortId()] = setting
		}
		settingsMapChan <- settingsMap
	}()
	//获取会话badge
	go func() {
		badges, err := GetConversationBadges(ctx, req.GetUserId(), conShortIds)
		if err != nil {
			logrus.Errorf("[GetMessageByUser] GetConversationBadge err. err = %v", err)
			badgesMapChan <- nil
			return
		}
		badgesMap := make(map[int64]int64)
		for i, conShortId := range conShortIds {
			badgesMap[conShortId] = badges[i]
		}
		badgesMapChan <- badgesMap
	}()
	msgBodiesMap := <-msgBodiesMapChan
	coresMap := <-coresMapChan
	settingsMap := <-settingsMapChan
	statusMap := <-statusMapChan
	badgesMap := <-badgesMapChan
	if globalErr != nil {
		logrus.Errorf("[GetMessageByUser] happend err.")
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: globalErr.Error()}
		return resp, globalErr
	}
	conMessages := make([]*im.ConversationMessage, 0)
	//TODO:通过minIndex过滤消息,过滤非成员，无core、setting
	for i, conShortId := range conShortIds {
		core := coresMap[conShortId]
		conInfo := &im.ConversationInfo{
			ConShortId:     core.GetConShortId(),
			ConId:          core.GetConId(),
			ConType:        core.GetConType(),
			UserConIndex:   userConIndexes[i],
			BadgeCount:     badgesMap[conShortId],
			IsMember:       statusMap[conShortId] == 1,
			ConCoreInfo:    core,
			ConSettingInfo: settingsMap[conShortId],
		}
		conMessage := &im.ConversationMessage{
			ConInfo:   conInfo,
			MsgBodies: msgBodiesMap[conShortId],
		}
		conMessages = append(conMessages, conMessage)
	}
	resp.Cons = conMessages
	return resp, nil
	//获取最近会话id(recent_conversation,abase,zset)->拉取会话链(inbox_api,V2)->获取消息内容(message_api)->隐藏撤回消息
	//->获取会话core,setting信息(im_conversation_api)->获取会话ext信息(conversation_ext)->获取群聊是否是成员,最近成员信息(im_conversation_api)
	//->获取消息总数(im_counter_manager_rust,mget)->(第一个请求：获取命令链最近index(inbox_api,V2)->获取最近会话version(recent_conversation))
	//->信息整理过滤
}

func GetCommandByUser(ctx context.Context, req *im.GetCommandByUserRequest) (resp *im.GetCommandByUserResponse, err error) {
	resp = &im.GetCommandByUserResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	userCmdIndex, limit := req.GetUserCmdIndex(), req.GetLimit()
	if limit > CmdLimit {
		limit = CmdLimit
	}
	if userCmdIndex == 0 {
		userCmdIndex = 1
	}
	msgIds, userCmdIndexes, err := PullUserCmdIndex(ctx, req.GetUserId(), userCmdIndex, limit+1)
	if err != nil {
		logrus.Errorf("[GetCommandByUser] PullUserCmdIndex err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	if int64(len(msgIds)) == limit+1 {
		resp.HasMore = true
		msgIds = msgIds[:len(msgIds)-1]
		userCmdIndexes = userCmdIndexes[:len(userCmdIndexes)-1]
	} else {
		resp.HasMore = false
	}
	if len(userCmdIndexes) == 0 {
		resp.UserCmdIndex = 0
	} else {
		resp.UserCmdIndex = userCmdIndexes[len(userCmdIndexes)-1]
	}
	msgBodies, err := GetCommands(ctx, req.GetUserId(), msgIds)
	if err != nil {
		logrus.Errorf("[GetCommandByUser] GetCommands err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	resp.MsgBodies = msgBodies
	return resp, nil
}

func GetMessageByConversation(ctx context.Context, req *im.GetMessageByConversationRequest) (resp *im.GetMessageByConversationResponse, err error) {
	resp = &im.GetMessageByConversationResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	cores, err := GetConversationCores(ctx, []int64{req.GetConShortId()}, false)
	if err != nil {
		logrus.Errorf("[GetMessageByConversation] GetConversationCores err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	if len(cores) == 0 || cores[0] == nil || cores[0].GetStatus() != 0 {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Not_Found_Error, StatusMessage: "conversation not found"}
		return resp, errors.New("conversation not found")
	}
	//是否群成员
	var isMember int32
	if cores[0].GetConType() == int32(im.ConversationType_One_Chat) || cores[0].GetConType() == int32(im.ConversationType_AI_Chat) {
		isMember = IsSingleMember(ctx, cores[0].GetConId(), req.GetUserId())
	} else if cores[0].GetConType() == int32(im.ConversationType_Group_Chat) {
		status, err := IsGroupsMember(ctx, []int64{req.GetConShortId()}, req.GetUserId())
		if err != nil {
			logrus.Errorf("[GetMessageByConversation] IsGroupsMember err. err = %v", err)
			resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
			return resp, err
		}
		isMember = status[req.GetConShortId()]
	}
	if isMember != 1 {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Not_Found_Error, StatusMessage: "conversation not member"}
		return resp, errors.New("not conversation member")
	}
	//拉取会话链
	msgIds, conIndexes, err := PullConversationIndex(ctx, req.GetConShortId(), req.GetConIndex(), req.GetLimit())
	if err != nil {
		logrus.Errorf("[GetMessageByConversation] PullConversationIndex err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	msgBodies, err := GetMessages(ctx, req.GetConShortId(), msgIds)
	if err != nil {
		logrus.Errorf("[GetMessageByConversation] GetMessages err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	for i, msgBody := range msgBodies {
		msgBody.ConIndex = conIndexes[i]
	}
	resp.MsgBodies = msgBodies
	return resp, nil
	//获取core信息(mysql)->获取成员数量(redis+mysql)->判断是否为成员(redis,mysql)->拉取会话链(loadmore)->隐藏撤回消息
}

func CallModel(ctx context.Context, message *im.MessageBody) (err error) {
	msgIds, _, err := PullConversationIndex(ctx, message.GetConShortId(), math.MaxInt64, 10)
	if err != nil {
		logrus.Errorf("[CallModel] PullConversationIndex err. err = %v", err)
		return err
	}
	go func() {
		msgBodies, err := GetMessages(ctx, message.GetConShortId(), msgIds)
		if err != nil {
			logrus.Errorf("[CallModel] GetMessages err. err = %v", err)
			return
		}
		model := "qwen"
		var content string
		if model == "qwen" {
			content, err = CallQwenModel(ctx, message, msgBodies)
			if err != nil {
				logrus.Errorf("[CallModel] CallQwenModel err. err = %v", err)
				return
			}
		} else {
			content, err = CallRagModel(ctx, message, msgBodies)
			if err != nil {
				logrus.Errorf("[CallModel] CallRagModel err. err = %v", err)
				return
			}
		}
		logrus.Infof("[CallModel] call model success. model = %v, content = %v", model, content)
		return
		_, err = SendMessage(ctx, &im.SendMessageRequest{
			UserId:     int64(common.SpecialUser_AI),
			ConShortId: message.GetConShortId(),
			ConId:      message.GetConId(),
			ConType:    message.GetConType(),
			MsgType:    int32(im.MessageType_Text),
			MsgContent: content,
		})
		if err != nil {
			logrus.Errorf("[CallModel] SendMessage err. err = %v", err)
			return
		}
	}()
	return nil
}

func CallQwenModel(ctx context.Context, message *im.MessageBody, messages []*im.MessageBody) (content string, err error) {
	msgParams := make([]openai.ChatCompletionMessageParamUnion, 0)
	msgParams = append(msgParams, openai.UserMessage(message.GetMsgContent()))
	for _, msgBody := range messages {
		if msgBody.GetUserId() == int64(common.SpecialUser_AI) {
			msgParams = append(msgParams, openai.AssistantMessage(msgBody.GetMsgContent()))
		} else {
			msgParams = append(msgParams, openai.UserMessage(msgBody.GetMsgContent()))
		}
	}
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("DASHSCOPE_API_KEY")),
		option.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1/"),
	)
	chatCompletion, err := client.Chat.Completions.New(
		context.TODO(), openai.ChatCompletionNewParams{
			Messages: openai.F(msgParams),
			Model:    openai.F("qwen2.5-1.5b-instruct"),
		},
	)
	if err != nil {
		logrus.Errorf("[CallQwenModel] call model err. err = %v", err)
		return "", err
	}
	return chatCompletion.Choices[0].Message.Content, nil
}

func CallRagModel(ctx context.Context, message *im.MessageBody, messages []*im.MessageBody) (content string, err error) {
	url := "http://115.190.94.13:8000/query"
	history := make([]*im.RagModelMessage, 0)
	for _, msgBody := range messages {
		if msgBody.GetUserId() == int64(common.SpecialUser_AI) {
			history = append(history, &im.RagModelMessage{
				Content: msgBody.MsgContent,
				Role:    "assistant",
			})
		} else {
			history = append(history, &im.RagModelMessage{
				Content: msgBody.MsgContent,
				Role:    "user",
			})
		}
	}
	reqBody, _ := json.Marshal(&im.RagModelRequest{
		Query:          message.MsgContent,
		ConversationId: strconv.FormatInt(message.GetConShortId(), 10),
		History:        history,
		Source:         "violet",
	})
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(reqBody)))
	if err != nil {
		logrus.Errorf("[CallRagModel] http.NewRequest err. err = %v", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("[CallRagModel] http client do err. err = %v", err)
		return "", err
	}
	defer resp.Body.Close()
	var respBody bytes.Buffer
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			respBody.Write(buffer[:n])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			logrus.Errorf("[CallRagModel] resp body read err. err = %v", err)
			return "", err
		}
	}
	return respBody.String(), nil
}
