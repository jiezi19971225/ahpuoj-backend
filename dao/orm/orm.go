package orm

import (
	"ahpuoj/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"log"
	"strings"
	"time"
)

var ORM *gorm.DB

func init() {
	var err error
	dbMasterCfg, _ := config.Conf.GetSection("mysql")
	dbSlaveCfg, _ := config.Conf.GetSection("mysql_slave")
	// 需要设置时区，否则解析时采用 +0 时区
	paramString := "?charset=utf8&parseTime=true&loc=Local"
	path := strings.Join([]string{dbMasterCfg["user"], ":", dbMasterCfg["password"], "@tcp(", dbMasterCfg["host"], ":", dbMasterCfg["port"], ")/", dbMasterCfg["database"], paramString}, "")
	slavePath := strings.Join([]string{dbSlaveCfg["user"], ":", dbSlaveCfg["password"], "@tcp(", dbSlaveCfg["host"], ":", dbSlaveCfg["port"], ")/", dbSlaveCfg["database"], paramString}, "")
	ORM, err = gorm.Open(mysql.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	ORM.Use(dbresolver.Register(dbresolver.Config{
		Replicas: []gorm.Dialector{mysql.Open(slavePath)},
	}))
	if err != nil {
		log.Println("gorm init failed", err)
	}

	sqlDB, err := ORM.DB()
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(100)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(500)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(2 * time.Minute)
}
