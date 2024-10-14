package util

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

const JWT_KEY = "SDFGjhdsfalshdfHFdsjkdsfds126232131afasdfac"
const JWT_TTL = time.Hour * 24 * 14

func GenerateUserToken(userId int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": strconv.FormatInt(userId, 10),
		"exp": time.Now().Add(JWT_TTL).Unix(),
	})
	return token.SignedString([]byte(JWT_KEY))
}

func ParseUserToken(userToken string) (int64, error) {
	token, err := jwt.Parse(userToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWT_KEY), nil
	}, jwt.WithExpirationRequired())
	if err != nil {
		return 0, err
	}
	if !token.Valid {
		return 0, errors.New("invalid token")
	}
	subject, err := token.Claims.GetSubject()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(subject, 10, 64)
}
