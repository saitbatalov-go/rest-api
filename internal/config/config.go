package config

import (
	"log"
	"os"

	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:":8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	User        string        `yaml:"user" env-required:"true"`
	Password    string        `yaml:"password" env-required:"true" env:"HTTP-SERVER_PASSWORD"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/local.yaml"
		log.Printf("CONFIG_PATH не установлен, используется значение по умолчанию: %s", configPath)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("CONFIG_PATH does not exist", configPath)
	}

	var config Config

	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	return &config
}
