package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"go_url_chortener_api/internal/env"
	"log"
	"os"
	"time"
)

type Config struct {
	Env        string     `yaml:"env"`
	HttpServer HttpServer `yaml:"http_server"`
	Storage    Storage    `yaml:"storage"`
}

type HttpServer struct {
	Address     string        `yaml:"address"`
	Port        string        `yaml:"port"`
	Timeout     time.Duration `yaml:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

type Storage struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Dbname   string `yaml:"dbname"`
	SslMode  string `yaml:"sslMode"`
	Password string `yaml:"password"`
}

func MustLoad(environment string) *Config {
	cfg := new(Config)
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal(err)
	}

	var cfgPath string

	switch environment {
	case env.EnvLocal:
		cfgPath = os.Getenv("LOCAL_CONFIG_PATH")
	case env.EnvDev:
		cfgPath = os.Getenv("DEV_CONFIG_PATH")
	}

	if _, err = os.Stat(cfgPath); os.IsNotExist(err) {
		log.Fatalf("file does not exists")
	}
	err = cleanenv.ReadConfig(cfgPath, cfg)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}
