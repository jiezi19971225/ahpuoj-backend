package entity

import (
	"gopkg.in/guregu/null.v4"
	"time"
)

type Solution struct {
	ID         int         `gorm:"column:solution_id;" json:"solution_id"`
	ProblemId  int         `json:"problem_id"`
	TeamId     int         `json:"team_id"`
	UserId     int         `json:"user_id"`
	ContestId  int         `json:"contest_id"`
	Num        int         `json:"num"`
	Time       int         `json:"time"`
	Memory     int         `json:"memory"`
	InDate     time.Time   `json:"in_date"`
	Result     int         `json:"result"`
	Language   int         `json:"language"`
	IP         string      `gorm:"column:ip;" json:"ip"`
	JudgeTime  null.String `gorm:"column:judgetime;" json:"judgetime"`
	Valid      int         `json:"valid"`
	CodeLength int         `json:"code_length"`
	PassRate   float32     `json:"pass_rate"`
	LintError  int         `json:"lint_error"`
	Judger     null.String `json:"judger"`
}

func (Solution) TableName() string {
	return "solution"
}
