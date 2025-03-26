package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const TOKEN_SECRET_TEST = "mylittlesecret"

func TestMakeJWT(t *testing.T) {
	userIdTest := uuid.New()
	token, err := MakeJWT(userIdTest, TOKEN_SECRET_TEST, time.Duration(time.Duration(time.Hour*1)))
	if err != nil {
		t.Fatal("couldn't create token")
	}
	id, err := ValidateJWT(token, TOKEN_SECRET_TEST)
	if err != nil {
		t.Fatal("err validating")
	}

	assert.Equal(t, userIdTest, id)

}
