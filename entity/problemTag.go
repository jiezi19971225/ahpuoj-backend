package entity

import "time"

type ProblemTag struct {
	ID        int       `json:"id"`
	ProblemID int       `json:"problem_id"`
	TagID     int       `json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ProblemTag) TableName() string {
	return "problem_tag"
}
