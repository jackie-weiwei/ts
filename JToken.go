package ts

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JClaims struct {
	jwt.RegisteredClaims
}

type LoginClaims struct {
	AppId string `json:"appId"`
	jwt.RegisteredClaims
}

// const accessTokenExpire = time.Hour * 24 * 7
// const refreshTokenExpire = time.Hour * 24 * 30

func GetToken(secretKey string, accessExpire time.Duration, refreshExpire time.Duration) (string, string, error) {
	accessToken, err := generateToken(accessExpire, secretKey)

	if err != nil {
		return "", "", err
	}

	refreshToken, err := generateToken(refreshExpire, secretKey)

	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func ParseAccessToken(token string, secretKey string) (*JClaims, error) {

	t, err := jwt.ParseWithClaims(token, &JClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	claims, ok := t.Claims.(*JClaims)

	if ok && t.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func VerifyLoginToken(token string, secretKey string, appId string) bool {
	t, err := jwt.ParseWithClaims(token, &LoginClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return false
	}

	claims, ok := t.Claims.(*LoginClaims)

	if ok && t.Valid {
		return claims.AppId == appId
	} else {
		return false
	}
}

func generateToken(expire time.Duration, secretKey string) (string, error) {

	claims := JClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))

	return accessToken, err
}
