package entity

import "time"

type TeamUser struct {
	ID        int       `db:"id" json:"id"`
	TeamID    int       `db:"team_id" json:"team_id"`
	UserID    int       `db:"user_id" json:"user_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (TeamUser) TableName() string {
	return "team_user"
}
