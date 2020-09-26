package controller

import (
	"ahpuoj/model"
	"ahpuoj/request"
	"ahpuoj/utils"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func IndexNew(c *gin.Context) {
	param := c.Query("param")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))

	query := ORM
	if len(param) > 0 {
		query.Where("title like ?", "%"+param+"%")
	}
	var total int64
	query.Count(&total)
	news := []model.New{}
	err := query.Scopes(Paginate(c)).Order("top desc,id desc").Find(&news).Error
	if utils.CheckError(c, err, "") != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    page,
		"perpage": perpage,
		"data":    news,
	})
}

func ShowNew(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	new := model.New{}
	err := ORM.First(&new, id).Error
	if utils.CheckError(c, err, "") != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"new":     new,
	})
}

func StoreNew(c *gin.Context) {
	var req request.New
	err := c.ShouldBindJSON(&req)
	if utils.CheckError(c, err, "请求参数错误") != nil {
		return
	}
	new := model.New{
		Title:   req.Title,
		Content: null.StringFrom(req.Content),
	}
	err = ORM.Create(&new).Error
	if utils.CheckError(c, err, "") != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "新建新闻成功",
		"new":     new,
	})
}

func UpdateNew(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req request.New
	err := c.ShouldBindJSON(&req)
	if utils.CheckError(c, err, "请求参数错误") != nil {
		return
	}
	new := model.New{
		ID:      id,
		Title:   req.Title,
		Content: null.StringFrom(req.Content),
	}
	err = ORM.Model(&new).Updates(new).Error
	if utils.CheckError(c, err, "") != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "编辑新闻成功",
		"new":     new,
	})
}

func DeleteNew(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := ORM.Delete(model.New{}, id).Error
	if utils.CheckError(c, err, "") != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "删除新闻成功",
	})
}

func ToggleNewStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	new := model.New{
		ID: id,
	}
	err := ORM.Model(&new).Update("defunct", gorm.Expr("not defunct")).Error
	if utils.CheckError(c, err, "更改新闻状态失败，该新闻不存在") != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "更改新闻状态成功",
	})
}

func ToggleNewTopStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	new := model.New{}
	err := ORM.First(&new, id).Error
	if utils.CheckError(c, err, "") != nil {
		return
	}

	var newtop int
	if new.Top == 0 {
		var maxtop int
		ORM.Model(model.New{}).Select("max(top)").Scan(&maxtop)
		newtop = maxtop + 1
	} else {
		newtop = 0
	}
	new.Top = newtop

	err = ORM.Save(&new).Error

	if utils.CheckError(c, err, "更改新闻置顶状态失败，该新闻不存在") != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "更改新闻置顶状态成功",
	})
}
