package entity

import "time"

type ContestSeries struct {
	ID        int       `json:"id"`
	ContestID int       `json:"contest_id"`
	SeriesID  int       `json:"team_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ContestSeries) TableName() string {
	return "contest_series"
}
