package service

import (
	"ahpuoj/dto"
	"ahpuoj/entity"
	"ahpuoj/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SeriesService struct {
	*gorm.DB
}

func (this *SeriesService) List(c *gin.Context) ([]dto.SeriesDto, int64) {
	param := c.Query("param")
	query := this.Model(entity.Series{})

	if len(param) > 0 {
		query.Where("name like ?", "%"+param+"%")
	}
	var total int64
	query.Count(&total)
	var results []dto.SeriesDto
	query.Scopes(utils.Paginate(c)).Order("series.id desc").Select("series.*", "user.username").Joins("inner join user on series.user_id = user.id").Find(&results)

	return results, total
}

func (this *SeriesService) AddContest(series *entity.Series, contest *entity.Contest) {
	// 检查系列赛是否存在
	err := this.Model(&series).First(&series).Error
	if err != nil {
		panic(err)
	}
	// 检查竞赛&作业是否存在
	err = this.Model(&contest).First(&contest).Error
	if err != nil {
		panic(err)
	}
	// 检查是否已经添加进了系列中
	var count int64
	this.Model(&entity.ContestSeries{}).Where("series_id = ? and contest_id = ?", series.ID, contest.ID).Count(&count)
	if count > 0 {
		panic(errors.New("该竞赛作业已经在该系列中"))
	}
	err = this.Create(&entity.ContestSeries{
		SeriesID:  series.ID,
		ContestID: contest.ID,
	}).Error
	if err != nil {
		panic(err)
	}
}
