package entity

import "time"

type ContestUser struct {
	ID        int       `json:"id"`
	ContestID int       `json:"contest_id"`
	UserID    int       `json:"user_id"`
	TeamID    int       `json:"team_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ContestUser) TableName() string {
	return "contest_user"
}
