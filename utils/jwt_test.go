package utils

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var JWTSecret = []byte{0xf9, 0x1, 0xd4, 0x9c, 0xc8, 0x55, 0xc6, 0xe9, 0x63, 0x32, 0xc2, 0xcd, 0xa8, 0x6f, 0x98, 0xf7, 0x60, 0xae, 0x5f, 0xb9, 0xd0, 0x5f, 0xf8, 0xe3, 0x5, 0x8f, 0x19, 0x5, 0x96, 0x1, 0x29, 0xe5, 0x83, 0x3a, 0x8e, 0xf4, 0xa, 0x38, 0xa9, 0xd, 0x87, 0xcd, 0x5f, 0x2d, 0x42, 0x78, 0xf9, 0xfd, 0x12, 0x22, 0xaf, 0xae, 0xc6, 0x3e, 0x84, 0xe3, 0x8a, 0xe0, 0xe3, 0x34, 0xbc, 0xc1, 0xbc, 0x1c}

func TestJWT(t *testing.T) {
	user := &User{
		Name: "Bob",
	}

	wayback := time.Date(1971, time.January, 1, 0, 0, 0, 0, time.UTC)
	jwt.TimeFunc = func() time.Time { return wayback }
	now = jwt.TimeFunc

	token, err := NewJWT(user, JWTSecret)
	if err != nil {
		t.Fatalf("Got error whilst making JWT: %v", err)
	}

	claims, err := VerifyJWT(token, JWTSecret)
	if err != nil {
		t.Fatalf("Got error verifying JWT: %v", err)
	}

	if claims.User.Name != user.Name {
		t.Errorf("User in JWT does not match, expected name %s got %s", user.Name, claims.User.Name)
	}

	beforeWayback := time.Date(1969, time.January, 1, 0, 0, 0, 0, time.UTC)
	jwt.TimeFunc = func() time.Time { return beforeWayback }
	now = jwt.TimeFunc

	claims, err = VerifyJWT(token, JWTSecret)
	if err == nil {
		t.Fatalf("Did not get error verifying invalid JWT")
	}

	jwt.TimeFunc = time.Now
	now = time.Now
}
