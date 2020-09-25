package middleware

import (
	redisDao "ahpuoj/dao/redis"
	"github.com/gomodule/redigo/redis"
)

var REDIS *redis.Pool

func init() {
	REDIS = redisDao.REDIS
}
