package config

import "github.com/Unknwon/goconfig"

var Conf *goconfig.ConfigFile

func init() {
	configFilePath := "config/config.ini"
	Conf, _ = goconfig.LoadConfigFile(configFilePath)
}
