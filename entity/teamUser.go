package entity

import "time"

type TeamUser struct {
	ID        int       `json:"id"`
	TeamID    int       `json:"team_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TeamUser) TableName() string {
	return "team_user"
}
