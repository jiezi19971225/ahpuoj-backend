package service

import (
	"ahpuoj/config"
	"ahpuoj/dto"
	"ahpuoj/entity"
	"ahpuoj/mq"
	"ahpuoj/utils"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
)

type ProblemService struct {
	*gorm.DB
}

func (this *ProblemService) List(c *gin.Context) ([]dto.ProblemDto, int64) {
	param := c.Query("param")
	query := this.Model(&entity.Problem{})

	if len(param) > 0 {
		query.Where("title like ?", "%"+param+"%")
	}
	var total int64
	query.Count(&total)
	var results []dto.ProblemDto
	query.Scopes(utils.Paginate(c)).Order("problem.id desc").Select("problem.*", "user.username").Joins("inner join user on problem.user_id = user.id").Find(&results)

	// 加载标签数据
	for _, result := range results {
		this.Model(&result.Problem).Association("Tags").Find(&result.Tags)
	}
	return results, total
}

func (this *ProblemService) SaveRecord(problem *entity.Problem) {
	err := this.Create(problem).Error
	if err != nil {
		panic(err)
	}
	// 创建题目数据文件夹
	dataDir, _ := config.Conf.GetValue("project", "datadir")
	baseDir := dataDir + "/" + strconv.FormatInt(int64(problem.ID), 10)
	utils.Consolelog(baseDir)
	err = os.MkdirAll(baseDir, 0777)
	if err != nil {
		panic(err)
	}
}

func (this *ProblemService) DeleteRecord(problem *entity.Problem) {
	err := this.Transaction(func(tx *gorm.DB) error {
		err := tx.Delete(problem).Error
		if err != nil {
			return err
		}
		// 删除其他相关数据
		// 删除source_code
		err = tx.Exec("delete source_code from source_code inner join solution on source_code.solution_id = solution.solution_id where solution.problem_id = ?", problem.ID).Error
		if err != nil {
			return err
		}
		err = tx.Exec("delete compileinfo from compileinfo inner join solution on compileinfo.solution_id = solution.solution_id where solution.problem_id = ?", problem.ID).Error
		tx.Exec("delete runtimeinfo from runtimeinfo inner join solution on runtimeinfo.solution_id = solution.solution_id where solution.problem_id = ?", problem.ID)
		if err != nil {
			return err
		}
		// 删除solution记录
		err = tx.Exec("delete from solution where problem_id = ?", problem.ID).Error
		if err != nil {
			return err
		}

		// 删除tag关联记录
		err = tx.Exec("delete from problem_tag where problem_id = ?", problem.ID).Error
		if err != nil {
			return err
		}
		// 删除reply
		err = tx.Exec("delete reply from reply inner join issue on reply.issue_id =issue.id where issue.problem_id = ?", problem.ID).Error
		if err != nil {
			return err
		}
		// 删除issue
		err = tx.Exec("delete from issue where problem_id = ?", problem.ID).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (this *ProblemService) AddTags(problem *entity.Problem, reqTags []interface{}) {
	for _, tag := range reqTags {
		switch t := tag.(type) {
		// 如果是标签id 直接插入 golang解析json会将数字转成float64
		case float64:
			this.Create(&entity.ProblemTag{
				ProblemID: problem.ID,
				TagID:     int(t),
			})
			// 否则生成新的tag并且插入
		case string:
			newTag := entity.Tag{
				Name: t,
			}
			err := this.Create(&newTag).Error
			if err == nil {
				this.Create(&entity.ProblemTag{
					ProblemID: problem.ID,
					TagID:     newTag.ID,
				})
			}
		}
	}
}

func (this *ProblemService) ReplaceTags(problem *entity.Problem, reqTags []interface{}) {
	this.Model(problem).Association("Tags").Clear()
	this.AddTags(problem, reqTags)
}

func (this *ProblemService) Reassignment(oldId int, newId int) {
	var oldProblem entity.Problem
	var newProblem entity.Problem

	err := this.First(&oldProblem, oldId).Error
	if err != nil {
		panic(err)
	}
	err = this.First(&newProblem, newId).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		panic(errors.New("重排问题失败，新问题ID已有问题"))
	}

	err = this.Transaction(func(tx *gorm.DB) error {
		// 更改问题ID
		if err := tx.Exec("update problem set id = ? where id = ?", newId, oldId).Error; err != nil {
			return err
		}
		// 更新solution表
		if err := tx.Exec("update solution set problem_id = ? where problem_id = ?", newId, oldId).Error; err != nil {
			return err
		}
		// 更新tag关联表
		if err := tx.Exec("update problem_tag set problem_id = ? where problem_id = ?", newId, oldId).Error; err != nil {
			return err
		}
		// 更新issue表
		if err := tx.Exec("update issue set problem_id = ? where problem_id = ?", newId, oldId).Error; err != nil {
			return err
		}
		// 移动文件夹
		dataDir, _ := config.Conf.GetValue("project", "datadir")
		oldDir := dataDir + "/" + strconv.Itoa(oldId)
		newDir := dataDir + "/" + strconv.Itoa(newId)
		err = os.Rename(oldDir, newDir)
		if err != nil {
			return err
		}
		return nil
	})

	// 更新自增起始ID
	var maxId int
	this.Model(entity.Problem{}).Select("max(id)").Scan(&maxId)
	newAutoIncrement := strconv.Itoa(maxId + 1)
	this.Exec("alter table problem auto_increment=" + newAutoIncrement)

	if err != nil {
		panic(err)
	}
}

func (this *ProblemService) RejudgeSolution(id int) {
	var info dto.RejudgeInfo
	solution := entity.Solution{ID: id}
	resultRows := this.Model(&solution).Select("solution.solution_id,solution.user_id,solution.problem_id,solution.language,problem.time_limit,problem.memory_limit,source_code.source").Joins("inner join problem on solution.problem_id=problem.id").Joins("inner join source_code on solution.solution_id=source_code.solution_id").Scan(&info).RowsAffected
	if resultRows == 0 {
		panic(errors.New("提交不存在"))
	}
	err := this.Model(&solution).Update("result", 1).Error
	if err != nil {
		panic(err)
	}
	jsondata, _ := json.Marshal(gin.H{
		"UserId":       info.UserId,
		"TestrunCount": 0,
		"SolutionId":   info.SolutionId,
		"ProblemId":    info.ProblemId,
		"Language":     info.Language,
		"TimeLimit":    info.TimeLimit,
		"MemoryLimit":  info.MemoryLimit,
		"Source":       info.Source,
		"InputText":    "",
	})
	mq.Publish("oj", "problem", jsondata)
}

func (this *ProblemService) RejudgeProblem(id int) {
	var infos []dto.RejudgeInfo
	problem := entity.Problem{ID: id}
	err := this.First(&problem).Error
	if err != nil {
		panic(err)
	}
	this.Model(&entity.Solution{}).Select("solution.solution_id,solution.user_id,solution.problem_id,solution.language,problem.time_limit,problem.memory_limit,source_code.source").Joins("inner join problem on solution.problem_id=problem.id").Joins("inner join source_code on solution.solution_id=source_code.solution_id").Where("solution.problem_id = ?", id).Scan(&infos)
	// 更新状态为等待重判
	this.Model(&entity.Solution{}).Where("problem_id = ?", id).Update("result", 1)
	go func() {
		for _, info := range infos {
			log.Print(info)
			jsondata, _ := json.Marshal(gin.H{
				"UserId":       info.UserId,
				"TestrunCount": 0,
				"SolutionId":   info.SolutionId,
				"ProblemId":    info.ProblemId,
				"Language":     info.Language,
				"TimeLimit":    info.TimeLimit,
				"MemoryLimit":  info.MemoryLimit,
				"Source":       info.Source,
				"InputText":    "",
			})
			mq.Publish("oj", "problem", jsondata)
		}
	}()
}
