package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

type Config struct {
	Env        string `env:"ENV" env_default:"local"`
	Host       string `env:"DB_HOST" env_default:"localhost"`
	Port       string `env:"DB_PORT" env_default:"27017"`
	DbName     string `env:"DB_NAME" env_default:"lab"`
	DbUser     string `env:"DB_USER" env_default:"paincake"`
	DbPassword string `env:"DB_PASSWORD" env_default:"biliberda9999"`
}

func MustLoad() *Config {
	cfgPath := os.Getenv("LAB_CFG_PATH")
	cfgPath = "D:\\GoProjects\\first-admin-lab\\configs\\local.env"
	if cfgPath == "" {
		//	log.Fatal("LAB_CFG_PATH env variable is not set")

	}
	if _, err := os.Stat(cfgPath); err != nil {
		log.Fatalf("error opening config file:%s", err)
	}
	var cfg Config
	err := cleanenv.ReadConfig(cfgPath, &cfg)
	if err != nil {
		log.Fatalf("error parsing config file:%s", err)
	}
	return &cfg
}
