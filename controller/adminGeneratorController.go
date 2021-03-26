package controller

import (
	"ahpuoj/config"
	"ahpuoj/constant"
	"ahpuoj/dto"
	"ahpuoj/entity"
	"ahpuoj/utils"
	"crypto/sha1"
	"fmt"
	"gopkg.in/guregu/null.v4"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func CompeteAccountGenerator(c *gin.Context) {
	defaultAvatar, _ := config.Conf.GetValue("preset", "avatar")

	var req struct {
		Prefix string `json:"prefix" binding:"required,max=15"`
		Number int    `json:"number" binding:"required,min=1,max=200"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		panic(err)
	}
	var infos []string
	users := []dto.UserWithPasswordDto{}
	for i := 1; i <= req.Number; i++ {
		username := req.Prefix + strconv.Itoa(i)
		randomPassword := utils.GetRandomString(15)
		h := sha1.New()
		h.Write([]byte(randomPassword))
		hashedPassword := fmt.Sprintf("%x", h.Sum(nil))
		user := entity.User{
			Username:      username,
			Nick:          username,
			Email:         null.StringFrom(""),
			Password:      hashedPassword,
			IsCompeteUser: 1,
			RoleId:        constant.ROLE_USER,
			Avatar:        defaultAvatar,
		}
		err := ORM.Create(&user).Error
		if err != nil {
			infos = append(infos, "用户"+username+"创建失败")
		} else {
			users = append(users, dto.UserWithPasswordDto{
				Username: user.Username,
				Password: randomPassword,
			})
			infos = append(infos, "用户"+username+"创建成功")
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "操作成功",
		"users":   users,
		"info":    infos,
	})
}

func UserAccountGenerator(c *gin.Context) {
	defaultAvatar, _ := config.Conf.GetValue("preset", "avatar")

	var req struct {
		UserList string `json:"userlist" binding:"required"`
	}
	err := c.ShouldBindJSON(&req)
	if utils.CheckError(c, err, "参数错误") != nil {
		return
	}

	pieces := strings.Split(req.UserList, "\n")

	var infos []string
	var users []interface{}
	if len(pieces) > 0 && len(pieces[0]) > 0 {
		for _, username := range pieces {

			password := "123456"
			// 更新密码
			// 加盐处理 16位随机字符串
			salt := utils.GetRandomString(16)
			h := sha1.New()
			h.Write([]byte(salt))
			h.Write([]byte(password))
			hashedPassword := fmt.Sprintf("%x", h.Sum(nil))

			user := entity.User{
				Username: username,
				Nick:     username,
				Email:    null.StringFrom(""),
				Password: hashedPassword,
				Passsalt: salt,
				RoleId:   constant.ROLE_USER,
				Avatar:   defaultAvatar,
			}
			err := ORM.Create(&user).Error
			if err == nil {
				users = append(users, map[string]interface{}{
					"username": username,
					"password": password,
				})
				infos = append(infos, "用户"+username+"创建成功")
			} else {
				infos = append(infos, "用户"+username+"创建失败")
			}

		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "操作成功",
		"users":   users,
		"info":    infos,
	})
}
