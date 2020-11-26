package controller

import (
	"ahpuoj/entity"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ToggleIssueStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	issue := entity.Issue{
		ID: id,
	}
	err := ORM.Model(&issue).Update("is_deleted", gorm.Expr("not is_deleted")).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "更改主题状态成功",
		"show":    true,
	})
}

func ToggleReplyStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	reply := entity.Reply{
		ID: id,
	}
	err := ORM.Model(&reply).Update("is_deleted", gorm.Expr("not is_deleted")).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "更改回复状态成功",
		"show":    true,
	})
}
