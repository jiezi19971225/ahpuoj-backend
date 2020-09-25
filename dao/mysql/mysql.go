package mysql

import (
	"ahpuoj/config"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func init() {
	var err error
	dbcfg, _ := config.Conf.GetSection("mysql")
	path := strings.Join([]string{dbcfg["user"], ":", dbcfg["password"], "@tcp(", dbcfg["host"], ":", dbcfg["port"], ")/", dbcfg["database"], "?charset=utf8"}, "")
	DB, err = sqlx.Open("mysql", path)
	DB.SetMaxIdleConns(100)
	DB.SetConnMaxLifetime(2 * time.Minute)
	if err != nil {
		log.Println("m=GetPool,msg=connection has failed", err)
	}
}
