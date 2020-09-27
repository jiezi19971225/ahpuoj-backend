package entity

type SourceCode struct {
	SolutionId int    `gorm:"primaryKey;"`
	Source     string `json:"source"`
	Public     int    `json:"public"`
}

func (SourceCode) SourceCode() string {
	return "source_code"
}
