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

type UserWithPasswordDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RejudgeInfo struct {
	UserId      int
	SolutionId  int
	ProblemId   int
	Language    int
	TimeLimit   int
	MemoryLimit int
	Source      string
}

type ContestTeamDto struct {
	entity.Team
	Userinfos []entity.User `json:"userinfos"`
}
