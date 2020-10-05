package service

import (
	"ahpuoj/dto"
	"ahpuoj/entity"
	"ahpuoj/model"
	"ahpuoj/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"strings"
)

type TeamService struct {
	*gorm.DB
}

func (this *TeamService) List(c *gin.Context) ([]dto.TeamDto, int64) {
	param := c.Query("param")
	query := this.Model(entity.Team{})

	if len(param) > 0 {
		query.Where("name like ?", "%"+param+"%")
	}
	var total int64
	query.Count(&total)

	results := []dto.TeamDto{}
	query.Scopes(utils.Paginate(c)).Order("team.id desc").Select("team.*", "user.username").Joins("inner join user on team.user_id = user.id").Find(&results)
	return results, total
}

func (this *TeamService) Users(team *entity.Team, c *gin.Context) ([]entity.User, int64) {
	query := this.Model(team)
	param := c.Query("param")

	if len(param) > 0 {
		query.Where("username like ?", "%"+param+"%")
	}

	total := query.Session(&gorm.Session{WithConditions: true}).Association("Users").Count()

	var users []entity.User
	err := query.Scopes(utils.Paginate(c)).Select("user.*").Order("user.id desc").Association("Users").Find(&users)
	if err != nil {
		panic(err)
	}
	return users, total
}

func (this *TeamService) AddUsers(team *entity.Team, userlist string) []string {
	var infos []string
	var users []entity.User
	pieces := strings.Split(userlist, "\n")
	if len(pieces) > 0 && len(pieces[0]) > 0 {
		for _, username := range pieces {
			var count int64
			var info string
			var user entity.User
			err := this.Model(entity.User{}).Where("username = ?", username).Take(&user).Error
			// 用户不存在不可以插入
			if errors.Is(err, gorm.ErrRecordNotFound) {
				info = "团队添加用户" + username + "失败，用户不存在"
			} else {
				this.Model(model.TeamUser{}).Where("team_id = ? and user_id = ?", team.ID, user.ID).Count(&count)
				if count != 0 {
					info = "团队添加用户" + username + "失败，用户不存在"
				} else {
					users = append(users, user)
					if err != nil {
						log.Print(err, "Error", err.Error())
					}
					this.Create(&entity.TeamUser{
						TeamID: team.ID,
						UserID: user.ID,
					})
					info = "团队添加用户" + username + "成功"
				}
			}
			infos = append(infos, info)
		}
	}
	return infos
}

func (this *TeamService) DeleteUser(team *entity.Team, user entity.User) {
	this.Model(team).Association("Users").Delete(&user)
	// 级联删除
	// DB.Exec(`delete contest_user from contest_user inner join contest_team_user on contest_user.contest_id = contest_team_user.contest_id
	// where contest_user.user_id = ? and contest_team_user.team_id = ?`, userId, teamId)
	// DB.Exec("delete from contest_team_user where team_id = ? and user_id = ?", teamId, userId)
}
