package entity

type SourceCode struct {
	SolutionId int    `gorm:"primaryKey;"`
	Source     string `json:"source"`
	Public     int    `json:"public"`
}

func (SourceCode) TableName() string {
	return "source_code"
}
