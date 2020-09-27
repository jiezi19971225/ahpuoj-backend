package service

import (
	"ahpuoj/config"
	"ahpuoj/dto"
	"ahpuoj/entity"
	"ahpuoj/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"os"
	"strconv"
)

type ProblemService struct {
	*gorm.DB
}

func (this *ProblemService) List(c *gin.Context) ([]dto.ProblemDto, int64) {
	param := c.Query("param")
	query := this.Model(entity.Problem{})

	if len(param) > 0 {
		query.Where("title like ?", "%"+param+"%")
	}
	var total int64
	query.Debug().Count(&total)
	var results []dto.ProblemDto
	query.Debug().Scopes(utils.Paginate(c)).Order("problem.id desc").Select("problem.*", "user.username").Joins("inner join user on problem.user_id = user.id").Find(&results)

	// 加载标签数据
	for _, result := range results {
		this.Model(&result.Problem).Association("Tags").Find(&result.Tags)
	}
	return results, total
}

func (this *ProblemService) SaveRecord(problem *entity.Problem) error {
	var err error
	this.Create(&problem)

	// 创建题目数据文件夹
	dataDir, _ := config.Conf.GetValue("project", "datadir")
	baseDir := dataDir + "/" + strconv.FormatInt(int64(problem.ID), 10)
	utils.Consolelog(baseDir)
	err = os.MkdirAll(baseDir, 0777)

	return err
}

func (this *ProblemService) DeleteRecord(problem *entity.Problem) {
	this.Delete(problem)
	// 删除其他相关数据
	// 删除source_code
	this.Exec("delete source_code from source_code inner join solution on source_code.solution_id = solution.solution_id where solution.problem_id = ?", problem.ID)
	this.Exec("delete compileinfo from compileinfo inner join solution on compileinfo.solution_id = solution.solution_id where solution.problem_id = ?", problem.ID)
	this.Exec("delete runtimeinfo from runtimeinfo inner join solution on runtimeinfo.solution_id = solution.solution_id where solution.problem_id = ?", problem.ID)

	// 删除solution记录
	this.Exec("delete from solution where problem_id = ?", problem.ID)

	// 删除tag关联记录
	this.Exec("delete from problem_tag where problem_id = ?", problem.ID)
	// 删除reply
	this.Exec("delete reply from reply inner join issue on reply.issue_id =issue.id where issue.problem_id = ?", problem.ID)
	// 删除issue
	this.Exec("delete from issue where problem_id = ?", problem.ID)
}

func (this *ProblemService) AddTags(problem *entity.Problem, reqTags []interface{}) {
	for _, tag := range reqTags {
		switch t := tag.(type) {
		// 如果是标签id 直接插入 golang解析json会将数字转成float64
		case float64:
			this.Debug().Create(&entity.ProblemTag{
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
				this.Debug().Create(&entity.ProblemTag{
					ProblemID: problem.ID,
					TagID:     newTag.ID,
				})
			}
		}
	}
}

func (this *ProblemService) ReplaceTags(problem *entity.Problem, reqTags []interface{}) {
	this.Debug().Model(problem).Association("Tags").Clear()
	this.AddTags(problem, reqTags)
}
