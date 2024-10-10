package handler

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

const (
	StateCode_Internal_ERROR = 1001
	StateCode_Param_ERROR    = 1002
)

//type MessageSendRequest struct {
//	ConversationId      *string `json:"conversation_id"`
//	ConversationShortId *int64  `json:"conversation_short_id"`
//	ConversationType    *int32  `json:"conversation_type"`
//	MessageType         *int32  `json:"message_type"`
//	MessageContent      *string `json:"message_content"`
//}
