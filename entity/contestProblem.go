package entity

import "time"

type ContestProblem struct {
	ID        int       `json:"id"`
	ContestID int       `json:"contest_id"`
	ProblemID int       `json:"problem_id"`
	Num       int       `json:"num"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ContestProblem) TableName() string {
	return "contest_problem"
}
