package service

import (
	"ahpuoj/dto"
	"ahpuoj/entity"
	"ahpuoj/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"strconv"
	"strings"
)

type ContestService struct {
	*gorm.DB
}

func (this *ContestService) List(c *gin.Context) ([]dto.ContestDto, int64) {
	param := c.Query("param")
	query := this.Model(entity.Contest{})

	if len(param) > 0 {
		query.Where("name like ?", "%"+param+"%")
	}
	var total int64
	query.Debug().Count(&total)
	var results []dto.ContestDto
	query.Debug().Scopes(utils.Paginate(c)).Order("contest.id desc").Select("contest.*", "user.username").Joins("inner join user on contest.user_id = user.id").Find(&results)

	return results, total
}

func (this *ContestService) AttachProblems(contest *entity.Contest) dto.ContestDetailDto {
	var problemIds []string
	type result struct {
		ID string
	}
	var res []result
	this.Debug().Model(contest).Select("problem.id").Association("Problems").Find(&res)

	for _, r := range res {
		problemIds = append(problemIds, r.ID)
		log.Print(r.ID)
	}

	return dto.ContestDetailDto{
		Contest:  *contest,
		Problems: strings.Join(problemIds, ","),
	}
}

func (this *ContestService) AddProblems(contest *entity.Contest, reqProblems string) {
	pieces := strings.Split(reqProblems, ",")
	if len(pieces) > 0 && len(pieces[0]) > 0 {
		cnt := 0
		for _, value := range pieces {
			problemId, _ := strconv.Atoi(value)
			problem := entity.Problem{
				ID: problemId,
			}
			err := this.Model(entity.Problem{}).First(&problem).Error
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				this.Debug().Create(&entity.ContestProblem{
					ContestID: contest.ID,
					ProblemID: problemId,
					Num:       cnt,
				})
				cnt++
			}
		}
	}
}

func (this *ContestService) ReplaceProblems(contest *entity.Contest, reqProblems string) {
	this.Model(&contest).Association("Problems").Clear()
	this.AddProblems(contest, reqProblems)
}

func (this *ContestService) Users(contest *entity.Contest, c *gin.Context) ([]entity.User, int64) {
	query := this.Model(contest)
	param := c.Query("param")

	if len(param) > 0 {
		query.Where("username like ? or nike like ?", "%"+param+"%", "%"+param+"%")
	}

	total := query.Debug().Association("Users").Count()

	var users []entity.User
	query.Debug().Scopes(utils.Paginate(c)).Select("user.*").Order("user.id desc").Association("Users").Find(&users)
	return users, total
}

/**
teamId大于0 则为向团队添加成员
*/
func (this *ContestService) AddUsers(contest *entity.Contest, userlist string, teamId int) []string {

	var infos []string
	pieces := strings.Split(userlist, "\n")
	if len(pieces) > 0 && len(pieces[0]) > 0 {
		for _, username := range pieces {
			var count int64
			var user entity.User
			err := this.Debug().Model(entity.User{}).Where("username = ?", username).Take(&user).Error
			// 用户不存在不可以插入
			if errors.Is(err, gorm.ErrRecordNotFound) {
				infos = append(infos, "竞赛&作业添加用户"+username+"失败，用户不存在")
				continue
			}

			this.Debug().Model(entity.ContestUser{}).Where("contest_id = ? and user_id = ?", contest.ID, user.ID).Count(&count)
			// 判断是否已经添加了用户进入竞赛作业中
			if count != 0 {
				infos = append(infos, "竞赛&作业添加用户"+username+"失败，用户不存在")
				continue
			}

			// 判断用户是否属于团队
			if teamId > 0 {
				this.Debug().Model(entity.TeamUser{}).Where("team_id = ? and user_id = ?", teamId, user.ID).Count(&count)
				if count == 0 {
					infos = append(infos, "竞赛&作业添加用户"+username+"失败，用户不属于该团队")
					continue
				}
			}
			if err != nil {
				log.Print(err, "Error", err.Error())
			}
			this.Debug().Create(&entity.ContestUser{
				ContestID: contest.ID,
				TeamID:    teamId,
				UserID:    user.ID,
			})
			infos = append(infos, "竞赛&作业添加用户"+username+"成功")
		}
	}
	return infos
}

func (this *ContestService) AddTeam(contest *entity.Contest, team *entity.Team) {

}

func (this *ContestService) DeleteUser(contest *entity.Contest, user *entity.User) {
	this.Debug().Model(contest).Association("Users").Delete(user)
}
