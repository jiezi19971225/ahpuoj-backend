package entity

import (
	"gopkg.in/guregu/null.v4"
	"time"
)

type User struct {
	ID            int         `json:"id"`
	Email         null.String `json:"email"`
	Username      string      `json:"username"`
	Nick          string      `json:"nick"`
	Avatar        string      `json:"avatar"`
	Password      string      `json:"-"`
	Passsalt      string      `json:"-"`
	Submit        int         `json:"submit"`
	Solved        int         `json:"solved"`
	Defunct       int         `json:"defunct"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	IsCompeteUser int         `json:"is_compete_user"`
	RoleId        int         `json:"role_id"`
}

func (User) TableName() string {
	return "user"
}
