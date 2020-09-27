package entity

import (
	"ahpuoj/utils"
	"encoding/json"
	"gopkg.in/guregu/null.v4"
	"time"
)

type Problem struct {
	ID           int         `json:"id"`
	Title        string      `json:"title"`
	Description  null.String `json:"description"`
	Level        int         `json:"level"`
	Input        null.String `json:"input"`
	Output       null.String `json:"output"`
	SampleInput  null.String `json:"sample_input"`
	SampleOutput null.String `json:"sample_output"`
	Spj          int         `json:"spj"`
	Hint         null.String `json:"hint"`
	Defunct      int         `json:"defunct"`
	TimeLimit    int         `json:"time_limit"`
	MemoryLimit  int         `json:"memory_limit"`
	Accepted     int         `json:"accepted"`
	Submit       int         `json:"submit"`
	Solved       int         `json:"solved"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	CreatorId    int         `gorm:"column:user_id;" json:"creator_id"`
	Tags         []Tag       `gorm:"many2many:problem_tag;" json:"tags"`
}

func (Problem) TableName() string {
	return "problem"
}

func (problem *Problem) MarshalJSON() ([]byte, error) {
	type Alias Problem
	problem.Description.String = utils.ConvertTextImgUrl(problem.Description.String)
	problem.Input.String = utils.ConvertTextImgUrl(problem.Input.String)
	problem.Output.String = utils.ConvertTextImgUrl(problem.Output.String)
	problem.Hint.String = utils.ConvertTextImgUrl(problem.Hint.String)
	return json.Marshal((*Alias)(problem))
}
