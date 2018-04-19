package utils

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

var now = time.Now

type User struct {
	ID    string `json:"uid"`
	Email string `json:"email"`
	Pass  string `json:"pass,omitempty"`
	Name  string `json:"name"`
}

type JWTClaims struct {
	User *User `json:"user"`
	jwt.StandardClaims
}

func NewJWT(user *User, secret []byte) (string, error) {
	claims := JWTClaims{
		User: user,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  now().Unix(),
			NotBefore: now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	s, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return s, nil
}

func VerifyJWT(tokenString string, secret []byte) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, nil
	}
}
