package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_hashing(t *testing.T){
	password := "my_secure_password"
	hashedPassword, err := HashPassword(password)

	assert.NoError(t, err, "HashPassword() should not return an error")
	assert.True(t, CompareHashAndPassword(hashedPassword, password), "CompareHashAndPassword() should return true for correct password")
}
