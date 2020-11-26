package controller

import (
	"ahpuoj/config"
	"ahpuoj/constant"
	"ahpuoj/entity"
	"ahpuoj/mq"
	"ahpuoj/utils"
	"encoding/json"
	"errors"
	"gopkg.in/guregu/null.v4"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func ImportProblemSet(c *gin.Context) {

	user, _ := GetUserInstance(c)

	filehead, err := c.FormFile("file")
	if utils.CheckError(c, err, "文件上传失败") != nil {
		return
	}
	file, err := filehead.Open()
	if utils.CheckError(c, err, "文件打开失败") != nil {
		return
	}
	fps, err := utils.ImportFps(file)
	if utils.CheckError(c, err, "问题导入失败") != nil {
		return
	}
	var infos []string

	for _, item := range fps.Item {
		timeLimit, _ := strconv.Atoi(item.TimeLimit.Content)
		memoryLimit, _ := strconv.Atoi(item.MemoryLimit.Content)
		if item.TimeLimit.Unit == "ms" {
			timeLimit /= 1000
		}
		if item.MemoryLimit.Unit == "kb" {
			memoryLimit /= 1024
		}
		problem := entity.Problem{
			Title:        item.Title,
			Description:  utils.RelativeNullString(null.StringFrom(item.Description)),
			Input:        utils.RelativeNullString(null.StringFrom(item.Description)),
			Output:       utils.RelativeNullString(null.StringFrom(item.Description)),
			SampleInput:  null.StringFrom(item.Description),
			SampleOutput: null.StringFrom(item.Description),
			Hint:         utils.RelativeNullString(null.StringFrom(item.Description)),
			TimeLimit:    timeLimit,
			MemoryLimit:  memoryLimit,
		}
		err := ORM.Create(problem).Error

		if err != nil {
			infos = append(infos, "问题"+problem.Title+"导入失败")
		} else {
			infos = append(infos, "问题"+problem.Title+"导入成功")
			pid := problem.ID
			dataDir, _ := config.Conf.GetValue("project", "datadir")
			baseDir := dataDir + "/" + strconv.Itoa(pid)
			err = os.MkdirAll(baseDir, 0777)
			if err != nil {
				utils.Consolelog(err.Error())
			}
			if len(item.SampleInput) > 0 {
				utils.Mkdata(pid, "sample.in", item.SampleInput)
			}
			if len(item.SampleOutput) > 0 {
				utils.Mkdata(pid, "sample.out", item.SampleOutput)
			}
			for index, testin := range item.TestInput {
				utils.Mkdata(pid, "test"+strconv.Itoa(index)+".in", testin)
			}
			for index, testout := range item.TestOutput {
				utils.Mkdata(pid, "test"+strconv.Itoa(index)+".out", testout)
			}
			// 提交默认答案
			for _, source := range item.Solution {

				var languageId int
				// 查找 language 的index
				for k, v := range constant.LanguageName {
					if v == source.Language {
						languageId = k
						break
					}
				}
				solution := entity.Solution{
					ProblemId:  problem.ID,
					TeamId:     0,
					UserId:     user.ID,
					ContestId:  0,
					Num:        0,
					Result:     0,
					InDate:     time.Now(),
					IP:         c.ClientIP(),
					Language:   languageId,
					CodeLength: len(source.Content),
				}
				err = ORM.Create(&solution).Error
				if err != nil {
					panic(errors.New("保存提交记录失败"))
				}
				sourceCode := entity.SourceCode{
					SolutionId: solution.ID,
					Source:     source.Content,
				}
				err = ORM.Create(&sourceCode).Error
				if err != nil {
					panic(errors.New("保存代码记录失败"))
				}
				// 将判题任务推入消息队列
				jsondata, _ := json.Marshal(gin.H{
					"UserId":       user.ID,
					"TestrunCount": 0,
					"SolutionId":   solution.ID,
					"ProblemId":    solution.ProblemId,
					"Language":     solution.Language,
					"TimeLimit":    problem.TimeLimit,
					"MemoryLimit":  problem.MemoryLimit,
					"Source":       sourceCode.Source,
					"InputText":    "",
				})
				mq.Publish("oj", "problem", jsondata)
			}
		}
	}

	c.JSON(200, gin.H{
		"message": "操作成功",
		"info":    infos,
	})
}
