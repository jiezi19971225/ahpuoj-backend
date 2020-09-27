package service

import (
	"ahpuoj/entity"
	"ahpuoj/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserService struct {
	*gorm.DB
}

func (this *UserService) List(c *gin.Context) ([]entity.User, int64) {
	param := c.Query("param")
	userType := c.Query("userType")
	query := this.Model(entity.User{})

	query.Where("is_compete_user = ?", userType)
	if len(param) > 0 {
		query.Where("name like ?", "%"+param+"%")
	}
	var total int64
	query.Count(&total)
	var results []entity.User
	query.Debug().Scopes(utils.Paginate(c)).Order("id desc").Find(&results)
	return results, total
}
