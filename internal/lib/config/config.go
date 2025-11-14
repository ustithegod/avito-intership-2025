package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string     `yaml:"env" env-default:"info"`
	HTTPServer HttpServer `yaml:"http_server" env-required:"true"`
}

type HttpServer struct {
	Address      string        `yaml:"address" env-default:"localhost:8080"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"5s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env-default:"10s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

// MustLoad panics if config can not be found.
func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is required")
	}

	if _, err := os.Stat(configPath); err != nil {
		panic("config file does not exist:" + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

// fetchConfigPath fetches config path from cmd flag or environment variable.
// flag > env > default.
// default = "".
func fetchConfigPath() string {
	var path string

	flag.StringVar(&path, "config", "", "Path to the configuration file")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}

	return path
}
