package dto

import "ahpuoj/entity"

type TeamDto struct {
	entity.Team
	Username string `json:"username"`
}
