package controller

import (
	"ahpuoj/config"
	mysqlDao "ahpuoj/dao/mysql"
	"ahpuoj/dao/orm"
	redisDao "ahpuoj/dao/redis"
	"ahpuoj/model"
	"ahpuoj/service"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"
	"strconv"
)

var DB *sqlx.DB
var ORM *gorm.DB
var REDIS *redis.Pool
var RedisCacheLiveTime int

/**
dao层初始化
*/
func init() {
	ORM = orm.ORM
	DB = mysqlDao.DB
	REDIS = redisDao.REDIS
	// 默认1800
	RedisCacheLiveTime = 1800
	if rcltstr, err := config.Conf.GetValue("redis", "cacheLiveTime"); err == nil {
		if rclt, err := strconv.Atoi(rcltstr); err == nil {
			RedisCacheLiveTime = rclt
		}
	}
}

/**
各个service初始化
*/
var teamService service.TeamService

func init() {
	teamService = service.TeamService{ORM}
}

// 获得user实例
func GetUserInstance(c *gin.Context) (model.User, bool) {
	var user model.User
	userInterface, loggedIn := c.Get("user")
	if userInterface, ok := userInterface.(model.User); ok {
		user = userInterface
	}
	return user, loggedIn
}

func Paginate(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("perpage", "10"))
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
