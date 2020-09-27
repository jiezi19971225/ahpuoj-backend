package entity

import (
	"gorm.io/gorm"
	"time"
)

type Team struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	IsDeleted gorm.DeletedAt `json:"is_deleted"`
	CreatorId int            `gorm:"column:user_id;" json:"creator_id"`
	Users     []User         `gorm:"association_autoupdate:false;many2many:team_user;" json:"userinfos"`
}

func (Team) TableName() string {
	return "team"
}
