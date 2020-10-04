package entity

import (
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"time"
)

type Contest struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
	Description null.String    `json:"description"`
	Defunct     int            `json:"defunct"`
	Private     int            `json:"private"`
	TeamMode    int            `json:"team_mode"`
	LangMask    int            `gorm:"column:langmask;"  json:"langmask"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	IsDeleted   gorm.DeletedAt `json:"is_deleted"`
	CreatorId   int            `gorm:"column:user_id;" json:"creator_id"`
	Problems    []Problem      `gorm:"many2many:contest_problem;" json:"problems"`
	Users       []User         `gorm:"many2many:contest_user;" json:"users"`
}

func (Contest) TableName() string {
	return "contest"
}
