package controller

import (
	"ahpuoj/config"
	"ahpuoj/entity"
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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perpage, _ := strconv.Atoi(c.DefaultQuery("perpage", "20"))
	c.JSON(http.StatusOK, gin.H{
		"message": "数据获取成功",
		"total":   total,
		"page":    page,
		"perpage": perpage,
		"data":    results,
	})
}

func ShowProblem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	problem := entity.Problem{}
	err := ORM.Preload("Tags").First(&problem, id).Error
	if err != nil {
		panic(err)
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
		Description:  utils.RelativeNullString(null.StringFrom(req.Description)),
		Input:        utils.RelativeNullString(null.StringFrom(req.Input)),
		Output:       utils.RelativeNullString(null.StringFrom(req.Output)),
		SampleInput:  null.StringFrom(req.SampleInput),
		SampleOutput: null.StringFrom(req.SampleOutput),
		Spj:          req.Spj,
		Level:        req.Level,
		Hint:         utils.RelativeNullString(null.StringFrom(req.Hint)),
		TimeLimit:    req.TimeLimit,
		MemoryLimit:  req.MemoryLimit,
		CreatorId:    user.ID,
	}

	problemService.SaveRecord(&problem)

	idStr := strconv.Itoa(user.ID)
	problemIdStr := strconv.Itoa(problem.ID)

	if user.Role != "admin" {
		enforcer := entity.GetCasbin()
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
		"problem": problem,
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
		Description:  utils.RelativeNullString(null.StringFrom(req.Description)),
		Input:        utils.RelativeNullString(null.StringFrom(req.Input)),
		Output:       utils.RelativeNullString(null.StringFrom(req.Output)),
		SampleInput:  null.StringFrom(req.SampleInput),
		SampleOutput: null.StringFrom(req.SampleOutput),
		Spj:          req.Spj,
		Level:        req.Level,
		Hint:         utils.RelativeNullString(null.StringFrom(req.Hint)),
		TimeLimit:    req.TimeLimit,
		MemoryLimit:  req.MemoryLimit,
	}
	err = ORM.Select("id", "title", "description", "input", "output", "sample_input", "sample_output", "spj", "level", "hint", "time_limit", "memory_limit").Model(&problem).Updates(problem).Error
	if err != nil {
		panic(err)
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

	idStr := strconv.Itoa(user.ID)
	problemIdStr := strconv.Itoa(problem.ID)
	enforcer := entity.GetCasbin()
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
	err := ORM.Model(&problem).Update("defunct", gorm.Expr("not defunct")).Error
	if err != nil {
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "更改问题状态成功",
		"show":    true,
	})
}

// 重判问题相关
func RejudgeSolution(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	problemService.RejudgeSolution(id)
	c.JSON(http.StatusOK, gin.H{
		"message": "重判提交成功",
		"show":    true,
	})
}

func RejudgeProblem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	problemService.RejudgeProblem(id)
	c.JSON(http.StatusOK, gin.H{
		"message": "重判问题成功",
		"show":    true,
	})
}

// 重排问题
func ReassignProblem(c *gin.Context) {
	// 判断原ID问题和新ID问题是否存在
	oldId, _ := strconv.Atoi(c.Param("id"))
	newId, _ := strconv.Atoi(c.Param("newid"))
	problemService.Reassignment(oldId, newId)
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
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fileInfos = append(fileInfos, map[string]interface{}{
			"filename": file.Name(),
			"size":     file.Size(),
			"mod_time": file.ModTime().Format("2006/1/2 15:04:05"),
		})
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
	if err != nil {
		panic(err)
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
	if err != nil {
		panic(err)
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
	if err != nil {
		panic(err)
	}
	file, _ := filehead.Open()
	dataDir, _ := config.Conf.GetValue("project", "datadir")
	baseDir := dataDir + "/" + strconv.FormatInt(int64(id), 10)
	filePath := baseDir + "/" + filehead.Filename
	outFile, _ := os.Create(filePath)
	_, err = io.Copy(outFile, file)
	if err != nil {
		panic(err)
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
	if err != nil {
		panic(err)
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
	if err != nil {
		panic(err)
	}

	id, _ := strconv.Atoi(c.Param("id"))
	filename := c.Param("filename")

	dataDir, _ := config.Conf.GetValue("project", "datadir")
	baseDir := dataDir + "/" + strconv.FormatInt(int64(id), 10)
	filepath := baseDir + "/" + filename
	err = ioutil.WriteFile(filepath, []byte(req.Content), 0755)

	if err != nil {
		panic(err)
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

	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除文件成功",
		"show":    true,
	})
}
