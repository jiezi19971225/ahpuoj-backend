package middleware

import (
	"ahpuoj/dao/orm"
	"ahpuoj/dto"
	"ahpuoj/entity"
	"ahpuoj/utils"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
)

func parseToken(c *gin.Context) (dto.UserWithRoleDto, error) {

	tokenString := c.GetHeader("Authorization")
	// 用于文件下载请求，从 cookies 中读取
	if tokenString == "" {
		tokenString, _ = c.Cookie("access-token")
	}
	var user dto.UserWithRoleDto
	var role entity.Role

	token, err := jwt.ParseWithClaims(tokenString, &utils.MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(utils.TokenSinature), nil
	})

	// 忽略超时错误
	if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&ve.Errors&jwt.ValidationErrorExpired != 0 {
		} else {
			return user, err
		}
	}

	if claims, ok := token.Claims.(*utils.MyCustomClaims); ok {
		c.Set("tokenExpireAt", claims.ExpiresAt)
		username := claims.UserName
		userEntity := entity.User{}
		conn := REDIS.Get()
		defer conn.Close()
		// 首先尝试从缓存中获取用户数据
		cachedUserInfo, err := redis.Bytes(conn.Do("get", "userinfo:"+username))
		if err == nil {
			json.Unmarshal(cachedUserInfo, &user)
		} else {
			err = orm.ORM.Model(entity.User{}).Where("username = ?", username).Find(&userEntity).Error
			if err != nil {
				return user, errors.New("用户不存在")
			}
			orm.ORM.Model(entity.Role{}).Where("id = ?", userEntity.RoleId).Find(&role)
			user.User = userEntity
			user.Role = role.Name
			serializedUserInfo, _ := json.Marshal(user)
			conn.Do("setex", "userinfo:"+username, 60*60*24, serializedUserInfo)
		}
		// 判断用户登录token是否存在redis缓存中
		storeToken, _ := redis.String(conn.Do("get", "token:"+username))
		if storeToken != tokenString {
			return user, errors.New("token已被废弃")
		}
	} else {
		return user, errors.New("token结构不匹配")
	}
	return user, nil
}

func ParseTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 判断用户是否存在
		user, err := parseToken(c)
		if err != nil {
			c.Next()
		} else {
			c.Set("user", user)
			c.Next()
		}
	}
}
