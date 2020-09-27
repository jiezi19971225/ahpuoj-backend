package controller

import (
	"ahpuoj/entity"
	"ahpuoj/utils"
	"crypto/sha1"
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func IndexUser(c *gin.Context) {

	results, total := userService.List(c)
	c.JSON(200, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    c.DefaultQuery("page", "1"),
		"perpage": c.DefaultQuery("perpage", "20"),
		"data":    results,
	})
}

func ToggleUserStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user := entity.User{
		ID: id,
	}
	ORM.Model(&user).Update("defunct", gorm.Expr("not defunct"))
	c.JSON(http.StatusOK, gin.H{
		"message": "更改用户状态成功",
		"show":    true,
	})
}

func ChangeUserPass(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Password string `json:"password" binding:"required,ascii,min=6,max=20"`
	}
	err := c.ShouldBindJSON(&req)
	if utils.CheckError(c, err, "请求参数错误") != nil {
		return
	}
	// 更新密码
	// 加盐处理 16位随机字符串
	salt := utils.GetRandomString(16)
	h := sha1.New()
	h.Write([]byte(salt))
	h.Write([]byte(req.Password))
	hashedPassword := fmt.Sprintf("%x", h.Sum(nil))
	user := entity.User{
		ID:       id,
		Password: hashedPassword,
		Passsalt: salt,
	}
	ORM.Model(&user).Updates(user)
	c.JSON(http.StatusOK, gin.H{
		"message": "更改用户密码成功",
		"show":    true,
	})
}
