package entity

import (
	"ahpuoj/utils"
	"gopkg.in/guregu/null.v4"
	"time"
)

type Problem struct {
	ID           int                      `json:"id"`
	Title        string                   `json:"title"`
	Description  utils.RelativeNullString `json:"description"`
	Level        int                      `json:"level"`
	Input        utils.RelativeNullString `json:"input"`
	Output       utils.RelativeNullString `json:"output"`
	SampleInput  null.String              `json:"sample_input"`
	SampleOutput null.String              `json:"sample_output"`
	Spj          int                      `json:"spj"`
	Hint         utils.RelativeNullString `json:"hint"`
	Defunct      int                      `json:"defunct"`
	TimeLimit    int                      `json:"time_limit"`
	MemoryLimit  int                      `json:"memory_limit"`
	Accepted     int                      `json:"accepted"`
	Submit       int                      `json:"submit"`
	Solved       int                      `json:"solved"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
	CreatorId    int                      `gorm:"column:user_id;" json:"creator_id"`
	Tags         []Tag                    `gorm:"many2many:problem_tag;joinForeignKey:problem_id;joinReferences:tag_id;" json:"tags"`
}

func (Problem) TableName() string {
	return "problem"
}
