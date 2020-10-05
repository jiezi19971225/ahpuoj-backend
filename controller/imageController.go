package controller

import (
	"ahpuoj/utils"
	"path"

	"github.com/gin-gonic/gin"
)

func StoreImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	ext := path.Ext(header.Filename)
	if err != nil {
		panic(err)
	}
	url, err := utils.SaveFile(file, ext, "images")
	if err != nil {
		panic(err)
	}
	c.JSON(200, gin.H{
		"message": "图片上传成功",
		"url":     url,
	})
}
