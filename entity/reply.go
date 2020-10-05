package entity

import (
	"ahpuoj/utils"
	"encoding/json"
)

type Reply struct {
	ID          int    `json:"id"`
	IssueId     int    `json:"issue_id"`
	UserId      int    `json:"user_id"`
	ReplyId     int    `json:"reply_id" binding:"gte=0"`
	ReplyUserId int    `json:"reply_user_id" binding:"gte=0"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Content     string `json:"content" binding:"required"`
	IsDeleted   int    `json:"is_deleted"`
	Status      int    `json:"status"`
}

func (reply *Reply) MarshalJSON() ([]byte, error) {
	type Alias Reply
	reply.Content = utils.ConvertTextImgUrl(reply.Content)
	return json.Marshal((*Alias)(reply))
}

func (Reply) TableName() string {
	return "reply"
}
