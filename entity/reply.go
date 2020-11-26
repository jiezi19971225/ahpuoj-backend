package entity

import (
	"ahpuoj/utils"
	"time"
)

type Reply struct {
	ID          int                      `json:"id"`
	IssueId     int                      `json:"issue_id:"`
	UserId      int                      `json:"user_id"`
	ReplyId     int                      `json:"reply_id"`
	ReplyUserId int                      `json:"reply_user_id"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	Content     utils.RelativeNullString `json:"content"`
	IsDeleted   int                      `json:"is_deleted"`
	Status      int                      `json:"status"`
}

func (Reply) TableName() string {
	return "reply"
}
