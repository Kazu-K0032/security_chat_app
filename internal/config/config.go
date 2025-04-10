package config

import (
	"log"

	utils "security_chat_app/internal/utils/log"

	"gopkg.in/go-ini/ini.v1"
)

type ConfigList struct {
	Port                      string
	LogFile                   string
	Static                    string
	FirebaseServiceAccountKey string
}

var Config ConfigList

func init() {
	LoadConfig()
	utils.LoggingSettings(Config.LogFile)
}

func LoadConfig() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalln(err)
	}
	Config = ConfigList{
		Port:                      cfg.Section("web").Key("port").MustString("8080"),
		LogFile:                   cfg.Section("web").Key("logfile").String(),
		Static:                    cfg.Section("web").Key("static").String(),
		FirebaseServiceAccountKey: cfg.Section("firebase").Key("serviceAccountKey").String(),
	}
}
