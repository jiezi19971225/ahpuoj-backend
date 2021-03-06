package controller

import (
	"ahpuoj/entity"
	"ahpuoj/request"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func IndexTag(c *gin.Context) {

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))
	results, total := tagService.List(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    page,
		"perpage": perpage,
		"data":    results,
	})
}

func GetAllTags(c *gin.Context) {
	var tags []entity.Tag
	ORM.Model(entity.Tag{}).Order("id desc").Find(&tags)
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"tags":    tags,
	})
}

func StoreTag(c *gin.Context) {
	var req request.Tag
	err := c.ShouldBindJSON(&req)
	if err != nil {
		panic(err)
	}
	tag := entity.Tag{
		Name: req.Name,
	}

	err = ORM.Create(&tag).Error
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "新建标签成功",
		"tag":     tag,
		"show":    true,
	})
}

func UpdateTag(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req request.Tag
	err := c.ShouldBindJSON(&req)
	if err != nil {
		panic(err)
	}
	tag := entity.Tag{
		ID:   id,
		Name: req.Name,
	}
	err = ORM.Model(&tag).Updates(tag).Error
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "编辑标签成功",
		"show":    true,
		"tag":     tag,
	})
}

func DeleteTag(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	tag := entity.Tag{
		ID: id,
	}
	tagService.DeleteRecord(&tag)
	c.JSON(http.StatusOK, gin.H{
		"message": "删除标签成功",
		"show":    true,
	})
}
