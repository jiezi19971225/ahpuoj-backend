package dto

import (
	"ahpuoj/entity"
	"ahpuoj/utils"
	"gopkg.in/guregu/null.v4"
)

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

type SeriesDto struct {
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

type UserWithRoleDto struct {
	entity.User
	Role string `json:"role"`
}

type ProblemListItemDto struct {
	ID       int          `json:"id"`
	Title    string       `json:"title"`
	Accepted int          `json:"accepted"`
	Submit   int          `json:"submit"`
	Solved   int          `json:"solved"`
	Status   int          `json:"status"`
	Tags     []entity.Tag `gorm:"many2many:problem_tag;" json:"tags"`
	Level    int          `json:"level"`
}

type ContestInfoDto struct {
	entity.Contest
	StartTime utils.JSONDateTime `json:"start_time"`
	EndTime   utils.JSONDateTime `json:"end_time"`
	Status    int                `json:"status"`
}

type SolutionInfoDto struct {
	entity.Solution
	InDate       utils.JSONDateTime `json:"in_date"`
	JudgeTime    utils.JSONDateTime `gorm:"column:judgetime;" json:"judgetime"`
	Public       int                `json:"public"`
	Username     string             `json:"username"`
	Nick         string             `json:"nick"`
	Avatar       string             `json:"avatar"`
	ProblemTitle string             `json:"problem_title"`
}

type IssueInfoDto struct {
	entity.Issue
	CreatedAt    utils.JSONDateTime `json:"created_at"`
	UpdatedAt    utils.JSONDateTime `json:"updated_at"`
	Username     string             `json:"username"`
	Nick         string             `json:"nick"`
	Avatar       string             `json:"avatar"`
	ReplyCount   int                `json:"reply_count"`
	ProblemTitle null.String        `json:"ptitle"`
}

type ReplyInfoDto struct {
	entity.Reply
	CreatedAt     utils.JSONDateTime `json:"created_at"`
	UpdatedAt     utils.JSONDateTime `json:"updated_at"`
	Username      string             `json:"username"`
	ReplyUserNick string             `json:"reply_user_nick"`
	Nick          string             `json:"user_nick"`
	Avatar        string             `json:"avatar"`
	ReplyCount    int                `json:"reply_count"`
	SubReplys     []ReplyInfoDto     `json:"sub_replys"`
}
