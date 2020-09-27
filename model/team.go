package model

import (
	"ahpuoj/utils"
	"errors"
	"gorm.io/gorm"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Team1 struct {
	ID        int            `db:"id" json:"id"`
	Name      string         `db:"name" json:"name" binding:"required,max=20"`
	CreatedAt string         `db:"created_at" json:"created_at"`
	UpdatedAt string         `db:"updated_at" json:"updated_at"`
	IsDeleted gorm.DeletedAt `db:"is_deleted" json:"is_deleted"`
	CreatorId int            `db:"creator_id" json:"creator_id"`
	Users     []User1        `gorm:"many2many:team_user;" json:"userinfos"`
}

func (Team1) TableName() string {
	return "team"
}

type Team struct {
	Id        int            `db:"id" json:"id"`
	Name      string         `db:"name" json:"name" binding:"required,max=20"`
	CreatedAt string         `db:"created_at" json:"created_at"`
	UpdatedAt string         `db:"updated_at" json:"updated_at"`
	IsDeleted gorm.DeletedAt `db:"is_deleted" json:"is_deleted"`
	CreatorId int            `db:"creator_id" json:"creator_id"`
	UserInfos []User         `json:"userinfos"`
	UserId    int            `db:"user_id" json:"user_id"`
	Username  string         `db:"username" json:"username"`
}

func (team *Team) Save() error {
	result, err := DB.Exec(`insert into team
	(name,user_id,created_at,updated_at) 
	values (?,?,NOW(),NOW())`, team.Name, team.UserId)
	if err != nil {
		return err
	}
	lastInsertId, _ := result.LastInsertId()
	team.Id = utils.Int64to32(lastInsertId)
	return err
}

func (team *Team) Update() error {
	result, err := DB.Exec(`update team set name = ?,updated_at = NOW() where id = ?`, team.Name, team.Id)
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("数据不存在")
	}
	return err
}

func (team *Team) Delete() error {
	result, err := DB.Exec("update team set is_deleted = 1 where id = ?", team.Id)
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("数据不存在")
	}
	return err
}

func (team *Team) AddUsers(userlist string) []string {
	pieces := strings.Split(userlist, "\n")
	var infos []string
	insertStmt, _ := DB.Preparex("insert into team_user(team_id,user_id,created_at,updated_at) VALUES (?,?,NOW(),NOW())")
	checkUserExistStmt, _ := DB.Preparex("select id from user where username = ?")
	checkHasUserStmt, _ := DB.Preparex(`select 1 from team_user
	inner join user on  team_user.user_id = user.id 
	where team_user.team_id = ? and user.username = ?`)

	if len(pieces) > 0 && len(pieces[0]) > 0 {
		for _, username := range pieces {
			var userId, count int
			var info string
			insertable := true

			err := checkUserExistStmt.Get(&userId, username)
			// 用户不存在不可以插入
			if err != nil {
				insertable = false
				info = "团队添加用户" + username + "失败，用户不存在"
			}

			err = checkHasUserStmt.Get(&count, team.Id, username)
			// 团队当前不含有该用户时会报error，可以插入
			if err == nil {
				insertable = false
				utils.Consolelog(err)
				info = "团队添加用户" + username + "失败，用户不存在"
			}
			if insertable {
				insertStmt.Exec(team.Id, userId)
				info = "团队添加用户" + username + "成功"
			}
			infos = append(infos, info)
		}
	}
	insertStmt.Close()
	checkUserExistStmt.Close()
	checkHasUserStmt.Close()
	return infos
}

// 附加属于该团队的人员信息 如果contestId大于0 则为竞赛作业中团队人员信息
func (team *Team) AttachUserInfo(contestId int) {
	var err error
	var rows *sqlx.Rows
	var userInfos []User
	if contestId > 0 {
		rows, err = DB.Queryx(`select user.* from contest_team_user inner join user on contest_team_user.user_id=user.id 
		where contest_team_user.contest_id = ? and contest_team_user.team_id = ?`, contestId, team.Id)
	} else {
		rows, err = DB.Queryx(`select user.* from team_user inner join user on team_user.user_id = user.id 
		where team_user.team_id = ?`, team.Id)
	}
	if err != nil {
		utils.Consolelog(err)
		return
	}
	for rows.Next() {
		var user User
		err = rows.StructScan(&user)
		userInfos = append(userInfos, user)
	}
	team.UserInfos = userInfos
}

func (team *Team) Response() map[string]interface{} {

	return map[string]interface{}{
		"id":   team.Id,
		"name": team.Name,
	}
}

func (team *Team) ResponseWithUsers() map[string]interface{} {

	return map[string]interface{}{
		"id":        team.Id,
		"name":      team.Name,
		"userinfos": team.UserInfos,
	}
}
