package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)


func GenerateToken(email string, userId int64, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":      email,
		"userId":     userId,
		"exp":        time.Now().Add(2 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(secretKey))
}

func ValidateToken(tokenString string, secretKey string) (int64, error) {
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return 0, err
	}
	if !parsedToken.Valid {
		return 0, errors.New("invalid token")
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}
	userId := int64(claims["userId"].(float64))
	return userId, nil
}
