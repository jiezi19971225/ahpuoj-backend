package utils

import (
	"ahpuoj/config"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Unknwon/goconfig"

	"github.com/gin-gonic/gin"
)

func IndexOf(array interface{}, val interface{}) (index int) {
	index = -1
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		{
			s := reflect.ValueOf(array)
			for i := 0; i < s.Len(); i++ {
				if reflect.DeepEqual(val, s.Index(i).Interface()) {
					index = i
					return
				}
			}
		}
	}
	return
}

func Consolelog(contents ...interface{}) {
	enviroment, _ := config.Conf.GetValue("project", "enviroment")
	if enviroment == "debug" {
		for _, v := range contents {
			fmt.Fprintln(gin.DefaultWriter, v)
		}
	}
}

func Int64to32(i64 int64) int {
	i32, _ := strconv.Atoi(strconv.FormatInt(i64, 10))
	return i32
}

func GetCurrentExecDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		Consolelog(err)
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}

func GetCurrentPath() string {
	_, filename, _, _ := runtime.Caller(1)

	return path.Dir(filename)
}

func GetRandomStringUseSeeds(seeds string, length int) string {
	bytes := []byte(seeds)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(seeds))])
	}
	return string(result)
}

func GetRandomString(length int) string {
	seeds := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return GetRandomStringUseSeeds(seeds, length)
}

func GetRandomPasswordString(length int) string {
	seeds := "23456789abcdefghjkmnpqrstuvwxyz"
	return GetRandomStringUseSeeds(seeds, length)
}

func SaveFile(file multipart.File, ext string, category string) (string, error) {
	uploadDir, _ := config.Conf.GetValue("project", "uploaddir")
	dateString := time.Now().Format("20060102150405")
	filename := dateString + GetRandomString(20) + ext
	filepath := uploadDir + "/" + category + "/" + filename
	storepath := "upload/" + category + "/" + filename
	os.MkdirAll(path.Dir(filepath), os.ModePerm)
	out, err := os.Create(filepath)
	defer out.Close()
	_, err = io.Copy(out, file)
	return storepath, err
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateDir(path string, mask os.FileMode) error {
	if isExist, _ := PathExists(path); isExist == false {
		return os.MkdirAll(path, mask)
	}
	return nil
}

func GetTestCfg(path string) *goconfig.ConfigFile {
	cfg, _ := goconfig.LoadConfigFile(path)
	return cfg
}

func ConvertTextImgUrl(origin string) string {
	server, _ := config.Conf.GetValue("project", "server")
	replaceTo := `<img src="` + server + "/"
	reg := regexp.MustCompile(`<img src="`)
	res := reg.ReplaceAllString(origin, replaceTo)
	return res
}

func EngNumToInt(engNum string) (int, error) {
	num := 0
	for _, v := range engNum {
		if !unicode.IsLetter(v) {
			return num, errors.New("格式错误")
		}
		if unicode.IsUpper(v) {
			num = num*26 + (int(v) - 64)
		} else {
			num = num*26 + (int(v) - 96)
		}
	}
	return num, nil
}

func CheckError(c *gin.Context, err error, msg string) error {
	if err != nil {
		Consolelog(err.Error())
		var message string
		if msg != "" {
			message = msg
		} else {
			message = err.Error()
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": message,
			"show":    true,
		})
		return err
	} else {
		return nil
	}
}
