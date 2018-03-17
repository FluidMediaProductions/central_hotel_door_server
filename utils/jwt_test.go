package utils

import (
	"github.com/dgrijalva/jwt-go"
	"testing"
	"time"
)

func TestJWT(t *testing.T) {
	user := &User{
		Name: "Bob",
	}

	wayback := time.Date(1971, time.January, 1, 0, 0, 0, 0, time.UTC)
	jwt.TimeFunc = func() time.Time { return wayback }
	now = jwt.TimeFunc

	token, err := NewJWT(user)
	if err != nil {
		t.Fatalf("Got error whilst making JWT: %v", err)
	}

	claims, err := VerifyJWT(token)
	if err != nil {
		t.Fatalf("Got error verifying JWT: %v", err)
	}

	if claims.User.Name != user.Name {
		t.Errorf("User in JWT does not match, expected name %s got %s", user.Name, claims.User.Name)
	}

	beforeWayback := time.Date(1969, time.January, 1, 0, 0, 0, 0, time.UTC)
	jwt.TimeFunc = func() time.Time { return beforeWayback }
	now = jwt.TimeFunc

	claims, err = VerifyJWT(token)
	if err == nil {
		t.Fatalf("Did not get error verifying invalid JWT")
	}

	jwt.TimeFunc = time.Now
	now = time.Now
}
