package entity

type RuntimeInfo struct {
	SolutionId int    `gorm:"primaryKey;"`
	Error      string `json:"error"`
}

func (RuntimeInfo) TableName() string {
	return "runtimeinfo"
}
