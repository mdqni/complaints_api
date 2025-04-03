package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	ConnString  string `yaml:"connString" env-required:"true"`
	RedisClient `yaml:"redisClient"`
	HTTPServer  `yaml:"http_server"`
}
type RedisClient struct {
	Addr        string        `yaml:"addr" env-default:"127.0.0.1:6379"`
	User        string        `yaml:"user"`
	Password    string        `yaml:"pass" env:"REDIS_PASSWORD" default:""`
	DB          int           `yaml:"db" default:"0"`
	MaxRetries  int           `yaml:"max_retries"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	Timeout     time.Duration `yaml:"timeout"`
}
type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	User        string        `yaml:"user" env-required:"true"`
	Password    string        `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"`
}

func MustLoad() *Config {
	defaultConfigPath := "./config/local.yaml"
	if err := os.Setenv("CONFIG_PATH", defaultConfigPath); err != nil {
		fmt.Println("Error setting environment variable:", err)
		return nil
	}
	if err := os.Setenv("CGO_ENABLED", "1"); err != nil {
		fmt.Println("Error setting environment variable:", err)
		return nil
	}
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	//check if fileExist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("CONFIG_PATH does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	return &cfg
}
