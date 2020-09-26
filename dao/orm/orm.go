package orm

import (
	"ahpuoj/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"strings"
)

var ORM *gorm.DB

func init() {

	var err error
	dbcfg, _ := config.Conf.GetSection("mysql")
	path := strings.Join([]string{dbcfg["user"], ":", dbcfg["password"], "@tcp(", dbcfg["host"], ":", dbcfg["port"], ")/", dbcfg["database"], "?charset=utf8", "&parseTime=true"}, "")
	ORM, err = gorm.Open(mysql.Open(path), &gorm.Config{})
	if err != nil {
		log.Println("gorm init failed", err)
	}
}
