package model

import (
	"ahpuoj/config"
	"ahpuoj/utils"
	"errors"
)

type User1 struct {
	ID            int        `db:"id" json:"id"`
	Email         NullString `db:"email" json:"email"`
	Username      string     `db:"username" json:"username"`
	Nick          string     `db:"nick" json:"nick"`
	Avatar        string     `db:"avatar" json:"avatar"`
	Password      string     `db:"password" json:"-"`
	PassSalt      string     `db:"passsalt" json:"-"`
	Submit        int        `db:"submit" json:"submit"`
	Solved        int        `db:"solved" json:"solved"`
	Defunct       int        `db:"defunct" json:"defunct"`
	CreatedAt     string     `db:"created_at" json:"created_at"`
	UpdatedAt     string     `db:"updated_at" json:"updated_at"`
	IsCompeteUser int        `db:"is_compete_user" json:"is_compete_user"`
	RoleId        int        `db:"role_id" json:"role_id"`
}

func (User1) TableName() string {
	return "user"
}

type User struct {
	Id            int        `db:"id" json:"id"`
	Email         NullString `db:"email" json:"email"`
	Username      string     `db:"username" json:"username"`
	Nick          string     `db:"nick" json:"nick"`
	Avatar        string     `db:"avatar" json:"avatar"`
	Password      string     `db:"password" json:"-"`
	PassSalt      string     `db:"passsalt" json:"-"`
	Submit        int        `db:"submit" json:"submit"`
	Solved        int        `db:"solved" json:"solved"`
	Defunct       int        `db:"defunct" json:"defunct"`
	CreatedAt     string     `db:"created_at" json:"created_at"`
	UpdatedAt     string     `db:"updated_at" json:"updated_at"`
	IsCompeteUser int        `db:"is_compete_user" json:"is_compete_user"`
	RoleId        int        `db:"role_id" json:"role_id"`
	Role          string     `json:"role"`
}

func (user *User) Save() error {
	defaultAvatar, _ := config.Conf.GetValue("preset", "avatar")
	result, err := DB.Exec(`insert into user
	(email,username,password,passsalt,nick,avatar,submit,solved,defunct,is_compete_user,created_at,updated_at) 
	values (?,?,?,?,?,?,0,0,0,?,NOW(),NOW())`, user.Email, user.Username, user.Password, user.PassSalt, user.Nick, defaultAvatar, user.IsCompeteUser)
	if err != nil {
		return err
	}
	lastInsertId, _ := result.LastInsertId()
	user.Id = utils.Int64to32(lastInsertId)
	return err
}

func (user *User) ToggleStatus() error {
	result, err := DB.Exec(`update user set defunct = not defunct,updated_at = NOW() where id = ?`, user.Id)
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("数据不存在")
	}
	return err
}

func (user *User) Update() error {
	result, err := DB.Exec(`update user set username = ?,updated_at = NOW() where id = ?`, user.Username, user.Id)
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("数据不存在")
	}
	return err
}

func (user *User) ChangePass() error {
	result, err := DB.Exec(`update user set password = ?, passsalt = ?,updated_at = NOW() where id = ?`, user.Password, user.PassSalt, user.Id)
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("数据不存在")
	}
	return err
}

func (user *User) Delete() error {
	result, err := DB.Exec(`delete from user where id = ?`, user.Id)
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("数据不存在")
	}
	return err
}

func (user *User) Response() map[string]interface{} {

	return map[string]interface{}{
		"id":       user.Id,
		"username": user.Username,
		"role":     user.Role,
		"nick":     user.Nick,
		"avatar":   user.Avatar,
		"submit":   user.Submit,
		"solved":   user.Solved,
		"defunct":  user.Defunct,
	}
}
