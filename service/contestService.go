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
	query.Count(&total)
	var results []dto.ContestDto
	query.Scopes(utils.Paginate(c)).Order("contest.id desc").Select("contest.*", "user.username").Joins("inner join user on contest.user_id = user.id").Find(&results)

	return results, total
}

func (this *ContestService) AttachProblems(contest *entity.Contest) dto.ContestDetailDto {
	var problemIds []string
	type result struct {
		ID string
	}
	var res []result
	this.Model(contest).Select("problem.id").Association("Problems").Find(&res)

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
				this.Create(&entity.ContestProblem{
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
		query.Where("username like ? or nick like ?", "%"+param+"%", "%"+param+"%")
	}

	total := query.Session(&gorm.Session{WithConditions: true}).Association("Users").Count()
	var users []entity.User
	err := query.Scopes(utils.Paginate(c)).Select("user.*").Order("user.id desc").Association("Users").Find(&users)
	if err != nil {
		panic(err)
	}
	return users, total
}

/**
teamId大于0 则为向团队添加成员
*/
func (this *ContestService) AddUsers(contest *entity.Contest, userlist string, teamId int) []string {

	var infos []string
	var contestUsers []entity.ContestUser
	pieces := strings.Split(userlist, "\n")
	if len(pieces) > 0 && len(pieces[0]) > 0 {
		for _, username := range pieces {
			var count int64
			var user entity.User
			err := this.Model(entity.User{}).Where("username = ?", username).Take(&user).Error
			// 用户不存在不可以插入
			if errors.Is(err, gorm.ErrRecordNotFound) {
				infos = append(infos, "竞赛&作业添加用户"+username+"失败，用户不存在")
				continue
			}

			this.Model(entity.ContestUser{}).Where("contest_id = ? and user_id = ?", contest.ID, user.ID).Count(&count)
			// 判断是否已经添加了用户进入竞赛作业中
			if count != 0 {
				infos = append(infos, "竞赛&作业添加用户"+username+"失败，用户不存在")
				continue
			}

			// 判断用户是否属于团队
			if teamId > 0 {
				this.Model(entity.TeamUser{}).Where("team_id = ? and user_id = ?", teamId, user.ID).Count(&count)
				if count == 0 {
					infos = append(infos, "竞赛&作业添加用户"+username+"失败，用户不属于该团队")
					continue
				}
			}
			if err != nil {
				log.Print(err, "Error", err.Error())
			}
			contestUsers = append(contestUsers, entity.ContestUser{
				ContestID: contest.ID,
				TeamID:    teamId,
				UserID:    user.ID,
			})
			infos = append(infos, "竞赛&作业添加用户"+username+"成功")
		}
	}
	this.Create(&contestUsers)
	return infos
}

func (this *ContestService) AddTeam(contest *entity.Contest, team *entity.Team) {
	err := this.First(&contest).Error
	if err != nil {
		panic(err)
	}
	// 检查团队是否存在
	err = this.First(&team).Error
	if err != nil {
		panic(err)
	}
	// 检查是否已经添加进了竞赛作业中
	var count int64
	this.Model(&entity.ContestTeam{}).Where("contest_id = ? and team_id = ?", contest.ID, team.ID).Count(&count)
	if count > 0 {
		panic(errors.New("该团队已经在该竞赛作业中"))
	}
	err = this.Create(&entity.ContestTeam{
		ContestID: contest.ID,
		TeamID:    team.ID,
	}).Error
	if err != nil {
		panic(err)
	}
}

func (this *ContestService) AddTeamAllUsers(contest *entity.Contest, team *entity.Team) []string {

	err := this.First(&contest).Error
	if err != nil {
		panic(err)
	}
	// 检查团队是否存在
	err = this.First(&team).Error
	if err != nil {
		panic(err)
	}
	var infos []string
	var users []entity.User
	var contestUsers []entity.ContestUser
	this.Model(&team).Association("Users").Find(&users)

	for _, user := range users {
		var info string
		var count int64
		this.Model(entity.ContestUser{}).Where("contest_id = ? and test_id =?", contest.ID, team.ID).Count(&count)
		// 有记录返回err==nil
		if count > 0 {
			info = "竞赛&作业添加用户" + user.Username + "失败，用户已被添加"
		} else {
			contestUsers = append(contestUsers, entity.ContestUser{
				ContestID: contest.ID,
				TeamID:    team.ID,
				UserID:    user.ID,
			})
			info = "竞赛&作业添加用户" + user.Username + "成功"
		}
		infos = append(infos, info)
	}
	err = this.Create(&contestUsers).Error
	if err != nil {
		panic(err)
	}
	return infos
}

func (this *ContestService) DeleteUser(contest *entity.Contest, user *entity.User) {
	err := this.Model(contest).Association("Users").Delete(user)
	if err != nil {
		panic(err)
	}
}
func (this *ContestService) DeleteTeamUser(contest *entity.Contest, team *entity.Team, user *entity.User) {
	err := this.Model(contest).Where("contest_user.team_id = ?", team.ID).Association("Users").Delete(user)
	if err != nil {
		panic(err)
	}
}

func (this *ContestService) TeamList(contest *entity.Contest) []dto.ContestTeamDto {
	teams := []entity.Team{}
	results := []dto.ContestTeamDto{}
	this.Model(&contest).Select("team.*").Association("Teams").Find(&teams)

	for _, team := range teams {
		contestTeamDto := dto.ContestTeamDto{Team: team}
		this.Model(&contest).Select("user.*").Where("contest_user.team_id = ?", team.ID).Association("Users").Find(&contestTeamDto.Userinfos)
		results = append(results, contestTeamDto)
	}
	return results
}

func (this *ContestService) DeleteTeam(contest *entity.Contest, team *entity.Team) {
	err := this.Transaction(func(tx *gorm.DB) error {
		if err := this.Model(contest).Association("Teams").Delete(team); err != nil {
			return err
		}
		// 级联删除
		if err := this.Model(contest).Where("contest_user.team_id = ?", team.ID).Association("Users").Clear(); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
