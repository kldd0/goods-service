package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string `yaml:"env" env-default:"local"`
	DBUri    string `yaml:"db_uri" env-default:""`
	NATSAddr string `yaml:"nats_addr" env-default:"4222"`

	Redis      `yaml:"redis"`
	HTTPServer `yaml:"http_server"`
}

type Redis struct {
	Uri      string `yaml:"uri"`
	Password string `yaml:"pass"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed reading config: " + err.Error())
	}

	return &cfg
}

// Priority: flag > env > default.
func fetchConfigPath() string {
	var p string

	flag.StringVar(&p, "config", "", "path to config file")
	flag.Parse()

	if p == "" {
		p = os.Getenv("CONFIG_PATH")
	}

	return p
}

func (c Config) RedisUri() string {
	return c.Redis.Uri
}

func (c Config) RedisPass() string {
	return c.Redis.Password
}
