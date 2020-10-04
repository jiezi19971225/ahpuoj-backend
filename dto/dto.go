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

type ContestDto struct {
	entity.Contest
	CreatorInfo
}

type ContestDetailDto struct {
	entity.Contest
	Problems string `json:"problems"`
}
