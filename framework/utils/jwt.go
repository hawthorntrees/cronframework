package utils

import (
	"encoding/base64"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hawthorntrees/cronframework/framework/config"
)

type CustomClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(userID string, username string) (string, error) {
	key := config.GetTokenKey()
	decodeString, _ := base64.StdEncoding.DecodeString(key)

	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(config.GetExpireTime()), // 过期时间
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(decodeString)
}

func ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		key := config.GetTokenKey()
		decodeString, _ := base64.StdEncoding.DecodeString(key)
		return decodeString, nil
	})
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
