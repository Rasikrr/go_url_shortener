package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
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

func MustLoad() *Config {
	cfg := new(Config)
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal(err)
	}
	cfgPath := os.Getenv("CONFIG_PATH")
	if _, err = os.Stat(cfgPath); os.IsNotExist(err) {
		log.Fatalf("file does not exists")
	}
	err = cleanenv.ReadConfig(cfgPath, cfg)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}
