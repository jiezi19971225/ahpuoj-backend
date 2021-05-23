package service

import (
	"ahpuoj/constant"
	"ahpuoj/dto"
	"ahpuoj/entity"
	"ahpuoj/utils"
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
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
		cnt := 1
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
	// 检查竞赛作业是否存在
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

func (this *ContestService) CalcStatus(contest *entity.Contest) (status int) {

	nowTime := time.Now()
	if nowTime.Unix() < contest.StartTime.Unix() {
		return constant.CONTEST_NOT_START
	} else if nowTime.Unix() > contest.EndTime.Unix() {
		return constant.CONTEST_FINISH
	} else {
		return constant.CONTEST_RUNNING
	}
}

// 查重
var mutex sync.Mutex

func (this *ContestService) SimCheck(contest *entity.Contest, solution *entity.Solution, source string) (hasRepeat bool) {

	currentRootDir := utils.GetCurrentExecDirectory()

	type SolutionWithSource struct {
		Source     string `db:"source" json:"source"`
		SolutionId int    `db:"solution_id" json:"solution_id"`
		Language   int    `db:"language" json:"language"`
	}
	var SolutionWithSourceList []SolutionWithSource
	this.Raw("select source_code.solution_id,source_code.source,language from solution "+
		"INNER JOIN source_code on solution.solution_id = source_code.solution_id "+
		"where contest_id = ? and num = ? and user_id != ? and result = 4"+
		" order by solution_id desc;", contest.ID, solution.Num, solution.UserId).Scan(&SolutionWithSourceList)

	// 下面这一段进行了加锁处理
	// TODO 这里的加锁处理应该可以优化
	mutex.Lock()

	// 将用户提交的代码写入临时文件
	userSourcePath := path.Join(currentRootDir, "simtmp", "simtmp.txt")
	ioutil.WriteFile(userSourcePath, []byte(source), 0777)

	// 将竞赛通过的代码写入临时目录
	contestSolutionsPath := path.Join(currentRootDir, "simtmp", "c"+strconv.Itoa(contest.ID))
	utils.CreateDir(contestSolutionsPath, 0777)

	for _, solution := range SolutionWithSourceList {
		filePath := path.Join(contestSolutionsPath, strconv.Itoa(solution.SolutionId)+"."+constant.LanguageExt[solution.Language])
		if pathExist, _ := utils.PathExists(filePath); !pathExist {
			ioutil.WriteFile(filePath, []byte(solution.Source), 0777)
		}
	}

	// 调用 sim 程序
	simExeName := "sim_text"
	if name, ok := constant.SimExeMap[solution.Language]; ok {
		simExeName = name
	}
	currentPath := path.Join(currentRootDir, "sim")

	if runtime.GOOS == "windows" {
		simExeName += ".exe"
	}

	var cmd *exec.Cmd
	var out bytes.Buffer

	cmd = exec.Command("./"+simExeName, "-pt "+strconv.Itoa(contest.CheckRepeatRate), userSourcePath, "|", contestSolutionsPath+"/*."+constant.LanguageExt[solution.Language])
	cmd.Dir = currentPath
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		if err != nil {
			panic(err)
		}
	}
	log.Print(out.String())
	result := strings.Contains(out.String(), "consists for")

	mutex.Unlock()

	return result
}
