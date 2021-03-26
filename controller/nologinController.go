package controller

import (
	"ahpuoj/config"
	"ahpuoj/constant"
	"ahpuoj/dto"
	"ahpuoj/entity"
	"ahpuoj/utils"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// 访客获取新闻列表的接口
func NologinGetNewList(c *gin.Context) {

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))

	query := ORM.Model(entity.New{}).Where("defunct=0")

	var total int64
	query.Count(&total)
	news := []entity.New{}
	err := query.Scopes(Paginate(c)).Order("top desc,id desc").Find(&news).Error
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    page,
		"perpage": perpage,
		"data":    news,
	})
}

// 返回跳转问题的序号
func NologinJumpProblem(c *gin.Context) {
	action := c.Query("action")
	id, _ := strconv.Atoi(c.Query("problemId"))
	var newId int
	var err error
	switch action {
	case "prev":
		ORM.Raw("select id from problem where id < ? ant defunct = 0 order by id desc limit 1", id).Scan(&newId)
	case "next":
		ORM.Raw("select id from problem where id > ? and defunct = 0 order by id asc limit 1", id).Scan(&newId)
	case "random":
		ORM.Raw(`SELECT t1.id FROM problem AS t1 JOIN (SELECT ROUND(RAND() * ((SELECT MAX(id) FROM problem)-(SELECT MIN(id) FROM problem))+(SELECT MIN(id) FROM problem)) AS id) AS t2 WHERE t1.id >= t2.id and t1.defunct = 0 ORDER BY t1.id LIMIT 1`).
			Scan(&newId)
	}
	log.Println(err)
	if newId == 0 {
		newId = id
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"data":    newId,
	})
}

// 访客获取问题列表的接口
func NologinGetProblemList(c *gin.Context) {

	user, loggedIn := GetUserInstance(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))
	param := c.Query("param")
	level := c.Query("level")
	tagId := c.Query("tag_id")

	query := ORM.Model(entity.Problem{}).Preload("Tags").Joins("left join problem_tag on problem.id = problem_tag.problem_id")
	// 非管理员无法查看隐藏的问题
	if !IsAdmin(c) {
		query.Where("problem.defunct=0")
	}

	if len(param) > 0 {
		_, err := strconv.Atoi(param)
		if err != nil {
			query.Where("problem.title like ?", "%"+param+"%")
		} else {
			query.Where("problem.id = ?", param)
		}
	}
	if len(tagId) > 0 {
		query.Where("problem_tag.tag_id = ?", tagId)
	}
	if len(level) > 0 {
		query.Where("problem.level = ?", level)
	}

	query.Group("problem.id")
	query.Order("problem.id desc")
	problems := []entity.Problem{}
	var total int64
	query.Count(&total)
	err := query.Select("problem.*").Scopes(Paginate(c)).Find(&problems).Error
	if err != nil {
		panic(err)
	}

	results := problemService.ConvertList(problems, 0, user, loggedIn)

	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    page,
		"perpage": perpage,
		"data":    results,
	})
}

// 访客获取竞赛列表的接口
func NologinGetContestList(c *gin.Context) {

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))
	param := c.Query("param")

	query := ORM.Model(entity.Contest{})
	// 非管理员无法查看隐藏的问题
	if !IsAdmin(c) {
		query.Where("defunct=0")
	}

	if len(param) > 0 {
		query.Where("name like ?", "%"+param+"%")
	}

	query.Order("id desc")
	var total int64
	query.Count(&total)
	contests := []entity.Contest{}
	err := query.Scopes(Paginate(c)).Find(&contests).Error

	if err != nil {
		panic(err)
	}
	results := []dto.ContestInfoDto{}

	for _, contest := range contests {
		results = append(results, dto.ContestInfoDto{
			Contest:   contest,
			StartTime: utils.JSONDateTime(contest.StartTime),
			EndTime:   utils.JSONDateTime(contest.EndTime),
			Status:    contestService.CalcStatus(&contest),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    page,
		"perpage": perpage,
		"data":    results,
	})
}

// 访客获取评测记录列表的接口
func NologinGetSolutionList(c *gin.Context) {
	if err := ReadFromCacheByRequestURI(c); err == nil {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))
	param := c.Query("param")
	username := c.Query("username")
	language := c.Query("language")
	result := c.Query("result")
	contestId, _ := strconv.Atoi(c.Query("contest_id"))

	query := ORM.Model(entity.Solution{}).
		Joins("inner join problem on solution.problem_id=problem.id").
		Joins("inner join user on solution.user_id = user.id").
		Joins("inner join source_code on solution.solution_id=source_code.solution_id")

	if len(username) > 0 {
		query.Where(query.Where("user.username = ?", username).Or("user.nick = ?", username))
	}
	if len(language) > 0 {
		query.Where("solution.language = ?", language)
	}
	if len(result) > 0 {
		query.Where("solution.result = ?", result)
	}

	// 查询比赛中的提交
	if contestId > 0 {
		query.Where("solution.contest_id = ?", contestId)
		num, err := utils.EngNumToInt(param)
		// 搜索格式不对 直接PASS
		if err != nil {
			panic(err)
		}
		if num > 0 {
			query.Where("solution.num = ?", num)
		}
	} else {
		// 平时不显示比赛提交
		query.Where("solution.contest_id = 0")
		if len(param) > 0 {
			_, err := strconv.Atoi(param)
			if err != nil {
				query.Where("problem.title like ?", "%"+param+"%")
			} else {
				query.Where("problem.id = ? ", param)
			}
		}
	}
	query.Order("solution.solution_id desc")
	var total int64
	query.Count(&total)
	solutions := []dto.SolutionInfoDto{}
	err := query.Scopes(Paginate(c)).Select("solution.*", "user.username", "user.nick", "user.avatar", "problem.title problem_title", "source_code.public").Find(&solutions).Error
	if err != nil {
		panic(err)
	}

	resposneData := gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    page,
		"perpage": perpage,
		"data":    solutions,
	}
	serializedData, _ := json.Marshal(resposneData)
	StoreToCacheByRequestURI(c, serializedData, 5)
	c.JSON(http.StatusOK, resposneData)
}

// 获取评测记录信息的接口
func NologinGetSolution(c *gin.Context) {
	var user dto.UserWithRoleDto
	user, loggedIn := GetUserInstance(c)

	var solution dto.SolutionInfoDto
	id, _ := strconv.Atoi(c.Param("id"))
	var err error

	err = ORM.Model(entity.Solution{}).Joins("inner join problem on solution.problem_id=problem.id").
		Joins("inner join user on solution.user_id = user.id").
		Joins("inner join source_code on solution.solution_id=source_code.solution_id").
		Select("solution.*", "user.username", "user.nick", "user.avatar", "problem.title problem_title", "source_code.public").
		Where("solution.solution_id", id).
		Find(&solution).Error
	if err != nil {
		panic(err)
	}

	seeable := false

	// 代码是否可以查看
	if loggedIn && user.Role != "user" {
		seeable = true
	} else {
		// 自己的代码可以查看
		if loggedIn && solution.UserId == user.ID {
			seeable = true
		}
		// 非比赛中可以查看公开的代码
		if solution.ContestId == 0 {
			if solution.Public == 1 {
				seeable = true
			}
		}
	}

	var runtimeInfo entity.RuntimeInfo
	var compileInfo entity.CompileInfo
	var sourceCode entity.SourceCode

	if seeable {
		ORM.Model(entity.SourceCode{}).Where("solution_id = ?", solution.ID).Find(&sourceCode)
	}
	// 当 result 为 WA TL ML PE  CE时，返回运行时错误信息
	if (solution.Result >= 5 && solution.Result <= 8) || solution.Result == 10 {
		ORM.Model(entity.RuntimeInfo{}).Where("solution_id = ?", solution.ID).Find(&runtimeInfo)
	}
	if solution.Result == 11 {
		ORM.Model(entity.CompileInfo{}).Where("solution_id = ?", solution.ID).Find(&compileInfo)
	}

	meta := make(map[string]interface{}, 0)
	meta["runtime_info"] = runtimeInfo.Error
	meta["compile_info"] = compileInfo.Error
	meta["source"] = sourceCode.Source

	c.JSON(http.StatusOK, gin.H{
		"message":  "数据获取成功",
		"solution": solution,
		"meta":     meta,
		"seeable":  seeable,
	})
}

// 获取全部标签的接口
func NologinGetAllTags(c *gin.Context) {
	tags := []entity.Tag{}
	err := ORM.Model(entity.Tag{}).Order("id desc").Find(&tags).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"tags":    tags,
	})
}

// 获取问题的接口
func NologinGetProblem(c *gin.Context) {
	var user dto.UserWithRoleDto
	user, loggedIn := GetUserInstance(c)

	var problem entity.Problem
	id, _ := strconv.Atoi(c.Param("id"))

	// 查询缓存
	conn := REDIS.Get()
	defer conn.Close()
	if cache, err := redis.Bytes(conn.Do("get", "problem:"+c.Param("id"))); err == nil {
		var jsonData map[string]interface{}
		json.Unmarshal(cache, &jsonData)
		// 非管理员 不能查看隐藏的问题
		if jsonData["defunct"].(float64) == 1 && !(loggedIn && user.Role != "user") {
			err = errors.New("权限不足")
		}
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "数据获取成功",
			"problem": jsonData,
		})
	} else {
		err := ORM.Model(entity.Problem{}).Preload("Tags").Find(&problem, id).Error
		// 查询成功 但是用户没有权限查看该题目
		if err == nil && problem.Defunct == 1 && !(loggedIn && user.Role != "user") {
			err = errors.New("权限不足")
		}
		// 缓存到 redis
		if stringify, err := json.Marshal(problem); err == nil {
			conn.Do("set", "problem:"+strconv.Itoa(problem.ID), stringify)
			conn.Do("expire", "problem:"+strconv.Itoa(problem.ID), RedisCacheLiveTime)
		}
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "数据获取成功",
			"problem": problem,
		})
	}
}

// 获取竞赛作业问题信息的接口
func NologinGetContestProblem(c *gin.Context) {
	var err error
	var user dto.UserWithRoleDto
	user, loggedIn := GetUserInstance(c)
	cid, _ := strconv.Atoi(c.Param("id"))
	num, _ := strconv.Atoi(c.Param("num"))
	var contest entity.Contest

	// 非管理员
	query := ORM.Model(entity.Contest{})
	if !(loggedIn && user.Role != "user") {
		query.Where("defunct = 0")
	}
	err = query.First(&contest, cid).Error
	if err != nil {
		panic(err)
	}

	var problem entity.Problem
	contestStatus := contestService.CalcStatus(&contest)
	seeable := true
	reason := ""

	if loggedIn {
		// 不是管理员
		if user.Role == "user" {
			// 如果竞赛作业尚未开始，题目不可见
			if contestStatus == constant.CONTEST_NOT_START {
				seeable = false
				reason = "竞赛尚未开始，题目不可见"
			} else if contestStatus == constant.CONTEST_FINISH { // 如果竞赛作业已经结束，题目可见
			} else { // 否则判断竞赛是否私有，私有判断是否有权限
				if contest.Private == 1 {
					var temp int
					ORM.Raw("select count(1) from contest_user where contest_user.contest_id = ? and contest_user.user_id = ?", contest.ID, user.ID).Scan(&temp)
					if temp == 0 {
						seeable = false
						reason = "对不起，你没有参加此次竞赛的权限"
					}
				}
			}
		}
	} else { // 游客可以查看已经结束的竞赛的题目
		if contestStatus == constant.CONTEST_NOT_START {
			seeable = false
			reason = "竞赛尚未开始，题目不可见"
		} else if contestStatus == constant.CONTEST_FINISH { // 如果竞赛作业已经结束，题目可见
		} else {
			if contest.Private == 1 { // 私有的竞赛作业无法查看
				seeable = false
				reason = "对不起，你没有参加此次竞赛的权限"
			}
		}
	}

	if !seeable {
		panic(errors.New(reason))
	}

	err = ORM.Model(entity.Problem{}).Joins("inner join contest_problem on contest_problem.problem_id = problem.id").Where("contest_problem.contest_id= ?", cid).
		Where("contest_problem.num = ?", num).First(&problem).Error
	if err != nil {
		panic(err)
	}
	var contestSubmit, contestAccepted int
	// 处理提交和通过 只显示竞赛作业中的提交和通过 单人通过多次只算一次
	ORM.Raw(`select count(1) from solution where contest_id =  ? and num = ?`, cid, num).Scan(&contestSubmit)
	ORM.Raw(`select count(1)  from (select count(1) from solution where contest_id =  ? and num = ? and result = 4 group by user_id) T`, cid, num).Scan(&contestAccepted)
	problem.Submit = contestSubmit
	problem.Accepted = contestAccepted

	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"seeable": seeable,
		"problem": problem,
	})
}

// 获取竞赛信息的接口
func NologinGetContest(c *gin.Context) {
	var err error
	var user dto.UserWithRoleDto
	user, loggedIn := GetUserInstance(c)
	var contest entity.Contest
	id, _ := strconv.Atoi(c.Param("id"))

	// 非管理员无法查看被保留的竞赛作业
	err = ORM.Model(entity.Contest{}).Preload("Problems").Find(&contest, id).Error
	if err == nil && contest.Defunct == 1 && !(loggedIn && user.Role != "user") {
		err = errors.New("权限不足")
	}
	if err != nil {
		panic(err)
	}

	contestInfo := dto.ContestInfoDto{
		Contest:   contest,
		StartTime: utils.JSONDateTime(contest.StartTime),
		EndTime:   utils.JSONDateTime(contest.EndTime),
	}
	contestInfo.Status = contestService.CalcStatus(&contest)
	seeable := true
	reason := ""

	if loggedIn {
		// 不是管理员
		if user.Role == "user" {
			// 如果竞赛作业尚未开始，题目不可见
			if contestInfo.Status == 1 {
				seeable = false
				reason = "竞赛尚未开始，题目不可见"
			} else if contestInfo.Status == 3 { // 如果竞赛作业已经结束，题目可见
			} else { // 否则判断竞赛是否私有，私有判断是否有权限
				if contestInfo.Private == 1 {
					var temp int64
					ORM.Model(entity.ContestUser{}).Where("contest_id = ?", contestInfo.ID).Where("user_id = ?", user.ID).Count(&temp)
					if temp == 0 {
						seeable = false
						reason = "对不起，你没有参加此次竞赛的权限"
					}
				}
			}
		}
	} else { // 游客可以查看已经结束的竞赛的题目列表
		if contestInfo.Status == 1 {
			seeable = false
			reason = "竞赛尚未开始，题目不可见"
		} else if contestInfo.Status == 3 { // 如果竞赛作业已经结束，题目可见
		} else {
			if contestInfo.Private == 1 { // 私有的竞赛作业无法查看
				seeable = false
				reason = "对不起，你没有参加此次竞赛的权限"
			}
		}
	}
	problemList := problemService.ConvertList(contest.Problems, contest.ID, user, loggedIn)

	contest.Problems = nil
	result := struct {
		dto.ContestInfoDto
		ProblemInfos []dto.ProblemListItemDto `json:"probleminfos"`
	}{
		ContestInfoDto: contestInfo,
		ProblemInfos:   problemList,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"seeable": seeable,
		"reason":  reason,
		"contest": result,
	})
}

// 获取竞赛作业排名的接口
func NologinGetContestRankList(c *gin.Context) {
	var user dto.UserWithRoleDto
	user, loggedIn := GetUserInstance(c)
	var contest entity.Contest
	id, _ := strconv.Atoi(c.Param("id"))

	err := ORM.Model(entity.Contest{}).First(&contest, id).Error
	if err != nil {
		panic(err)
	}

	contestStatus := contestService.CalcStatus(&contest)
	seeable := true
	reason := ""

	if loggedIn {
		// 不是管理员
		if user.Role == "user" {
			// 如果竞赛作业尚未开始，排名不可见
			if contestStatus == constant.CONTEST_NOT_START {
				seeable = false
				reason = "竞赛尚未开始，排名不可见"
			} else if contestStatus == constant.CONTEST_FINISH { // 如果竞赛作业已经结束，排名可见
			} else { // 否则判断竞赛是否私有，私有判断是否有权限
				if contest.Private == 1 {
					var temp int
					ORM.Raw("select count(1) from contest_user where contest_user.contest_id = ? and contest_user.user_id = ?", contest.ID, user.ID).Scan(&temp)
					if temp == 0 {
						seeable = false
						reason = "对不起，你没有参加此次竞赛的权限"
					}
				}
			}
		}
	} else { // 游客可以查看已经结束的竞赛的题目列表
		if contestStatus == constant.CONTEST_NOT_START {
			seeable = false
			reason = "竞赛尚未开始，排名不可见"
		} else if contestStatus == constant.CONTEST_FINISH { // 如果竞赛作业已经结束，题目可见
		} else {
			if contest.Private == 1 { // 私有的竞赛作业无法查看
				seeable = false
				reason = "对不起，你没有参加此次竞赛的权限"
			}
		}
	}

	userRankInfoList := dto.UserRankInfoList{}
	var problemCount int64
	if seeable {
		// 获得竞赛作业题目总数
		ORM.Model(entity.ContestProblem{}).Where("contest_id = ?", id).Count(&problemCount)

		rankItems := []dto.RankItem{}
		ORM.Raw(`select s.problem_id,s.team_id,s.user_id,s.contest_id,s.num,s.in_date,s.result,u.username,u.nick,u.avatar as user_avatar,r.name as user_role from
		solution s inner join user u on s.user_id = u.id
		inner join role r on u.role_id = r.id
		where s.contest_id = ? order by s.user_id, s.in_date asc`, id).Find(&rankItems)

		lastUserId := 0
		var userRankInfo dto.UserRankInfo

		for _, rankItem := range rankItems {
			// 忽略管理员的提交
			if rankItem.UserRole != "user" {
				continue
			}
			// 如果是新的用户的数据
			if rankItem.UserId != lastUserId {
				if userRankInfo.User.Id != 0 {
					userRankInfoList = append(userRankInfoList, userRankInfo)
				}
				userRankInfo = dto.UserRankInfo{
					Solved:  0,
					Time:    0,
					WaCount: make([]int, problemCount),
					AcTime:  make([]int, problemCount),
					AcFlag:  make([]bool, problemCount),
					User: struct {
						Id       int    `json:"id"`
						Username string `json:"username"`
						Nick     string `json:"nick"`
					}{
						Id:       rankItem.UserId,
						Username: rankItem.Username,
						Nick:     rankItem.Nick,
					},
				}
			}
			userRankInfo.Add(rankItem, contest.StartTime)
			lastUserId = rankItem.UserId
		}
		if userRankInfo.User.Id != 0 {
			userRankInfoList = append(userRankInfoList, userRankInfo)
		}
	}
	sort.Sort(userRankInfoList)
	c.JSON(http.StatusOK, gin.H{
		"message":  "数据获取成功",
		"seeable":  seeable,
		"reason":   reason,
		"ranklist": userRankInfoList,
		"contest": struct {
			ProblemCount int64  `json:"problem_count"`
			Name         string `json:"name"`
			Id           int    `json:"id"`
		}{
			ProblemCount: problemCount,
			Name:         contest.Name,
			Id:           contest.ID,
		},
	})
}

// 获取竞赛作业团队排名的接口
func NologinGetContestTeamRankList(c *gin.Context) {
	var user dto.UserWithRoleDto
	user, loggedIn := GetUserInstance(c)
	var contest entity.Contest
	id, _ := strconv.Atoi(c.Param("id"))
	err := ORM.Model(entity.Contest{}).First(&contest, id).Error
	if err != nil {
		panic(err)
	}
	if contest.TeamMode != 1 {
		panic(errors.New("竞赛&作业不是团队模式"))
	}

	contestStatus := contestService.CalcStatus(&contest)
	seeable := true
	reason := ""

	if loggedIn {
		// 不是管理员
		if user.Role == "user" {
			// 如果竞赛作业尚未开始，排名不可见
			if contestStatus == constant.CONTEST_NOT_START {
				seeable = false
				reason = "竞赛尚未开始，题目不可见"
			} else if contestStatus == constant.CONTEST_FINISH { // 如果竞赛作业已经结束，排名可见
			} else { // 否则判断竞赛是否私有，私有判断是否有权限
				if contest.Private == 1 {
					var temp int
					ORM.Raw("select count(1) from contest_user where contest_user.contest_id = ? and contest_user.user_id = ?", contest.ID, user.ID).Scan(&temp)
					if temp == 0 {
						seeable = false
						reason = "对不起，你没有参加此次竞赛的权限"
					}
				}
			}
		}
	} else { // 游客可以查看已经结束的竞赛的题目列表
		if contestStatus == constant.CONTEST_NOT_START {
			seeable = false
			reason = "竞赛尚未开始，排名不可见"
		} else if contestStatus == constant.CONTEST_FINISH { // 如果竞赛作业已经结束，题目可见
		} else {
			if contest.Private == 1 { // 私有的竞赛作业无法查看
				seeable = false
				reason = "对不起，你没有参加此次竞赛的权限"
			}
		}
	}

	// 按照team_id来排序
	var userRankInfoList dto.UserRankSortByTeam
	var problemCount int
	if seeable {
		// 获得竞赛作业题目总数
		ORM.Raw("select count(1) from contest_problem where contest_id = ?", id).Scan(&problemCount)
		rankItems := []dto.RankItem{}
		ORM.Raw(`select s.problem_id,s.team_id,s.user_id,s.contest_id,s.num,s.in_date,s.result,u.username,u.nick,u.avatar,r.name from
		solution s inner join user u on s.user_id = u.id inner join role r on u.role_id = r.id where s.contest_id = ? order by s.user_id, s.in_date asc`, id).Find(&rankItems)

		lastUserId := 0
		var userRankInfo dto.UserRankInfo

		for _, rankItem := range rankItems {
			// 忽略管理员的提交
			if rankItem.UserRole != "user" {
				continue
			}
			// 如果是新的用户的数据
			if rankItem.UserId != lastUserId {
				if userRankInfo.User.Id != 0 {
					userRankInfoList = append(userRankInfoList, userRankInfo)
				}
				userRankInfo = dto.UserRankInfo{
					Solved:  0,
					Time:    0,
					WaCount: make([]int, problemCount),
					AcTime:  make([]int, problemCount),
					TeamId:  rankItem.TeamId,
					User: struct {
						Id       int    `json:"id"`
						Username string `json:"username"`
						Nick     string `json:"nick"`
					}{
						Id:       rankItem.UserId,
						Username: rankItem.Username,
						Nick:     rankItem.Nick,
					},
				}
			}
			userRankInfo.TeamId = rankItem.TeamId
			userRankInfo.Add(rankItem, contest.StartTime)
			lastUserId = rankItem.UserId
		}
		userRankInfoList = append(userRankInfoList, userRankInfo)
	}
	sort.Sort(userRankInfoList)

	var teamRankInfoList dto.TeamRankInfoList

	// 获取全部参赛队伍数据
	teams := []entity.Team{}
	ORM.Raw(`select team.* from
	contest_team inner join team on contest_team.team_id = team.id
	where team.is_deleted = 0 and contest_team.contest_id = ? order by team.id asc`, contest.ID).Find(&teams)
	for _, team := range teams {
		var teamRankInfo = dto.TeamRankInfo{
			Solved:  0,
			Time:    0,
			WaCount: make([]int, problemCount),
			AcTime:  make([]int, problemCount),
			AcCount: make([]int, problemCount),
			Team: struct {
				Id   int    `json:"id"`
				Name string `json:"name"`
			}{
				Id:   team.ID,
				Name: team.Name,
			},
		}
		teamRankInfoList = append(teamRankInfoList, teamRankInfo)
	}

	// team排名信息和个人信息都是按照teamid递增排列的  o(n)方式来统计

	userCount := len(userRankInfoList)
	cnt := 0

out:
	for k, v := range teamRankInfoList {
		// 如果用户信息已经统计完 break
		if cnt >= userCount {
			break
		}
		// 如果个人信息的teamid大于当前team的id continue
		if userRankInfoList[cnt].TeamId > v.Team.Id {
			continue
		}
		for userRankInfoList[cnt].TeamId == v.Team.Id {
			teamRankInfoList[k].Add(userRankInfoList[cnt])
			cnt++
			if cnt >= userCount {
				break out
			}
		}
	}
	sort.Sort(teamRankInfoList)
	c.JSON(http.StatusOK, gin.H{
		"message":      "数据获取成功",
		"seeable":      seeable,
		"reason":       reason,
		"teamranklist": teamRankInfoList,
		"contest": struct {
			ProblemCount int    `json:"problem_count"`
			Name         string `json:"name"`
			Id           int    `json:"id"`
		}{
			ProblemCount: problemCount,
			Name:         contest.Name,
			Id:           contest.ID,
		},
	})
}

// 访客获取竞赛列表的接口
func NologinGetSeriesList(c *gin.Context) {

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))
	param := c.Query("param")

	query := ORM.Model(entity.Series{}).Where("defunct=0")
	if len(param) > 0 {
		query.Where("name like ?", "%"+param+"%")
	}
	var total int64
	query.Count(&total)
	serieses := []entity.Series{}
	err := query.Scopes(Paginate(c)).Order("id desc").Find(&serieses).Error
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    page,
		"perpage": perpage,
		"data":    serieses,
	})
}

// 访客获取系列赛信息的接口
func NologinGetSeries(c *gin.Context) {
	var err error
	var series entity.Series
	id, _ := strconv.Atoi(c.Param("id"))

	err = ORM.Model(entity.Series{}).First(&series, id).Error
	if err != nil {
		panic(err)
	}
	err = ORM.Model(entity.Contest{}).Joins("inner join contest_series on contest.id =  contest_series.contest_id").Where("contest_series.series_id = ? and defunct = 0 and team_mode = ?", &series.ID, &series.TeamMode).Find(&series.Contests).Error
	if err != nil {
		panic(err)
	}

	contestIdStrList := []string{}
	for _, contest := range series.Contests {
		contestIdStrList = append(contestIdStrList, strconv.Itoa(contest.ID))
	}

	var contestCount int64
	var problemCount int64

	var userSeriesRankInfo dto.UserSeriesRankInfo
	userSeriesRankInfoList := dto.UserSeriesRankInfoList{}

	ORM.Model(entity.Contest{}).Joins("inner join contest_series on contest_series.contest_id = contest.id").Where("contest.defunct = 0").Where("series_id = ?", id).Where("team_mode = ?", series.TeamMode).Count(&contestCount)

	ORM.Model(entity.Contest{}).Joins("inner join contest_series on contest.id = contest_series.contest_id").
		Joins("inner join contest_problem on contest.id = contest_problem.contest_id").
		Where("contest.defunct = 0").Where("series_id = ?", id).Where("team_mode = ?", series.TeamMode).Count(&problemCount)

	contestIdStrListStr := strings.Join(contestIdStrList, ",")
	// 个人模式排名汇总,取得系列赛全部的提交记录
	rankItems := []dto.RankItem{}
	ORM.Raw(`select s.problem_id,s.team_id,s.user_id,s.contest_id,s.num,s.in_date,s.result,u.username,u.nick,u.avatar as user_avatar,r.name as user_role from
	solution s inner join user u on s.user_id = u.id
	inner join role r on u.role_id = r.id
	inner join contest c on s.contest_id = c.id
	where s.contest_id in (`+contestIdStrListStr+`) and c.defunct = 0 and c.team_mode = ? order by s.user_id, s.in_date asc`, series.TeamMode).Find(&rankItems)
	lastUserId := 0

	for _, rankItem := range rankItems {
		var contest entity.Contest

		// 获取当前提交的竞赛信息
		for _, c := range series.Contests {
			if rankItem.ContestId == c.ID {
				contest = c
				// break
			}
		}

		// 忽略管理员的提交
		if rankItem.UserRole == "admin" {
			continue
		}

		// 如果是新的用户的数据
		if rankItem.UserId != lastUserId {
			if userSeriesRankInfo.User.Id != 0 {
				userSeriesRankInfoList = append(userSeriesRankInfoList, userSeriesRankInfo)
			}
			userSeriesRankInfo = dto.UserSeriesRankInfo{
				Solved:  make(map[int]int, contestCount),
				Time:    make(map[int]int, contestCount),
				WaCount: make(map[int][]int, problemCount),
				AcTime:  make(map[int][]int, problemCount),
				User: struct {
					Id       int    `json:"id"`
					Username string `json:"username"`
					Nick     string `json:"nick"`
				}{
					Id:       rankItem.UserId,
					Username: rankItem.Username,
					Nick:     rankItem.Nick,
				},
			}
		}
		userSeriesRankInfo.Add(rankItem, contest.ID, contest.StartTime, problemCount)
		lastUserId = rankItem.UserId
	}
	if userSeriesRankInfo.User.Id != 0 {
		userSeriesRankInfoList = append(userSeriesRankInfoList, userSeriesRankInfo)
	}

	type Result struct {
		entity.Series
		Contestinfos []dto.ContestInfoDto `json:"contestinfos"`
	}
	result := Result{
		Series:       series,
		Contestinfos: []dto.ContestInfoDto{},
	}

	for _, contest := range series.Contests {
		contestStatus := contestService.CalcStatus(&contest)
		result.Contestinfos = append(result.Contestinfos, dto.ContestInfoDto{
			Contest:   contest,
			StartTime: utils.JSONDateTime(contest.StartTime),
			EndTime:   utils.JSONDateTime(contest.EndTime),
			Status:    contestStatus,
		})

	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "数据获取成功",
		"series":       result,
		"userranklist": userSeriesRankInfoList,
	})
}

// 获取系统可用语言列表的接口
func NologinGetLanguageList(c *gin.Context) {
	numberStr, _ := config.Conf.GetValue("language", "number")
	number, _ := strconv.Atoi(numberStr)
	langmaskStr, _ := config.Conf.GetValue("language", "mask")
	langmask, _ := strconv.Atoi(langmaskStr)
	langname, _ := config.Conf.GetValue("language", "langname")
	langNameList := strings.Split(langname, ",")
	languages := []map[string]interface{}{}
	for i := 0; i < number; i++ {
		available := false
		if (langmask & (1 << uint(i))) > 0 {
			available = true
		}
		lang := map[string]interface{}{
			"name":      langNameList[i],
			"available": available,
		}
		languages = append(languages, lang)
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "获取语言信息成功",
		"languages": languages,
	})
}

// 获取讨论列表的接口
func NologinGetIssueList(c *gin.Context) {
	// 判断当前是否已经关闭讨论版
	type EnableIssue struct {
		Value string
	}
	var enableIssue EnableIssue
	err := ORM.Table("config").Where("item = 'enable_issue'").Scan(&enableIssue).Error
	if err != nil {
		panic(err)
	}
	if enableIssue.Value == "false" {
		c.JSON(http.StatusOK, gin.H{
			"message":      "数据获取成功",
			"issue_enable": false,
		})
		return
	}

	problemId, _ := strconv.Atoi(c.Param("id"))
	user, loggedIn := GetUserInstance(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))
	// 检查问题是否存在
	if problemId != 0 {
		var problem entity.Problem
		err := ORM.Model(entity.Problem{}).First(&problem, problemId).Error
		if err != nil {
			panic(err)
		}
	}

	query := ORM.Model(entity.Issue{}).Joins("inner join user on issue.user_id = user.id").
		Joins("left join problem on issue.problem_id = problem.id")

	// problem=0时显示所有主题
	if problemId != 0 {
		query.Where("problem_id = ?", problemId)
	}
	// 管理员可以查看被删除的主题
	if loggedIn && user.Role != "user" {
		query.Unscoped()
	}
	var total int64
	query.Count(&total)
	results := []dto.IssueInfoDto{}
	err = query.Scopes(Paginate(c)).Select("user.username,user.nick,user.avatar,issue.*,problem.title problem_title,(select count(1) from reply where issue_id = issue.id) as reply_count").
		Order("issue.created_at desc").Find(&results).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message":      "数据获取成功",
		"total":        total,
		"page":         page,
		"perpage":      perpage,
		"issue_enable": true,
		"data":         results,
	})
}

// 获得讨论以及评论的接口
func NologinGetIssue(c *gin.Context) {
	// 判断当前是否已经关闭讨论版
	type EnableIssue struct {
		Value string
	}
	var enableIssue EnableIssue
	err := ORM.Table("config").Where("item = 'enable_issue'").Scan(&enableIssue).Error
	if err != nil {
		panic(err)
	}
	if enableIssue.Value == "false" {
		c.JSON(http.StatusOK, gin.H{
			"message":      "数据获取成功",
			"issue_enable": false,
		})
		return
	}
	issueId, _ := strconv.Atoi(c.Param("id"))
	user, loggedIn := GetUserInstance(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))
	var issue dto.IssueInfoDto
	// 检查讨论是否存在
	if issueId != 0 {
		err := ORM.Model(entity.Issue{}).Joins("inner join user on issue.user_id = user.id").
			Joins("left join problem on issue.problem_id = problem.id").
			Select("user.username,user.nick,user.avatar,issue.*,problem.title ptitle,(select count(1) from reply where issue_id = issue.id) as reply_count").
			Where("issue.id = ?", issueId).Find(&issue).Error
		if err != nil {
			panic(err)
		}
	}
	// 获取回复
	replys := []dto.ReplyInfoDto{}
	query := ORM.Model(entity.Reply{}).Joins("left join user on reply.user_id = user.id").
		Joins("left join user u2 on reply.reply_user_id = u2.id").
		Select("user.username,user.nick,user.avatar,reply.*,u2.nick as reply_user_nick,(select count(1) from reply  r where reply.id = r.reply_id) as reply_count").
		Where("issue_id = ?", issueId)

	// 管理员可以查看被删除的回复
	if loggedIn && user.Role != "user" {
		query.Unscoped()
	}
	err = query.Scopes(Paginate(c)).Order("reply.created_at asc").Find(&replys).Error
	if err != nil {
		panic(err)
	}
	// 对回复进行处理

	issueReplys := []dto.ReplyInfoDto{}

	for _, reply := range replys {
		if reply.ReplyId == 0 {
			issueReplys = append(issueReplys, reply)
		}
	}
	for _, reply := range replys {
		if reply.ReplyId == 0 {
			continue
		}
		for index, issueReply := range issueReplys {
			if reply.ReplyId == issueReply.ID {
				issueReplys[index].SubReplys = append(issueReply.SubReplys, reply)
			}
		}
	}
	total := len(issueReplys)

	c.JSON(http.StatusOK, gin.H{
		"message":      "数据获取成功",
		"total":        total,
		"page":         page,
		"perpage":      perpage,
		"issue":        issue,
		"replys":       issueReplys,
		"issue_enable": true,
	})
}

// 获取用户信息的接口
func NologinGetUserInfo(c *gin.Context) {
	var err error
	userId, _ := strconv.Atoi(c.Param("id"))

	var user entity.User
	// 检查用户是否存在
	err = ORM.Model(entity.User{}).Find(&user, userId).Error
	if err != nil {
		panic(err)
	}

	var solvedProblemList = []int{}
	var waProblemList = []int{}
	var unsolvedProblemList = []int{}

	type StatisticUnit struct {
		Date  utils.JSONDate `json:"date"`
		Count int            `json:"count"`
	}
	var recentSolvedStatistic = []StatisticUnit{}
	var recentSubmitStatistic = []StatisticUnit{}
	// 不统计比赛中的数据
	ORM.Model(entity.Solution{}).Where("user_id = ?", userId).Where("contest_id = 0").Where("result = ?", constant.ACCEPT).Distinct("problem_id").
		Pluck("problem_id", &solvedProblemList)

	ORM.Model(entity.Solution{}).Where("user_id = ?", userId).Where("contest_id = 0").Where("result != ?", constant.ACCEPT).Distinct("problem_id").
		Pluck("problem_id", &waProblemList)

	for _, pid := range waProblemList {
		if utils.IndexOf(solvedProblemList, pid) == -1 {
			unsolvedProblemList = append(unsolvedProblemList, pid)
		}
	}

	// 这是一段神奇的SQL 获得15天内累计通过的变化
	ORM.Raw(`select  dualdate.date,count(distinct(problem_id)) count from
	(select * from solution where user_id=? and contest_id = 0 and result = 4) s
	right join
	(select date_sub(curdate(), interval(cast(help_topic_id as signed integer)) day) date
	from mysql.help_topic
	where help_topic_id  <= 14)  dualdate
	on date(s.in_date) <= dualdate.date
	group by dualdate.date order by dualdate.date asc`, user.ID).Find(&recentSolvedStatistic)

	// 这还是一段神奇的SQL 获得15天内累计提交的变化
	ORM.Raw(`
	select  dualdate.date,count(distinct(problem_id)) count from
	(select * from solution where user_id=? and contest_id = 0) s
	right join
	(select date_sub(curdate(), interval(cast(help_topic_id as signed integer)) day) date
	from mysql.help_topic
	where help_topic_id  <= 14)  dualdate
on date(s.in_date) <= dualdate.date
	group by dualdate.date order by dualdate.date asc`, user.ID).Find(&recentSubmitStatistic)

	var rank int64
	ORM.Model(entity.User{}).Where("solved > ?", user.Solved).Or(ORM.Where("solved = ?", user.Solved).Where("submit < ?", user.Submit)).Count(&rank)

	type UserInfoDto struct {
		entity.User
		Rank                  int64           `json:"rank"`
		SolvedProblemList     []int           `json:"solved_problem_list"`
		UnsolvedProblemList   []int           `json:"unsolved_problem_list"`
		RecentSolvedStatistic []StatisticUnit `json:"recent_solved_statistic"`
		RecentSubmitStatistic []StatisticUnit `json:"recent_submit_statistic"`
	}

	userInfo := UserInfoDto{
		User:                  user,
		Rank:                  rank + 1,
		SolvedProblemList:     solvedProblemList,
		UnsolvedProblemList:   unsolvedProblemList,
		RecentSolvedStatistic: recentSolvedStatistic,
		RecentSubmitStatistic: recentSubmitStatistic,
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "获取个人信息成功",
		"userinfo": userInfo,
	})
}

// 获取排名的接口
func NologinGetRankList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "50"))
	// 不统计比赛用户的数据
	query := ORM.Model(entity.User{}).Where("is_compete_user = 0").Order("solved desc").Order("submit asc")
	results := []entity.User{}
	var total int64
	query.Count(&total)
	err := query.Scopes(Paginate(c)).Find(&results).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    page,
		"perpage": perpage,
		"data":    results,
	})
}
