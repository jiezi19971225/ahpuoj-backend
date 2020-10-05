package controller

import (
	"ahpuoj/utils"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func GetSubmitStatistic(c *gin.Context) {
	// 这还是一段神奇的SQL 获得15天内累计提交的变化

	type StatisticUnit struct {
		Date  utils.JSONDate `json:"date"`
		Count int            `json:"count"`
	}

	var results []StatisticUnit
	ORM.Raw(`
	select  dualdate.date,count(*) count from 
	(select * from solution) s 
	right join  
	(select date_sub(curdate(), interval(cast(help_topic_id as signed integer)) day) date
	from mysql.help_topic
	where help_topic_id  <= 14)  dualdate 
	on date(s.in_date) <= dualdate.date 
	group by dualdate.date order by dualdate.date asc`).Scan(&results)
	log.Println(results)
	c.JSON(http.StatusOK, gin.H{
		"message":                 "获取个人信息成功",
		"recent_submit_statistic": results,
	})
}
