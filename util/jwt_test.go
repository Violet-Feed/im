package util

import (
	"testing"
)

func TestJwt(t *testing.T) {
	userId := int64(1887497212032712704)
	t.Log(userId)
	token, err := GenerateUserToken(userId)
	if err != nil {
		t.Error(err)
	}
	t.Log(token)
	//time.Sleep(1 * time.Second )
	userId, err = ParseUserToken(token)
	if err != nil {
		t.Error(err)
	}
	t.Log(userId)
}
