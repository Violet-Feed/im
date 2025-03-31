package handler

import (
	"im/proto_gen/common"
)

type HttpResponse struct {
	Code    common.StatusCode `json:"code"`
	Message string            `json:"message"`
	Data    interface{}       `json:"data"`
}
