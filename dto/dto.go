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

type RejudgeInfo struct {
	UserId      int    `db:"user_id"`
	SolutionId  int    `db:"solution_id"`
	ProblemId   int    `db:"problem_id"`
	Language    int    `db:"language"`
	TimeLimit   int    `db:"time_limit"`
	MemoryLimit int    `db:"memory_limit"`
	Source      string `db:"source"`
}
