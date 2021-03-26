package controller

import (
	"ahpuoj/config"
	"ahpuoj/dao/orm"
	"ahpuoj/dto"
	"ahpuoj/entity"
	"ahpuoj/utils"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/guregu/null.v4"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,ascii,max=20"`
		Password string `json:"password" binding:"required,ascii,min=6,max=20"`
	}
	var user entity.User
	err := c.ShouldBindJSON(&req)
	if utils.CheckError(c, err, "参数错误") != nil {
		return
	}
	err = ORM.Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		panic(err)
	}
	h := sha1.New()
	h.Write([]byte(user.Passsalt))
	h.Write([]byte(req.Password))
	hashedPassword := fmt.Sprintf("%x", h.Sum(nil))
	if hashedPassword != user.Password {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "密码错误",
		})
	} else {
		// 根据用户名payload生成token
		token := utils.CreateToken(req.Username)
		// 更新redis的token,过期时间为15天
		utils.Consolelog("登录成功")
		utils.Consolelog(token)
		conn := REDIS.Get()
		defer conn.Close()
		conn.Do("setex", "token:"+req.Username, 60*60*24*15, token)

		// 用户登录刷新redis缓存信息
		var role entity.Role
		var userWithRole dto.UserWithRoleDto
		orm.ORM.Model(entity.Role{}).Where("id = ?", user.RoleId).Find(&role)
		userWithRole.User = user
		userWithRole.Role = role.Name
		serializedUserInfo, _ := json.Marshal(userWithRole)
		conn.Do("setex", "userinfo:"+user.Username, 60*60*24, serializedUserInfo)

		// 设置cookies
		domain, _ := config.Conf.GetValue("project", "cookiedomain")
		cookieLiveTimeStr, _ := config.Conf.GetValue("project", "cookielivetime")
		cookieLiveTime, _ := strconv.Atoi(cookieLiveTimeStr)
		c.SetCookie("access-token", token, cookieLiveTime, "/", domain, false, false)
		c.JSON(http.StatusOK, gin.H{
			"message": "登录成功",
			"show":    true,
		})
	}
}

func Register(c *gin.Context) {

	defaultAvatar, _ := config.Conf.GetValue("preset", "avatar")

	var req struct {
		Email           string `json:"email" binding:"required,email,max=40"`
		Username        string `json:"username" binding:"required,ascii,max=20"`
		Nick            string `json:"nick" binding:"required,max=20"`
		Password        string `json:"password" binding:"required,ascii,min=6,max=20,eqfield=ConfirmPassword"`
		ConfirmPassword string `json:"confirmpassword" binding:"required`
	}
	err := c.ShouldBindJSON(&req)
	if utils.CheckError(c, err, "参数错误") != nil {
		return
	}
	// 加盐处理 16位随机字符串
	salt := utils.GetRandomString(16)
	h := sha1.New()
	h.Write([]byte(salt))
	h.Write([]byte(req.Password))
	hashedPassword := fmt.Sprintf("%x", h.Sum(nil))
	user := entity.User{
		Username: req.Username,
		Nick:     req.Nick,
		Email:    null.StringFrom(req.Email),
		Password: hashedPassword,
		Passsalt: salt,
		RoleId:   1,
		Avatar:   defaultAvatar,
	}
	err = ORM.Create(&user).Error
	if utils.CheckError(c, err, "注册失败，邮箱/用户名/昵称可能已被注册") != nil {
		return
	}
	token := utils.CreateToken(user.Username)
	// 更新redis的token,过期时间为15天
	conn := REDIS.Get()
	defer conn.Close()
	conn.Do("setex", "token:"+user.Username, 60*60*24*15, token)
	// 设置cookies
	domain, _ := config.Conf.GetValue("project", "cookiedomain")
	cookieLiveTimeStr, _ := config.Conf.GetValue("project", "cookielivetime")
	cookieLiveTime, _ := strconv.Atoi(cookieLiveTimeStr)
	c.SetCookie("access-token", token, cookieLiveTime, "/", domain, false, false)
	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功",
		"token":   token,
		"show":    true,
	})
}

// 发送重设密码邮件的接口
func SendFindPassEmail(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,max=40"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		panic(err)
	}
	var user entity.User
	err = ORM.Model(entity.User{}).Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		panic(errors.New("用户不存在"))
	}
	if !user.Email.Valid {
		panic(errors.New("用户未绑定邮箱"))
	}
	// 生成随机字符串
	token := utils.GetRandomString(30)
	ORM.Exec("insert into resetpassword(user_id,token,expired_at) values(?,?,date_add(NOW(),INTERVAL 1 hour)) on duplicate key update token = ?,expired_at=date_add(NOW(),INTERVAL 1 hour)", user.ID, token, token)
	server, _ := config.Conf.GetValue("project", "server")
	mailTo := []string{
		user.Email.String,
	}
	//邮件主题
	subject := "AHPUOJ重设密码邮件"
	// 邮件正文
	body := fmt.Sprintf("请访问以下连接重设您的密码，链接将会在1小时内失效，请尽快进行设置 <a href=\"%s/resetpass?token=%s\">%s/resetpass?token=%s</a>", server, token, server, token)
	utils.SendMail(mailTo, subject, body)
	c.JSON(http.StatusOK, gin.H{
		"message": "已成功发送重设密码邮件，请前往邮箱查看",
		"show":    true,
	})
}

// 验证重设密码token是否正确
func VeriryResetPassToken(c *gin.Context) {
	token := c.Query("token")
	var resetPassword entity.ResetPassword
	err := ORM.Model(entity.ResetPassword{}).Where("token = ?", token).First(&resetPassword).Error
	if err != nil {
		panic(errors.New("token非法"))
	}
	now := time.Now()
	if now.After(resetPassword.ExpiredAt) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "token已过期，请重新发送邮件",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "token验证成功，请立即重设密码",
		"show":    true,
	})
}

func ResetPassByToken(c *gin.Context) {
	var req struct {
		Token           string `json:"token" binding:"required"`
		Password        string `json:"password" binding:"required,ascii,min=6,max=20,eqfield=ConfirmPassword"`
		ConfirmPassword string `json:"confirmpassword" binding:"required`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		panic(err)
	}

	// 验证token
	var resetPassword entity.ResetPassword
	err = ORM.Model(entity.ResetPassword{}).Where("token = ?", req.Token).First(&resetPassword).Error
	if err != nil {
		panic(errors.New("token非法"))
	}

	now := time.Now()
	if now.After(resetPassword.ExpiredAt) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "token已过期，请重新发送邮件",
		})
		return
	}
	// 更新密码
	// 加盐处理 16位随机字符串
	h := sha1.New()
	salt := utils.GetRandomString(16)
	h.Reset()
	h.Write([]byte(salt))
	h.Write([]byte(req.Password))
	hashedPassword := fmt.Sprintf("%x", h.Sum(nil))
	ORM.Model(entity.User{ID: resetPassword.UserId}).Updates(map[string]interface{}{"password": hashedPassword, "passsalt": salt})
	ORM.Delete(&resetPassword)
	c.JSON(http.StatusOK, gin.H{
		"message": "密码修改成功",
		"show":    true,
	})

}
