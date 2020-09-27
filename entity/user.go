package entity

import "gopkg.in/guregu/null.v4"

type User struct {
	ID            int         `db:"id" json:"id"`
	Email         null.String `db:"email" json:"email"`
	Username      string      `db:"username" json:"username"`
	Nick          string      `db:"nick" json:"nick"`
	Avatar        string      `db:"avatar" json:"avatar"`
	Password      string      `db:"password" json:"-"`
	PassSalt      string      `db:"passsalt" json:"-"`
	Submit        int         `db:"submit" json:"submit"`
	Solved        int         `db:"solved" json:"solved"`
	Defunct       int         `db:"defunct" json:"defunct"`
	CreatedAt     string      `db:"created_at" json:"created_at"`
	UpdatedAt     string      `db:"updated_at" json:"updated_at"`
	IsCompeteUser int         `db:"is_compete_user" json:"is_compete_user"`
	RoleId        int         `db:"role_id" json:"role_id"`
}

func (User) TableName() string {
	return "user"
}
