package entity

import "time"

type ContestTeam struct {
	ID        int       `json:"id"`
	ContestID int       `json:"contest_id"`
	TeamID    int       `json:"team_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ContestTeam) TableName() string {
	return "contest_team"
}
