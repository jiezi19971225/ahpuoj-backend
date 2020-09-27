package entity

import (
	"ahpuoj/utils"
	"encoding/json"
	"gopkg.in/guregu/null.v4"
	"time"
)

type New struct {
	ID        int         `json:"id" uri:"id"`
	Title     string      `json:"title" binding:"required,max=20"`
	Content   null.String `json:"content"`
	Top       int         `json:"top" gorm:"default:0"`
	Defunct   int         `json:"defunct" gorm:"default:0"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func (New) TableName() string {
	return "new"
}

func (new *New) MarshalJSON() ([]byte, error) {
	type Alias New
	new.Content.String = utils.ConvertTextImgUrl(new.Content.String)
	return json.Marshal((*Alias)(new))
}
