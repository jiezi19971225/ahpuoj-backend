package controller

import (
	"ahpuoj/config"
	"ahpuoj/constant"
	"ahpuoj/dto"
	"ahpuoj/entity"
	"ahpuoj/mq"
	"ahpuoj/request"
	"ahpuoj/utils"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

/*
前台用户登陆后可访问api的控制器
*/

// 用户获取用户信息
func GetUser(c *gin.Context) {

	user, _ := c.Get("user")
	if user, ok := user.(dto.UserWithRoleDto); ok {
		c.JSON(http.StatusOK, gin.H{
			"message": "用户信息获取成功",
			"user":    user,
		})
	}
}

// 账号设置中重设昵称的接口
func ResetNick(c *gin.Context) {
	var user dto.UserWithRoleDto
	user, _ = GetUserInstance(c)
	var req struct {
		Nick string `json:"nick" binding:"required,max=20"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		panic(err)
	}
	user.Nick = req.Nick
	err = ORM.Model(user.User).Update("nick", req.Nick).Error
	if err != nil {
		panic(errors.New("该昵称已被使用"))
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "昵称修改成功",
		"show":    true,
		"user":    user,
	})
}

// 账号设置中重设密码的接口
func ResetPassword(c *gin.Context) {
	var user dto.UserWithRoleDto
	user, _ = GetUserInstance(c)
	var req struct {
		OldPassword     string `json:"oldpassword" binding:"required,ascii,min=6,max=20"`
		Password        string `json:"password" binding:"required,ascii,min=6,max=20"`
		ConfirmPassword string `json:"confirmpassword" binding:"required,ascii,min=6,max=20"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		panic(err)
	}
	h := sha1.New()
	h.Write([]byte(user.Passsalt))
	h.Write([]byte(req.OldPassword))
	hashedOldPassword := fmt.Sprintf("%x", h.Sum(nil))

	if hashedOldPassword != user.Password {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "原密码错误",
			"show":    true,
		})
		return
	}

	// 更新密码
	// 加盐处理 16位随机字符串
	salt := utils.GetRandomString(16)
	h.Reset()
	h.Write([]byte(salt))
	h.Write([]byte(req.Password))
	hashedPassword := fmt.Sprintf("%x", h.Sum(nil))
	if hashedPassword == user.Password {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "新密码不能和原密码相同",
			"show":    true,
		})
		return
	}
	err = ORM.Model(user.User).Updates(map[string]interface{}{
		"password": hashedPassword,
		"passsalt": salt,
	}).Error

	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "密码修改成功",
		"show":    true,
	})
}

// 用户提交测试运行的接口
func SubmitToTestRun(c *gin.Context) {
	var err error
	var req struct {
		Language  int    `json:"language" binding:"gte=0,lte=17"`
		InputText string `json:"input_text"  binding:"max=65535"`
		Source    string `json:"source"  binding:"required,min=2,max=65535"`
	}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		panic(errors.New("提交失败,表单参数错误"))
	}

	conn := REDIS.Get()
	defer conn.Close()

	testrunCount, _ := redis.Int(conn.Do("incr", "testrun:count"))
	testrunCountStr := strconv.Itoa(testrunCount)
	jsondata, _ := json.Marshal(gin.H{
		"UserId":       0,
		"TestrunCount": testrunCountStr,
		"SolutionId":   0,
		"ProblemId":    0,
		"Language":     req.Language,
		"TimeLimit":    1,
		"MemoryLimit":  64,
		"Source":       req.Source,
		"InputText":    req.InputText,
	})
	mq.Publish("oj", "problem", jsondata)
	//等待评测机评判
	var reinfo, ceinfo, costomOut string
	queryTimes := 0
	for {
		queryTimes += 1
		if queryTimes > 100 {
			break
		}
		exist, _ := redis.Int(conn.Do("exists", "testrun:"+testrunCountStr))
		if exist == 1 {
			values, _ := redis.Values(conn.Do("hmget", "testrun:"+testrunCountStr, "ceinfo", "reinfo", "custom_out"))
			redis.Scan(values, &reinfo, &ceinfo, &costomOut)
		} else {
			time.Sleep(time.Second)
		}
	}
	customOutput := reinfo
	if len(ceinfo) > 0 {
		customOutput = ceinfo
	}
	if len(costomOut) > 0 {
		customOutput = costomOut
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "测试运行成功",
		"show":          true,
		"custom_output": customOutput,
	})
}

// 用户提交评测的接口
func SubmitToJudge(c *gin.Context) {
	var err error
	var contest entity.Contest
	user, _ := GetUserInstance(c)
	var req struct {
		ProblemId int    `json:"problem_id" binding:"required"`
		Language  int    `json:"language" binding:"gte=0,lte=17"`
		ContestId int    `json:"contest_id"`
		Num       int    `json:"num" binding:"omitempty,gte=0"`
		Source    string `json:"source"  binding:"required,min=2,max=65535"`
	}

	err = c.ShouldBindJSON(&req)
	if err != nil {
		panic(errors.New("提交失败"))
	}

	var problem entity.Problem
	err = ORM.Model(entity.Problem{}).First(&problem, req.ProblemId).Error

	if err != nil {
		panic(errors.New("提交失败，问题不存在"))
	}

	submitable := false

	// 管理员无提交限制
	if user.Role != "user" {
		submitable = true
	} else {
		// 比赛的提交
		if req.ContestId > 0 {

			err = ORM.Model(entity.Contest{}).First(&contest, req.ContestId).Error
			if err != nil {
				panic(errors.New("提交失败，竞赛不存在"))
			}
			// 非管理员只有在比赛进行过程中并且有参加权限才能提交

			status := contestService.CalcStatus(&contest)
			// 比赛进行中
			if status == constant.CONTEST_RUNNING {
				// 公开
				if contest.Private == 0 {
					submitable = true
				} else {
					// 检测是否有提交权限
					var temp int
					ORM.Raw("select count(1) from contest_user where contest_id = ? and user_id = ?", req.ContestId, user.ID).Scan(&temp)
					if temp > 0 {
						submitable = true
					}
				}
			}
		} else { // 平时的提交
			// 如果只是一般用户无法提交保留中的题目
			if problem.Defunct == 0 {
				submitable = true
			}
		}
	}

	if submitable {
		var teamId int
		// 如果为团队赛模式，并且非管理员提交，查询当前用户的teamId
		if contest.TeamMode == 1 && user.Role != "admin" {
			ORM.Raw("select team_id from contest_team_user ctu where ctu.contest_id = ? and ctu.user_id = ?", contest.ID, user.ID).Scan(&teamId)
		}
		solution := entity.Solution{
			ProblemId:  req.ProblemId,
			TeamId:     teamId,
			UserId:     user.ID,
			ContestId:  req.ContestId,
			Num:        req.Num,
			InDate:     time.Now(),
			IP:         c.ClientIP(),
			Language:   req.Language,
			CodeLength: len(req.Source),
		}
		err = ORM.Create(&solution).Error
		if err != nil {
			panic(errors.New("保存提交记录失败"))
		}
		sourceCode := entity.SourceCode{
			SolutionId: solution.ID,
			Source:     req.Source,
		}
		err = ORM.Create(&sourceCode).Error
		if err != nil {
			panic(errors.New("保存提交记录失败"))
		}
		// 将判题任务推入消息队列
		jsondata, _ := json.Marshal(gin.H{
			"UserId":       user.ID,
			"TestrunCount": 0,
			"SolutionId":   solution.ID,
			"ProblemId":    req.ProblemId,
			"Language":     req.Language,
			"TimeLimit":    problem.TimeLimit,
			"MemoryLimit":  problem.MemoryLimit,
			"Source":       req.Source,
			"InputText":    "",
		})

		mq.Publish("oj", "problem", jsondata)
		c.JSON(http.StatusOK, gin.H{
			"message":  "提交成功",
			"show":     true,
			"solution": solution,
		})
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "对不起，你没有提交权限",
			"show":    true,
		})
	}
}

// 切换代码公开状态
func ToggleSolutionStatus(c *gin.Context) {
	var user dto.UserWithRoleDto
	id, _ := strconv.Atoi(c.Param("id"))

	user, _ = GetUserInstance(c)

	var solutionUserId int
	ORM.Raw("select user_id from solution where solution_id = ?", id).Scan(&solutionUserId)
	if user.ID != solutionUserId {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "对不起，你没有修改权限",
		})
		return
	}
	ORM.Model(entity.SourceCode{}).Where("solution_id = ?", id).Update("public", gorm.Expr("not public"))
	c.JSON(http.StatusOK, gin.H{
		"message": "修改代码公开状态成功",
		"show":    true,
	})
}

// 下载题目数据文件
func DownloadDataFile(c *gin.Context) {
	var user dto.UserWithRoleDto
	user, _ = GetUserInstance(c)

	pidStr := c.Query("pid")
	sidStr := c.Query("sid")
	filename := c.Query("filename")
	pid, _ := strconv.Atoi(pidStr)
	sid, _ := strconv.Atoi(sidStr)

	// 检验提交是否存在
	var temp int
	ORM.Raw("select count(1) from solution where solution_id = ? and problem_id = ? and user_id = ?", sid, pid, user.ID).Scan(&temp)
	if temp == 0 {
		panic(errors.New("数据不存在"))
	}
	// 检验错误信息是否与数据库信息匹配
	var runtimeinfo entity.RuntimeInfo
	err := ORM.Model(entity.RuntimeInfo{}).Where("solution_id = ?", sid).Error
	if err != nil {
		panic(err)
	}
	errFilename := runtimeinfo.Error
	errFilenameWithoutSuffix := strings.TrimSuffix(errFilename, filepath.Ext(errFilename))
	filenameWithoutSuffix := strings.TrimSuffix(filename, filepath.Ext(filename))
	if errFilenameWithoutSuffix != filenameWithoutSuffix {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{
				"message": "数据不存在",
				"show":    true,
			})
		return
	}

	// 读取文件
	dataDir, _ := config.Conf.GetValue("project", "datadir")
	baseDir := dataDir + "/" + strconv.FormatInt(int64(pid), 10)
	dataFileName := baseDir + "/" + filename

	c.Header("Content-Disposition", `attachment; filename=`+filename+".txt")
	c.Header("Content-Type", "application/octet-stream")
	c.File(dataFileName)
}

// 账号设置上传头像
func UploadAvatar(c *gin.Context) {
	file, header, err := c.Request.FormFile("img")
	ext := path.Ext(header.Filename)
	if err != nil {
		panic(errors.New("头像上传失败，参数错误"))
	}
	url, err := utils.SaveFile(file, ext, "avatars")
	if err != nil {
		panic(err)
	}

	var user dto.UserWithRoleDto
	user, _ = GetUserInstance(c)
	// 如果不是默认头像 删除原头像
	defaultAvatar, _ := config.Conf.GetValue("preset", "avatar")
	if user.Avatar != defaultAvatar {
		webDir, _ := config.Conf.GetValue("project", "webdir")
		projectPath := webDir + "/"
		os.Remove(projectPath + user.Avatar)
	}
	ORM.Model(user.User).Update("avatar", url)

	c.JSON(http.StatusOK, gin.H{
		"message": "头像上传成功",
		"show":    true,
		"url":     url,
	})
}

// 发布主题帖
func PostIssue(c *gin.Context) {
	var err error
	var user dto.UserWithRoleDto

	user, _ = GetUserInstance(c)
	var req request.Issue
	err = c.ShouldBindJSON(&req)
	if err != nil {
		panic(err)
	}
	if req.ProblemId != 0 {
		var temp int
		ORM.Raw("select count(1) from problem where id = ?", req.ProblemId).Scan(&temp)
		if temp == 0 {
			panic(errors.New("发布讨论主题失败，问题不存在"))
		}
	}
	issue := entity.Issue{
		Title:     req.Title,
		ProblemId: req.ProblemId,
		UserId:    user.ID,
	}
	err = ORM.Create(&issue).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "发布讨论主题成功",
		"show":    true,
		"issue":   issue,
	})
}

// 回复主题帖
func ReplyToIssue(c *gin.Context) {
	var err error
	var user dto.UserWithRoleDto
	issueId, _ := strconv.Atoi(c.Param("id"))

	user, _ = GetUserInstance(c)
	var req request.Reply
	err = c.ShouldBindJSON(&req)
	if err != nil {
		panic(err)
	}
	// 主题是否存在
	var temp int

	ORM.Raw("select count(1) from issue where id = ?", issueId).Scan(&temp)
	if temp == 0 {
		panic(errors.New("发布回复失败，目标主题不存在"))
	}
	// 如果是对回复的回复 检查该回复是否存在
	if req.ReplyId != 0 {
		ORM.Raw("select count(1) from reply where id = ?", req.ReplyId).Scan(&temp)
		if temp == 0 {
			panic(errors.New("发布回复失败，目标回复不存在"))
		}
	}
	reply := entity.Reply{
		IssueId:     issueId,
		UserId:      user.ID,
		ReplyId:     req.ReplyId,
		ReplyUserId: req.ReplyUserId,
		Content:     utils.RelativeNullString(null.StringFrom(req.Content)),
	}
	err = ORM.Create(&reply).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "发布回复成功",
		"reply":   reply,
		"show":    true,
	})
}

// 获取回复我的信息帖子列表
func GetMyReplys(c *gin.Context) {
	var user dto.UserWithRoleDto
	user, _ = GetUserInstance(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))

	query := ORM.Model(entity.Reply{}).Joins("inner join user on reply.user_id = user.id").Joins("inner join issue on reply.issue_id = issue.id").
		Where("reply.reply_user_id = ?", user.ID)
	// 管理员可以查看被删除的回复
	if user.Role != "user" {
		query.Unscoped()
	}
	var total int64
	query.Count(&total)
	replys := []dto.ReplyInfoDto{}
	err := query.Scopes(Paginate(c)).Select("user.username,user.nick,user.avatar,reply.*,'' as rnick,(select count(1) from reply  r where reply.id = r.reply_id) as reply_count,issue.title as issue_title").
		Find(&replys).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    page,
		"perpage": perpage,
		"replys":  replys,
	})

}

// 获取最近一次提交的代码
func GetLatestSource(c *gin.Context) {
	var user dto.UserWithRoleDto
	user, _ = GetUserInstance(c)
	problemId, _ := strconv.Atoi(c.Param("id"))
	// 查找提交
	type SourceCode struct {
		Source   string `json:"source"`
		Language int    `json:"language"`
	}
	var sourceCode SourceCode
	ORM.Raw(`select source_code.source,solution.language from solution inner join source_code 
	on source_code.solution_id = solution.solution_id where solution.problem_id = ? and solution.user_id = ? order by solution.in_date desc limit 1`, problemId, user.ID).
		Scan(&sourceCode)
	c.JSON(http.StatusOK, gin.H{
		"message":    "获取最近提交信息成功",
		"sourcecode": sourceCode,
	})
}

// 获取最近一次比赛中提交的代码
func GetLatestContestSource(c *gin.Context) {
	var user dto.UserWithRoleDto

	user, _ = GetUserInstance(c)
	contestId, _ := strconv.Atoi(c.Param("id"))
	num, _ := strconv.Atoi(c.Param("num"))

	// 查找提交
	type SourceCode struct {
		Source   string `json:"source"`
		Language int    `json:"language"`
	}
	var sourceCode SourceCode
	ORM.Raw(`select source_code.source,solution.language from solution inner join source_code 
	on source_code.solution_id = solution.solution_id where solution.contest_id = ? and solution.num = ? and solution.user_id = ? order by solution.in_date desc limit 1`, contestId, num, user.ID).
		Scan(&sourceCode)
	c.JSON(http.StatusOK, gin.H{
		"message":    "获取最近提交信息成功",
		"sourcecode": sourceCode,
	})
}
