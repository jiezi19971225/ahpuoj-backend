package main

import (
	_ "ahpuoj/config"
	_ "ahpuoj/dao/mysql"
	_ "ahpuoj/dao/orm"
	_ "ahpuoj/dao/redis"
	"ahpuoj/router"
)

func main() {
	r := router.InitRouter()
	r.Run(":8080")
}
