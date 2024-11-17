package model

type Todo struct {
	Id          int64
	Name        string
	Description string
}

type Session struct {
	Id        string
	UserId    int64
	ExpiresAt int64
}

type User struct {
	Id           int64
	Username     string
	PasswordHash string
}
