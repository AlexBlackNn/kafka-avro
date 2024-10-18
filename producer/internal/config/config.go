package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type KafkaConfig struct {
	KafkaURL          string `yaml:"kafkaUrl" env-required:"true"`
	SchemaRegistryURL string `yaml:"schemaRegistryURL" env-required:"true"`
}

type Config struct {
	// without this param will be used "local" as param value
	Env   string      `yaml:"env" env-default:"local"`
	Kafka KafkaConfig `yaml:"kafka"`
}

func (c *Config) String() string {
	return fmt.Sprintf(
		"env: %s, kafka url %s, schema registry url %s",
		c.Env, c.Kafka.KafkaURL, c.Kafka.SchemaRegistryURL,
	)
}

func New() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadByPath(configPath)
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exists: " + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config " + err.Error())
	}
	return &cfg
}

// fetchConfigPath fetches config path from command line flag or env var
// Priority: flag -> env -> default
// Default value is empty string
func fetchConfigPath() string {
	var res string
	// --config="path/to/config.yaml"
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}
	return res
}
