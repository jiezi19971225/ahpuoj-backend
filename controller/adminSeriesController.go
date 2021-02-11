package controller

import (
	"ahpuoj/entity"
	"ahpuoj/request"
	"ahpuoj/utils"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func IndexSeries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))
	results, total := seriesService.List(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    page,
		"perpage": perpage,
		"data":    results,
	})
}

func ShowSeries(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))
	series := entity.Series{}
	err := ORM.First(&series, id).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"series":  series,
	})
}

func IndexSeriesContest(c *gin.Context) {
	param := c.Query("param")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))
	seriesId, _ := strconv.Atoi(c.Param("id"))
	query := ORM.Model(entity.Contest{}).Joins("inner join contest_series on contest.id = contest_series.contest_id").Where("contest_series.series_id = ?", seriesId)
	if len(param) > 0 {
		query.Where("contest.name like", "%"+param+"%")
	}
	contests := []entity.Contest{}
	var total int64
	query.Count(&total)
	err := query.Scopes(utils.Paginate(c)).Order("contest.id desc").Find(&contests).Error

	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    page,
		"perpage": perpage,
		"data":    contests,
	})
}

func StoreSeries(c *gin.Context) {
	var req request.Series
	err := c.ShouldBindJSON(&req)
	user, _ := GetUserInstance(c)
	if err != nil {
		panic(err)
	}
	series := entity.Series{
		Name:        req.Name,
		Description: req.Description,
		TeamMode:    req.TeamMode,
		CreatorId:   user.ID,
	}
	err = ORM.Create(&series).Error

	if err != nil {
		panic(err)
	}
	idStr := strconv.Itoa(user.ID)
	seriesIdStr := strconv.Itoa(series.ID)
	if user.Role != "admin" {
		enforcer := entity.GetCasbin()
		enforcer.AddPolicy(idStr, "/api/admin/series/"+seriesIdStr, "PUT")
		enforcer.AddPolicy(idStr, "/api/admin/series/"+seriesIdStr, "DELETE")
		enforcer.AddPolicy(idStr, "/api/admin/series/"+seriesIdStr+"/status", "PUT")
		enforcer.AddPolicy(idStr, "/api/admin/series/"+seriesIdStr+"/contest/:contestid", "POST")
		enforcer.AddPolicy(idStr, "/api/admin/series/"+seriesIdStr+"/contest/:contestid", "DELETE")
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "新建系列赛成功",
		"show":    true,
		"series":  series,
	})
}

func UpdateSeries(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req request.Series
	err := c.ShouldBindJSON(&req)
	if err != nil {
		panic(err)
	}
	series := entity.Series{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		TeamMode:    req.TeamMode,
	}
	// 默认 updates 只会更新非0值，需要手动指定
	err = ORM.Select("name", "description", "team_mode").Model(&series).Updates(series).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "编辑系列赛成功",
		"show":    true,
		"series":  series,
	})
}

func ToggleSeriesStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	series := entity.Series{
		ID: id,
	}
	err := ORM.Model(&series).Update("defunct", gorm.Expr("not defunct")).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "更改系列赛状态成功",
		"show":    true,
	})
}

func DeleteSeries(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	series := entity.Series{
		ID: id,
	}
	err := ORM.Delete(&series).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "删除系列赛成功",
		"show":    true,
	})
}

func AddSeriesContest(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	contestId, _ := strconv.Atoi(c.Param("contestid"))

	series := entity.Series{ID: id}
	contest := entity.Contest{ID: contestId}
	seriesService.AddContest(&series, &contest)

	c.JSON(http.StatusOK, gin.H{
		"message": "添加竞赛&作业成功",
		"show":    true,
	})
}

func DeleteSeriesContest(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	contestId, _ := strconv.Atoi(c.Param("contestid"))
	series := entity.Series{ID: id}
	contest := entity.Contest{ID: contestId}
	err := ORM.Model(series).Association("Contests").Delete(&contest)
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "删除系列赛竞赛&作业成功",
		"show":    true,
	})
}
