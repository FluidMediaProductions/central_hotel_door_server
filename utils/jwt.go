package utils

import (
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
)

var JWTSecret = []byte{0xf9, 0x1, 0xd4, 0x9c, 0xc8, 0x55, 0xc6, 0xe9, 0x63, 0x32, 0xc2, 0xcd, 0xa8, 0x6f, 0x98, 0xf7, 0x60, 0xae, 0x5f, 0xb9, 0xd0, 0x5f, 0xf8, 0xe3, 0x5, 0x8f, 0x19, 0x5, 0x96, 0x1, 0x29, 0xe5, 0x83, 0x3a, 0x8e, 0xf4, 0xa, 0x38, 0xa9, 0xd, 0x87, 0xcd, 0x5f, 0x2d, 0x42, 0x78, 0xf9, 0xfd, 0x12, 0x22, 0xaf, 0xae, 0xc6, 0x3e, 0x84, 0xe3, 0x8a, 0xe0, 0xe3, 0x34, 0xbc, 0xc1, 0xbc, 0x1c}

var now = time.Now

type User struct {
	gorm.Model
	Email string `json:"email"`
	Pass  string `json:"-"`
	Name  string `json:"name"`
}

type JWTClaims struct {
	User *User `json:"user"`
	jwt.StandardClaims
}

func NewJWT(user *User) (string, error) {
	claims := JWTClaims{
		User: user,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  now().Unix(),
			NotBefore: now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	s, err := token.SignedString(JWTSecret)
	if err != nil {
		return "", err
	}
	return s, nil
}

func VerifyJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		log.Println(claims.NotBefore, now())
		return claims, nil
	} else {
		return nil, nil
	}
}
