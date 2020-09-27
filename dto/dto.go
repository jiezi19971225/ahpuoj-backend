package dto

import "ahpuoj/entity"

type CreatorInfo struct {
	Username string `json:"username"`
}

type TeamDto struct {
	entity.Team
	CreatorInfo
}

type ProblemDto struct {
	entity.Problem
	CreatorInfo
}
