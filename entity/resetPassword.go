package entity

import (
	"time"
)

type ResetPassword struct {
	ID        int
	UserId    int
	Token     string
	ExpiredAt time.Time
}

func (ResetPassword) TableName() string {
	return "resetpassword"
}
