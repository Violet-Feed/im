package util

import (
	"testing"
)

func TestJwt(t *testing.T) {
	//userId := int64(1887497212032712704)
	//t.Log(userId)
	//token, err := GenerateUserToken(userId)
	//if err != nil {
	//	t.Error(err)
	//}
	token := "eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiI5MzAwODE4N2VlNTQ0MTk5ODJkMGIyNDViNGNlOTQzYyIsInN1YiI6IjE4ODc0OTcyMTIwMzI3MTI3MDQiLCJpc3MiOiJzZyIsImlhdCI6MTc0MDQ3MDcwMSwiZXhwIjoxNzQxNjgwMzAxfQ.hG89-qRETdi8h20sHe5pt9rTXOIu2S1GBvxXoGh7o44"
	t.Log(token)
	userId, err := ParseUserToken(token)
	if err != nil {
		t.Error(err)
	}
	t.Log(userId)
}
