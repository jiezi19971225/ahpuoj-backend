package entity

type Role struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (Role) TableName() string {
	return "role"
}
