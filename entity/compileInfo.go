package entity

type CompileInfo struct {
	SolutionId int    `gorm:"primaryKey;"`
	Error      string `json:"error"`
}

func (CompileInfo) TableName() string {
	return "compileinfo"
}
