package model

type Todo struct {
	Id          int64
	Name        string
	UserId      int64
	Description string
	IsComplete  bool
}

type Session struct {
	Id        string
	UserId    int64
	ExpiresAt int64
}

type User struct {
	Id            int64
	Email         string
	EmailVerified bool
}
