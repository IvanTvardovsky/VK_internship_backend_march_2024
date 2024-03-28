package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"sync"
	"vk_march_backend/internal/structures"
	"vk_march_backend/pkg/logger"
)

var cfg *structures.Config
var once sync.Once

func GetConfig() *structures.Config {
	once.Do(func() {
		logger.Log.Infoln("Reading app configuration...")
		cfg = &structures.Config{}
		if err := cleanenv.ReadConfig("./config.yml", cfg); err != nil {
			help, _ := cleanenv.GetDescription(cfg, nil)
			logger.Log.Errorln(help)
			logger.Log.Fatalln(err)
		}
	})
	return cfg
}
