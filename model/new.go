package model

import (
	"ahpuoj/utils"
	"encoding/json"
	"gopkg.in/guregu/null.v4"
	"time"
)

type New struct {
	ID        int         `db:"id" json:"id" uri:"id"`
	Title     string      `db:"title" json:"title" binding:"required,max=20"`
	Content   null.String `db:"content" json:"content"`
	Top       int         `db:"top" json:"top" gorm:"default:0"`
	Defunct   int         `db:"defunct" json:"defunct" gorm:"default:0"`
	CreatedAt time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt time.Time   `db:"updated_at" json:"updated_at"`
}

func (New) TableName() string {
	return "new"
}

func (new *New) MarshalJSON() ([]byte, error) {
	type Alias New
	new.Content.String = utils.ConvertTextImgUrl(new.Content.String)
	return json.Marshal((*Alias)(new))
}
