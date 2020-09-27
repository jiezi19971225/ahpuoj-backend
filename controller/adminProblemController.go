package controller

import (
	"ahpuoj/config"
	"ahpuoj/entity"
	"ahpuoj/model"
	"ahpuoj/mq"
	"ahpuoj/request"
	"ahpuoj/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func IndexProblem(c *gin.Context) {

	results, total := problemService.List(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    c.DefaultQuery("page", "1"),
		"perpage": c.DefaultQuery("perpage", "20"),
		"data":    results,
	})
}

func ShowProblem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	problem := entity.Problem{}
	err := ORM.Preload("Tags").First(&problem, id).Error
	if utils.CheckError(c, err, "") != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"problem": problem,
	})
}

func StoreProblem(c *gin.Context) {
	user, _ := GetUserInstance(c)
	conn := REDIS.Get()
	defer conn.Close()

	var req request.Problem
	err := c.ShouldBindJSON(&req)
	if utils.CheckError(c, err, "请求参数错误") != nil {
		return
	}
	problem := entity.Problem{
		Title:        req.Title,
		Description:  null.StringFrom(req.Description),
		Input:        null.StringFrom(req.Input),
		Output:       null.StringFrom(req.Output),
		SampleInput:  null.StringFrom(req.SampleInput),
		SampleOutput: null.StringFrom(req.SampleOutput),
		Spj:          req.Spj,
		Level:        req.Level,
		Hint:         null.StringFrom(req.Hint),
		TimeLimit:    req.TimeLimit,
		MemoryLimit:  req.MemoryLimit,
		CreatorId:    user.Id,
	}

	err = problemService.SaveRecord(&problem)
	if utils.CheckError(c, err, "") != nil {
		return
	}

	idStr := strconv.Itoa(user.Id)
	problemIdStr := strconv.Itoa(problem.ID)

	if user.Role != "admin" {
		enforcer := model.GetCasbin()
		enforcer.AddPolicy(idStr, "/api/admin/problem/"+problemIdStr, "PUT")
		enforcer.AddPolicy(idStr, "/api/admin/problem/"+problemIdStr, "DELETE")
		enforcer.AddPolicy(idStr, "/api/admin/problem/"+problemIdStr+"/status", "PUT")
		enforcer.AddPolicy(idStr, "/api/admin/problem/"+problemIdStr+"/data", "POST")
		enforcer.AddPolicy(idStr, "/api/admin/problem/"+problemIdStr+"/datafile", "POST")
		enforcer.AddPolicy(idStr, "/api/admin/problem/"+problemIdStr+"/data/:filename", "PUT")
		enforcer.AddPolicy(idStr, "/api/admin/problem/"+problemIdStr+"/data/:filename", "DELETE")
	}
	problemService.AddTags(&problem, req.Tags)
	// 同步到 redis 缓存
	if stringify, err := json.Marshal(problem); err == nil {
		conn.Do("set", "problem:"+strconv.Itoa(problem.ID), stringify)
		conn.Do("expire", "problem:"+strconv.Itoa(problem.ID), RedisCacheLiveTime)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "新建问题成功",
		"show":    true,
	})
}

func UpdateProblem(c *gin.Context) {
	conn := REDIS.Get()
	defer conn.Close()

	id, _ := strconv.Atoi(c.Param("id"))
	var req request.Problem
	err := c.ShouldBindJSON(&req)
	if utils.CheckError(c, err, "请求参数错误") != nil {
		return
	}
	problem := entity.Problem{
		ID:           id,
		Title:        req.Title,
		Description:  null.StringFrom(req.Description),
		Input:        null.StringFrom(req.Input),
		Output:       null.StringFrom(req.Output),
		SampleInput:  null.StringFrom(req.SampleInput),
		SampleOutput: null.StringFrom(req.SampleOutput),
		Spj:          req.Spj,
		Level:        req.Level,
		Hint:         null.StringFrom(req.Hint),
		TimeLimit:    req.TimeLimit,
		MemoryLimit:  req.MemoryLimit,
	}
	// 首先清除当前标签
	err = ORM.Model(&problem).Updates(problem).Error
	if utils.CheckError(c, err, "") != nil {
		return
	}

	problemService.ReplaceTags(&problem, req.Tags)
	// 同步到 redis 缓存
	if stringify, err := json.Marshal(problem); err == nil {
		conn.Do("set", "problem:"+strconv.Itoa(problem.ID), stringify)
		conn.Do("expire", "problem:"+strconv.Itoa(problem.ID), RedisCacheLiveTime)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "编辑问题成功",
		"show":    true,
		"problem": problem,
	})
}

func DeleteProblem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, _ := GetUserInstance(c)
	problem := entity.Problem{
		ID: id,
	}
	problemService.DeleteRecord(&problem)

	idStr := strconv.Itoa(user.Id)
	problemIdStr := strconv.Itoa(problem.ID)
	enforcer := model.GetCasbin()
	enforcer.RemovePolicy(idStr, "/api/admin/problem/"+problemIdStr, "PUT")
	enforcer.RemovePolicy(idStr, "/api/admin/problem/"+problemIdStr, "DELETE")
	enforcer.RemovePolicy(idStr, "/api/admin/problem/"+problemIdStr+"/status", "PUT")
	enforcer.RemovePolicy(idStr, "/api/admin/problem/"+problemIdStr+"/data", "POST")
	enforcer.RemovePolicy(idStr, "/api/admin/problem/"+problemIdStr+"/datafile", "POST")
	enforcer.RemovePolicy(idStr, "/api/admin/problem/"+problemIdStr+"/data/:filename", "PUT")
	enforcer.RemovePolicy(idStr, "/api/admin/problem/"+problemIdStr+"/data/:filename", "DELETE")
	var maxId int
	// 更新自增起始ID
	ORM.Model(entity.Problem{}).Select("max(id)").Scan(&maxId)
	newAutoIncrement := strconv.Itoa(maxId + 1)
	ORM.Exec("alter table problem auto_increment=" + newAutoIncrement)

	c.JSON(http.StatusOK, gin.H{
		"message": "删除问题成功",
		"show":    true,
	})
}

func ToggleProblemStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	problem := entity.Problem{
		ID: id,
	}
	ORM.Model(&problem).Update("defunct", gorm.Expr("not defunct"))
	c.JSON(http.StatusOK, gin.H{
		"message": "更改问题状态成功",
		"show":    true,
	})
}

// 重判问题相关
func RejudgeSolution(c *gin.Context) {
	var err error
	id, _ := strconv.Atoi(c.Param("id"))
	//	// 判断提交是否存在

	type RejudgeInfo struct {
		UserId      int    `db:"user_id"`
		SolutionId  int    `db:"solution_id"`
		ProblemId   int    `db:"problem_id"`
		Language    int    `db:"language"`
		TimeLimit   int    `db:"time_limit"`
		MemoryLimit int    `db:"memory_limit"`
		Source      string `db:"source"`
	}
	var info RejudgeInfo
	err = DB.Get(&info, `select solution.solution_id,solution.user_id,solution.problem_id,solution.language,problem.time_limit,problem.memory_limit,source_code.source from solution 
	inner join problem on solution.problem_id=problem.id 
	inner join source_code on solution.solution_id=source_code.solution_id 
	where solution.solution_id = ?`, id)
	if utils.CheckError(c, err, "重判提交失败，该提交不存在") != nil {
		return
	}
	//更改提交状态
	DB.Exec("update solution set result = 1 where solution_id = ?", id)
	//将判题数据推入消息队列

	jsondata, _ := json.Marshal(gin.H{
		"UserId":       info.UserId,
		"TestrunCount": 0,
		"SolutionId":   info.SolutionId,
		"ProblemId":    info.ProblemId,
		"Language":     info.Language,
		"TimeLimit":    1,
		"MemoryLimit":  64,
		"Source":       info.Source,
		"InputText":    "",
	})
	mq.Publish("oj", "problem", jsondata)
	c.JSON(http.StatusOK, gin.H{
		"message": "重判提交成功",
		"show":    true,
	})
}

func RejudgeProblem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var temp int
	// 判断提交是否存在
	DB.Get(&temp, "select count(1) from problem where id = ?", id)
	if temp == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "重判问题失败，该问题不存在",
		})
		return
	}
	DB.Exec("update solution set result = 1 where problem_id = ?", id)
	type RejudgeInfo struct {
		UserId      int    `db:"user_id"`
		SolutionId  int    `db:"solution_id"`
		ProblemId   int    `db:"problem_id"`
		Language    int    `db:"language"`
		TimeLimit   int    `db:"time_limit"`
		MemoryLimit int    `db:"memory_limit"`
		Source      string `db:"source"`
	}
	var infos []RejudgeInfo
	DB.Select(&infos, `select solution.solution_id,solution.user_id,solution.problem_id,solution.language,problem.time_limit,problem.memory_limit,source_code.source from solution 
	inner join problem on solution.problem_id=problem.id 
	inner join source_code on solution.solution_id=source_code.solution_id 
	where solution.problem_id = ?`, id)

	for _, info := range infos {
		jsondata, _ := json.Marshal(gin.H{
			"UserId":       info.UserId,
			"TestrunCount": 0,
			"SolutionId":   info.SolutionId,
			"ProblemId":    info.ProblemId,
			"Language":     info.Language,
			"TimeLimit":    1,
			"MemoryLimit":  64,
			"Source":       info.Source,
			"InputText":    "",
		})
		mq.Publish("oj", "problem", jsondata)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "重判问题成功",
		"show":    true,
	})
}

// 重排问题
func ReassignProblem(c *gin.Context) {
	// 判断原ID问题和新ID问题是否存在
	var err error
	var temp, maxId int
	oldId, _ := strconv.Atoi(c.Param("id"))
	newId, _ := strconv.Atoi(c.Param("newid"))

	DB.Get(&temp, "select count(1) from problem where id = ?", oldId)
	// 原问题不存在
	if temp == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "重排问题失败，原问题不存在",
		})
		return
	}
	DB.Get(&temp, "select count(1) from problem where id = ?", newId)
	// 新ID已有问题
	if temp > 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "重排问题失败，新问题ID已有问题",
		})
		return
	}
	// 更改问题ID
	DB.Exec("update problem set id = ? where id = ?", newId, oldId)
	// 更新solution表
	DB.Exec("update solution set problem_id = ? where problem_id = ?", newId, oldId)
	// 更新tag关联表
	DB.Exec("update problem_tag set problem_id = ? where problem_id = ?", newId, oldId)
	// 更新issue表
	DB.Exec("update issue set problem_id = ? where problem_id = ?", newId, oldId)
	// 更新自增起始ID
	DB.Get(&maxId, "select max(id) from problem")
	newAutoIncrement := strconv.Itoa(maxId + 1)
	DB.Exec("alter table problem auto_increment=" + newAutoIncrement)
	// 移动文件夹
	dataDir, _ := config.Conf.GetValue("project", "datadir")
	oldDir := dataDir + "/" + strconv.Itoa(oldId)
	newDir := dataDir + "/" + strconv.Itoa(newId)
	err = os.Rename(oldDir, newDir)

	if utils.CheckError(c, err, "移动文件夹失败，请检查文件服务器文件权限设置") != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "重排问题成功",
		"show":    true,
	})
}

func IndexProblemData(c *gin.Context) {
	var err error
	fileInfos := []map[string]interface{}{}

	id, _ := strconv.Atoi(c.Param("id"))
	dataDir, _ := config.Conf.GetValue("project", "datadir")
	baseDir := dataDir + "/" + strconv.FormatInt(int64(id), 10)
	// 如果目录不存在 则创建之
	if isExist, _ := utils.PathExists(baseDir); isExist == false {
		err = os.MkdirAll(baseDir, 0777)
	}
	files, err := ioutil.ReadDir(baseDir)
	for _, file := range files {
		utils.Consolelog(file.Name())
		fileInfos = append(fileInfos, map[string]interface{}{
			"filename": file.Name(),
			"size":     file.Size(),
			"mod_time": file.ModTime().Format("2006/1/2 15:04:05"),
		})
	}

	if utils.CheckError(c, err, "获取数据目录信息失败，请检查权限设置") != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取数据文件列表成功",
		"files":   fileInfos,
	})
}

func AddProblemData(c *gin.Context) {
	var err error
	var req struct {
		FileName string `json:"filename" binding:"required,max=20"`
	}
	id, _ := strconv.Atoi(c.Param("id"))
	err = c.ShouldBindJSON(&req)
	if utils.CheckError(c, err, "请求参数错误") != nil {
		return
	}
	dataDir, _ := config.Conf.GetValue("project", "datadir")
	baseDir := dataDir + "/" + strconv.FormatInt(int64(id), 10)
	inFileName := baseDir + "/" + req.FileName + ".in"
	outFileName := baseDir + "/" + req.FileName + ".out"
	infos := []string{}
	_, err = os.Open(inFileName)
	if os.IsNotExist(err) {
		_, err = os.Create(inFileName)
		infos = append(infos, "文件"+req.FileName+".in"+"创建成功")
	} else {
		infos = append(infos, "文件"+req.FileName+".in"+"已经存在")
	}
	_, err = os.Open(outFileName)
	if os.IsNotExist(err) {
		_, err = os.Create(outFileName)
		infos = append(infos, "文件"+req.FileName+".out"+"创建成功")
	} else {
		infos = append(infos, "文件"+req.FileName+".out"+"已经存在")
	}

	if utils.CheckError(c, err, "创建文件失败，请检查权限设置") != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "操作成功",
		"show":    true,
		"info":    infos,
	})
}

func AddProblemDataFile(c *gin.Context) {
	var err error
	id, _ := strconv.Atoi(c.Param("id"))
	filehead, err := c.FormFile("file")
	file, _ := filehead.Open()
	if utils.CheckError(c, err, "文件上传失败") != nil {
		return
	}
	dataDir, _ := config.Conf.GetValue("project", "datadir")
	baseDir := dataDir + "/" + strconv.FormatInt(int64(id), 10)
	filePath := baseDir + "/" + filehead.Filename
	outFile, _ := os.Create(filePath)
	_, err = io.Copy(outFile, file)

	if utils.CheckError(c, err, "保存数据文件失败，请检查权限设置") != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "文件上传成功",
		"show":    true,
	})
}

func GetProblemData(c *gin.Context) {
	var err error
	id, _ := strconv.Atoi(c.Param("id"))
	filename := c.Param("filename")

	dataDir, _ := config.Conf.GetValue("project", "datadir")
	baseDir := dataDir + "/" + strconv.FormatInt(int64(id), 10)
	filepath := baseDir + "/" + filename
	content, err := ioutil.ReadFile(filepath)

	if utils.CheckError(c, err, "读取数据文件失败") != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "读取数据文件成功",
		"content": string(content),
	})
}

func DownloadProblemData(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	filename := c.Param("filename")
	dataDir, _ := config.Conf.GetValue("project", "datadir")
	baseDir := dataDir + "/" + strconv.FormatInt(int64(id), 10)
	filepath := baseDir + "/" + filename
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s.txt", filename))
	c.Writer.Header().Add("Content-Type", "application/octet-stream")
	c.File(filepath)
}

func EditProblemData(c *gin.Context) {
	var err error
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	err = c.ShouldBindJSON(&req)
	if utils.CheckError(c, err, "请求参数错误") != nil {
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	filename := c.Param("filename")

	dataDir, _ := config.Conf.GetValue("project", "datadir")
	baseDir := dataDir + "/" + strconv.FormatInt(int64(id), 10)
	filepath := baseDir + "/" + filename
	err = ioutil.WriteFile(filepath, []byte(req.Content), 0755)

	if utils.CheckError(c, err, "写入数据文件失败") != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "写入数据文件成功",
		"show":    true,
	})

}

func DeleteProblemData(c *gin.Context) {
	var err error
	id, _ := strconv.Atoi(c.Param("id"))
	filename := c.Param("filename")

	dataDir, _ := config.Conf.GetValue("project", "datadir")
	baseDir := dataDir + "/" + strconv.FormatInt(int64(id), 10)
	filepath := baseDir + "/" + filename
	err = os.Remove(filepath)

	if utils.CheckError(c, err, "删除文件失败") != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除文件成功",
		"show":    true,
	})
}
