package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
)

type Config struct {
	Env         string `env:"ENV" env-default:"local"`
	ConnString  string `env:"CONN_STRING" env-default:"postgres://admin:Aitusa2025!@34.89.141.175:1005/complaints?sslmode=disable"`
	RedisClient RedisClient
	HTTPServer  HTTPServer
}

type RedisClient struct {
	Addr        string `env:"REDIS_ADDR" env-default:"127.0.0.1:6379"`
	User        string `env:"REDIS_USER" env-default:"default"`
	Password    string `env:"REDIS_PASS" env-default:""`
	DB          int    `env:"REDIS_DB" env-default:"0"`
	MaxRetries  int    `env:"REDIS_MAX_RETRIES" env-default:"3"`
	DialTimeout string `env:"REDIS_DIAL_TIMEOUT" env-default:"5s"`
	Timeout     string `env:"REDIS_TIMEOUT" env-default:"10s"`
}

type HTTPServer struct {
	Address     string `env:"HTTP_ADDRESS" env-default:"localhost:8080"`
	Timeout     string `env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout string `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	User        string `env:"HTTP_USER"`
	Password    string `env:"HTTP_PASSWORD"`
}

func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read config from env: %s", err)
	}
	return &cfg
}
