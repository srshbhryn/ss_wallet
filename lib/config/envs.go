package config

import (
	"os"

	"github.com/spf13/viper"
)

type environment string

const PROD = environment("PROD")
const DEV = environment("DEV")

func (e environment) Get() environment {
	return e
}

var Env = func() environment {
	if os.Getenv("ENV") == "DEV" {
		return DEV
	}
	return PROD
}()

var LogDir string

var ConfigFile string

func init() {
	initLogDir()
	initConfigFile()
}

func initLogDir() {
	if Env == DEV {
		LogDir = "./logs/"
	} else {
		LogDir = "/logs/"
	}
}

func initConfigFile() {
	ConfigFile = os.Getenv("CONFIG_PATH")
	if ConfigFile == "" {
		panic("no config file")
	}
	viper.SetConfigFile(ConfigFile)
	viper.ReadInConfig()
}
