package entity

import (
	"gorm.io/gorm"
	"time"
)

type Series struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	TeamMode    int            `json:"team_mode"`
	Defunct     int            `json:"defunct"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	IsDeleted   gorm.DeletedAt `json:"is_deleted"`
	CreatorId   int            `gorm:"column:user_id;" json:"creator_id"`
	Contests    []Contest      `gorm:"many2many:contest_series;" json:"contests"`
}

func (Series) TableName() string {
	return "series"
}
