package controller

import (
	"ahpuoj/config"
	"ahpuoj/dao/orm"
	redisDao "ahpuoj/dao/redis"
	"ahpuoj/dto"
	"ahpuoj/service"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

var ORM *gorm.DB
var REDIS *redis.Pool
var RedisCacheLiveTime int

/**
dao层初始化
*/
func init() {
	ORM = orm.ORM
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
var tagService service.TagService
var problemService service.ProblemService
var contestService service.ContestService
var seriesService service.SeriesService
var userService service.UserService

func init() {
	teamService = service.TeamService{ORM}
	tagService = service.TagService{ORM}
	problemService = service.ProblemService{ORM}
	contestService = service.ContestService{ORM}
	userService = service.UserService{ORM}
	seriesService = service.SeriesService{ORM}
}

// 获得user实例
func GetUserInstance(c *gin.Context) (dto.UserWithRoleDto, bool) {
	var user dto.UserWithRoleDto
	userInterface, loggedIn := c.Get("user")
	if userInterface, ok := userInterface.(dto.UserWithRoleDto); ok {
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

// 根据请求URL缓存读取数据返回
func ReadFromCacheByRequestURI(c *gin.Context) error {
	user, loggedIn := GetUserInstance(c)
	keyPrefix := "nologin"
	if loggedIn {
		keyPrefix = user.Role
	}
	conn := REDIS.Get()
	defer conn.Close()
	cachedData, err := redis.Bytes(conn.Do("get", keyPrefix+":"+c.Request.RequestURI))
	if err == nil {
		var jsonData gin.H
		json.Unmarshal(cachedData, &jsonData)
		c.JSON(http.StatusOK, jsonData)
	}
	return err
}

// 根据请求URL缓存数据
func StoreToCacheByRequestURI(c *gin.Context, serializedData []byte, expire_seconds int) {
	user, loggedIn := GetUserInstance(c)
	keyPrefix := "nologin"
	if loggedIn {
		keyPrefix = user.Role
	}
	conn := REDIS.Get()
	defer conn.Close()
	conn.Do("set", keyPrefix+":"+c.Request.RequestURI, serializedData)
	conn.Do("expire", keyPrefix+":"+c.Request.RequestURI, expire_seconds)
}
