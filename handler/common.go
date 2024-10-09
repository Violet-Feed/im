package handler

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var (
	INTERNAL_SERVER_ERROR = Response{Code: 201, Message: "Internal Server Error"}
)
