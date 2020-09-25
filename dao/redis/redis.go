package redis

import (
	"ahpuoj/config"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

var REDIS *redis.Pool

func init() {
	rediscfg, _ := config.Conf.GetSection("redis")
	path := strings.Join([]string{rediscfg["host"], ":", rediscfg["port"]}, "")

	REDIS = &redis.Pool{
		MaxIdle:     100,
		MaxActive:   10000,
		Wait:        true,
		IdleTimeout: 5 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", path)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", rediscfg["password"]); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
