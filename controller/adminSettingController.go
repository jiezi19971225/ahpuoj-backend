package controller

import (
	"ahpuoj/entity"
	"ahpuoj/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetSettings(c *gin.Context) {

	var config = make(map[string]interface{})
	var enableIssueString string
	err := ORM.Raw("select value from config where item = 'enable_issue'").Scan(&enableIssueString).Error
	if utils.CheckError(c, err, "") != nil {
		return
	}

	if enableIssueString == "true" {
		config["enable_issue"] = true
	} else {
		config["enable_issue"] = false
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "获取系统配置项成功",
		"config":  config,
	})
}

func SetSettings(c *gin.Context) {
	var req struct {
		EnableIssue bool `json:"enable_issue" binding:"required"`
	}
	var config = make(map[string]interface{})
	c.ShouldBindJSON(&req)
	var enableIssueString string
	if req.EnableIssue == true {
		enableIssueString = "true"
	} else {
		enableIssueString = "false"
	}
	err := ORM.Exec("update config set value = ? where item = 'enable_issue'", enableIssueString).Error
	if utils.CheckError(c, err, "") != nil {
		return
	}
	if enableIssueString == "true" {
		config["enable_issue"] = true
	} else {
		config["enable_issue"] = false
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "修改系统配置项成功",
		"config":  config,
		"show":    true,
	})
}

// TODO 功能待测试
// 后续系统更新新功能接口的权限二级管理员是没有的，调用这个接口手动同步
func FixPermission(c *gin.Context) {
	enforcer := entity.GetCasbin()
	var teams []struct {
		UserId int `db:"user_id"`
		TeamId int `db:"team_id"`
	}
	ORM.Model(entity.Team{}).Select("team.id as team_id,team.user_id").
		Joins(" inner join user on team.user_id = user.id").Joins("inner join role on user.role_id = role.id").
		Where("role.name = ?", "secondaryadmin").Find(&teams)
	for _, team := range teams {
		idStr := strconv.Itoa(team.UserId)
		teamIdStr := strconv.Itoa(team.TeamId)
		enforcer.AddPolicy(idStr, "/api/admin/team/"+teamIdStr+"/users", "POST")
		enforcer.AddPolicy(idStr, "/api/admin/team/"+teamIdStr, "PUT")
		enforcer.AddPolicy(idStr, "/api/admin/team/"+teamIdStr+"/user/:userid", "DELETE")
		enforcer.AddPolicy(idStr, "/api/admin/team/"+teamIdStr, "DELETE")
	}
	var problems []struct {
		UserId    int `db:"user_id"`
		ProblemId int `db:"problem_id"`
	}
	ORM.Model(entity.Problem{}).Select("team.id as team_id,team.user_id").
		Joins(" inner join user on team.user_id = user.id").Joins("inner join role on user.role_id = role.id").
		Where("role.name = ?", "secondaryadmin").Find(&problems)
	for _, problem := range problems {
		idStr := strconv.Itoa(problem.UserId)
		problemIdStr := strconv.Itoa(problem.ProblemId)
		enforcer.AddPolicy(idStr, "/api/admin/problem/"+problemIdStr, "PUT")
		enforcer.AddPolicy(idStr, "/api/admin/problem/"+problemIdStr, "DELETE")
		enforcer.AddPolicy(idStr, "/api/admin/problem/"+problemIdStr+"/status", "PUT")
		enforcer.AddPolicy(idStr, "/api/admin/problem/"+problemIdStr+"/data", "POST")
		enforcer.AddPolicy(idStr, "/api/admin/problem/"+problemIdStr+"/datafile", "POST")
		enforcer.AddPolicy(idStr, "/api/admin/problem/"+problemIdStr+"/data/:filename", "PUT")
		enforcer.AddPolicy(idStr, "/api/admin/problem/"+problemIdStr+"/data/:filename", "DELETE")
	}
	var contests []struct {
		UserId    int `db:"user_id"`
		ContestId int `db:"contest_id"`
	}
	for _, contest := range contests {
		idStr := strconv.Itoa(contest.UserId)
		contestIdStr := strconv.Itoa(contest.ContestId)
		DB.Select(&teams, "select contest.id as contest_id,contest.user_id from contest inner join user on contest.user_id = user.id inner join role on user.role_id = role.id where role.name = 'secondaryadmin'")
		enforcer.AddPolicy(idStr, "/api/admin/contest/"+contestIdStr, "PUT")
		enforcer.AddPolicy(idStr, "/api/admin/contest/"+contestIdStr, "DELETE")
		enforcer.AddPolicy(idStr, "/api/admin/contest/"+contestIdStr+"/status", "PUT")
		enforcer.AddPolicy(idStr, "/api/admin/contest/"+contestIdStr+"/users", "POST")
		enforcer.AddPolicy(idStr, "/api/admin/contest/"+contestIdStr+"/user/:userid", "DELETE")
		enforcer.AddPolicy(idStr, "/api/admin/contest/"+contestIdStr+"/team/:teamid", "POST")
		enforcer.AddPolicy(idStr, "/api/admin/contest/"+contestIdStr+"/team/:teamid", "DELETE")
		enforcer.AddPolicy(idStr, "/api/admin/contest/"+contestIdStr+"/team/:teamid/users", "POST")
		enforcer.AddPolicy(idStr, "/api/admin/contest/"+contestIdStr+"/team/:teamid/allusers", "POST")
		enforcer.AddPolicy(idStr, "/api/admin/contest/"+contestIdStr+"/team/:teamid/user/:userid", "DELETE")
	}
	var serieses []struct {
		UserId   int `db:"user_id"`
		SeriesId int `db:"contest_id"`
	}
	for _, series := range serieses {
		idStr := strconv.Itoa(series.UserId)
		seriesIdStr := strconv.Itoa(series.SeriesId)
		enforcer.AddPolicy(idStr, "/api/admin/series/"+seriesIdStr, "PUT")
		enforcer.AddPolicy(idStr, "/api/admin/series/"+seriesIdStr, "DELETE")
		enforcer.AddPolicy(idStr, "/api/admin/series/"+seriesIdStr+"/status", "PUT")
		enforcer.AddPolicy(idStr, "/api/admin/series/"+seriesIdStr+"/contest/:contestid", "POST")
		enforcer.AddPolicy(idStr, "/api/admin/series/"+seriesIdStr+"/contest/:contestid", "DELETE")
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "权限修复成功",
	})
}
