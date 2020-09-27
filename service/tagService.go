package service

import (
	"ahpuoj/entity"
	"ahpuoj/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TagService struct {
	*gorm.DB
}

func (this *TagService) Delete(tag *entity.Tag) {
	this.Delete(tag)
	//if err != nil {
	//	return err
	//}
	//result, err := DB.Exec(`delete from tag where id = ?`, tag.Id)
	//rowsAffected, _ := result.RowsAffected()
	//if rowsAffected == 0 {
	//	return errors.New("数据不存在")
	//}
}

func (this *TagService) List(c *gin.Context) ([]entity.Tag, int64) {
	param := c.Query("param")
	query := this.Model(entity.Tag{})

	if len(param) > 0 {
		query.Where("name like ?", "%"+param+"%")
	}
	var total int64
	query.Count(&total)
	var results []entity.Tag
	query.Debug().Scopes(utils.Paginate(c)).Order("id desc").Find(&results)
	return results, total
}
