package entity

import (
	"ahpuoj/utils"
)

type Reply struct {
	ID          int                      `json:"id"`
	IssueId     int                      `json:"issue_id:"`
	UserId      int                      `json:"user_id"`
	ReplyId     int                      `json:"reply_id" binding:"gte=0"`
	ReplyUserId int                      `json:"reply_user_id" binding:"gte=0"`
	CreatedAt   utils.JSONDateTime       `json:"created_at"`
	UpdatedAt   utils.JSONDateTime       `json:"updated_at"`
	Content     utils.RelativeNullString `json:"content" binding:"required"`
	IsDeleted   int                      `json:"is_deleted"`
	Status      int                      `json:"status"`
}

func (Reply) TableName() string {
	return "reply"
}
