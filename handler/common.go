package handler

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

const (
	StatusCode_Success       = 1000
	StateCode_Internal_ERROR = 1001
	StateCode_Param_ERROR    = 1002
)
