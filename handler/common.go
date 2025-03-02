package handler

import "im/proto_gen/im"

type HttpResponse struct {
	Code    im.StatusCode `json:"code"`
	Message string        `json:"message"`
	Data    interface{}   `json:"data"`
}
