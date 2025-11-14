package utils

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func Test_Generate_And_Validate_Token(t *testing.T) {
	//Test GenerateToken and ValidateToken functions correctly
	secret := "my_secret_key"
	expectedUserEmail := "user_12345"
	token, err := GenerateToken(expectedUserEmail, secret)
	assert.NoError(t, err, "Error generating token")
	gotUserEmail, err := ValidateToken(token, secret)
	assert.NoError(t, err, "Error validating token")
	assert.Equal(t, expectedUserEmail, gotUserEmail, "User Email does not match")

	//Test ValidateToken with invalid token
	invalidToken := token + "invalid"
	_, err = ValidateToken(invalidToken, secret)
	assert.Error(t, err, "Expected error for invalid token")

	//Test ValidateToken with empty token
	_, err = ValidateToken("", secret)
	assert.Error(t, err, "Expected error for empty token")

	//Test GenerateToken and ValidateToken with different secret
	token, err = GenerateToken(expectedUserEmail, secret)
	assert.NoError(t, err, "Error generating token")
	_, err = ValidateToken(token, "different_secret")
	assert.Error(t, err, "Expected error for different secret")
}
