package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env  string     `yaml:"env" env:"ENV" env-default:"local"`
	HTTP HTTPConfig `yaml:"http"`
	DB   DBConfig   `yaml:"db"`
}

type DBConfig struct {
	DSN string `yaml:"dsn" env:"DB_DSN" env-required:"true"`
}

type HTTPConfig struct {
	Port              string        `yaml:"port"               env:"HTTP_PORT"               env-default:"8080"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" env:"HTTP_READ_HEADER_TIMEOUT" env-default:"5s"`
	ReadTimeout       time.Duration `yaml:"read_timeout"        env:"HTTP_READ_TIMEOUT"        env-default:"10s"`
	WriteTimeout      time.Duration `yaml:"write_timeout"       env:"HTTP_WRITE_TIMEOUT"       env-default:"30s"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"        env:"HTTP_IDLE_TIMEOUT"        env-default:"1m"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout"    env:"HTTP_SHUTDOWN_TIMEOUT"    env-default:"10s"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		configPath = "config/local.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
