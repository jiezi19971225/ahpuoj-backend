package controller

import (
	"ahpuoj/dto"
	"ahpuoj/entity"
	"ahpuoj/request"
	"ahpuoj/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetRoleList(c *gin.Context) {
	roles := []entity.Role{}
	err := ORM.Model(entity.Role{}).Where("id != 1").Find(&roles).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"data":    roles,
	})
}

func GetAdminList(c *gin.Context) {
	adminList := []dto.UserWithRoleDto{}
	err := ORM.Model(entity.User{}).Select("user.*,role.name as role").Joins("inner join role on user.role_id = role.id").Where("user.role_id != 1").Find(&adminList).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"data":    adminList,
	})
}

func AssignRole(c *gin.Context) {
	var req request.AssignRole
	err := c.ShouldBindJSON(&req)
	// TODO 先粗略处理 无法授予超级管理员角色
	if err != nil {
		panic(err)
	}
	if req.RoleId == 2 {
		panic(errors.New("无法授予超级管理员角色"))
	}

	user := entity.User{}
	err = ORM.Model(entity.User{}).Where("username = ?", req.UserName).Take(&user).Error
	if err != nil {
		panic(err)
	}
	role := entity.Role{ID: req.RoleId}
	err = ORM.Model(&role).First(&role).Error
	if err != nil {
		panic(err)
	}
	err = ORM.Model(&user).Update("role_id", req.RoleId).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "操作成功",
		"show":    true,
	})
}

func UnassignRole(c *gin.Context) {
	var req request.UnassignRole
	err := c.ShouldBindJSON(&req)
	if utils.CheckError(c, err, "请求参数错误") != nil {
		return
	}
	user := entity.User{ID: req.UserId}
	err = ORM.Model(entity.User{}).First(&user).Error
	if err != nil {
		panic(err)
	}
	// TODO 先粗略处理 无法撤销超级管理员角色
	if user.RoleId == 2 {
		panic(errors.New("无法撤销超级管理员角色"))
	}
	err = ORM.Model(&user).Update("role_id", 1).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "操作成功",
		"show":    true,
	})
}
