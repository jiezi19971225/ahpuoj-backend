package entity

import (
	"ahpuoj/utils"
)

type Issue struct {
	ID        int                `json:"id"`
	Title     string             `json:"title"`
	ProblemId int                `json:"problem_id"`
	UserId    int                `json:"user_id"`
	CreatedAt utils.JSONDateTime `json:"created_at"`
	UpdatedAt utils.JSONDateTime `json:"updated_at"`
	IsDeleted int                `json:"is_deleted"`
	Replys    []Reply            `json:"replys" gorm:"foreignKey:IssueId"`
}

func (Issue) TableName() string {
	return "issue"
}
