package entity

import (
	"ahpuoj/utils"
	"time"
)

type New struct {
	ID        int                      `json:"id" uri:"id"`
	Title     string                   `json:"title" binding:"required,max=20"`
	Content   utils.RelativeNullString `json:"content"`
	Top       int                      `json:"top" gorm:"default:0"`
	Defunct   int                      `json:"defunct" gorm:"default:0"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
}

func (New) TableName() string {
	return "new"
}
