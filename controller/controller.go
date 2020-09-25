package controller

import (
	"ahpuoj/config"
	mysqlDao "ahpuoj/dao/mysql"
	redisDao "ahpuoj/dao/redis"
	"ahpuoj/model"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"strconv"
)

var DB *sqlx.DB
var REDIS *redis.Pool
var RedisCacheLiveTime int

func init() {
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

// 获得user实例
func GetUserInstance(c *gin.Context) (model.User, bool) {
	var user model.User
	userInterface, loggedIn := c.Get("user")
	if userInterface, ok := userInterface.(model.User); ok {
		user = userInterface
	}
	return user, loggedIn
}
