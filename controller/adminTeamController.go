package controller

import (
	"ahpuoj/entity"
	"ahpuoj/request"
	"ahpuoj/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetTeam(c *gin.Context) {
	var team entity.Team
	id, _ := strconv.Atoi(c.Param("id"))
	err := ORM.Model(entity.Team{}).First(&team, id).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"team":    team,
	})
}

func IndexTeam(c *gin.Context) {

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))
	results, total := teamService.List(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    page,
		"perpage": perpage,
		"data":    results,
	})
}

func GetAllTeams(c *gin.Context) {
	var teams []entity.Team
	ORM.Model(entity.Team{}).Order("id desc").Find(&teams)
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"teams":   teams,
	})
}

func IndexTeamUser(c *gin.Context) {

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))
	teamId, _ := strconv.Atoi(c.Param("id"))

	team := entity.Team{ID: teamId}
	users, total := teamService.Users(&team, c)

	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    page,
		"perpage": perpage,
		"data":    users,
	})
}

func AddTeamUsers(c *gin.Context) {
	var req struct {
		UserList string `json:"userlist" binding:"required"`
	}
	id, _ := strconv.Atoi(c.Param("id"))
	c.ShouldBindJSON(&req)
	var team entity.Team
	ORM.First(&team, id)
	infos := teamService.AddUsers(&team, req.UserList)
	c.JSON(http.StatusOK, gin.H{
		"message": "操作成功",
		"show":    true,
		"info":    infos,
	})
}

func StoreTeam(c *gin.Context) {
	var req request.Team
	user, _ := GetUserInstance(c)
	err := c.ShouldBindJSON(&req)
	if err != nil {
		panic(err)
	}
	team := entity.Team{
		Name:      req.Name,
		CreatorId: user.ID,
	}
	err = ORM.Create(&team).Error
	// TODO 先这样处理 给次级管理员添加权限
	if utils.CheckError(c, err, "") != nil {
		return
	}
	idStr := strconv.Itoa(user.ID)
	teamIdStr := strconv.Itoa(team.ID)
	if user.Role != "admin" {
		enforcer := entity.GetCasbin()
		enforcer.AddPolicy(idStr, "/api/admin/team/"+teamIdStr+"/users", "POST")
		enforcer.AddPolicy(idStr, "/api/admin/team/"+teamIdStr, "PUT")
		enforcer.AddPolicy(idStr, "/api/admin/team/"+teamIdStr+"/user/:userid", "DELETE")
		enforcer.AddPolicy(idStr, "/api/admin/team/"+teamIdStr, "DELETE")
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "新建团队成功",
		"show":    true,
		"team":    team,
	})
}

func UpdateTeam(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req request.Team
	err := c.ShouldBindJSON(&req)
	if err != nil {
		panic(err)
	}
	team := entity.Team{
		ID:   id,
		Name: req.Name,
	}
	err = ORM.Model(&team).Updates(team).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "编辑团队成功",
		"show":    true,
		"team":    team,
	})
}

func DeleteTeam(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := ORM.Delete(entity.Team{}, id).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "删除团队成功",
		"show":    true,
	})
}

func DeleteTeamUser(c *gin.Context) {
	teamId, _ := strconv.Atoi(c.Param("id"))
	userId, _ := strconv.Atoi(c.Param("userid"))
	teamService.DeleteUser(&entity.Team{ID: teamId}, entity.User{ID: userId})
	c.JSON(http.StatusOK, gin.H{
		"message": "删除团队成员成功",
		"show":    true,
	})
}
