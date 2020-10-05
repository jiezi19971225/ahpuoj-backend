package entity

type Issue struct {
	ID        int    `json:"id"`
	Title     string `json:"title" binding:"required,max=20"`
	ProblemId int    `json:"problem_id" binding:"gte=0"`
	UserId    int    `json:"user_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	IsDeleted int    `json:"is_deleted"`
}

func (Issue) TableName() string {
	return "issue"
}
