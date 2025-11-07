package models

type User struct {
	Id int64 `json:"id"`
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,password"`
	PasswordHash string
}
