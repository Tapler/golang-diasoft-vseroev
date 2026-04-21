package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Logger    LoggerConf    `toml:"logger"`
	Database  DatabaseConf  `toml:"database"`
	Kafka     KafkaConf     `toml:"kafka"`
	Scheduler SchedulerConf `toml:"scheduler"`
}

type LoggerConf struct {
	Level string `toml:"level"`
}

type DatabaseConf struct {
	DSN string `toml:"dsn"`
}

type KafkaConf struct {
	Brokers []string `toml:"brokers"`
	Topic   string   `toml:"topic"`
}

type SchedulerConf struct {
	ScanInterval string `toml:"scan_interval"`
}

func NewConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if config.Logger.Level == "" {
		config.Logger.Level = "INFO"
	}
	if config.Kafka.Topic == "" {
		config.Kafka.Topic = "notifications"
	}
	if config.Scheduler.ScanInterval == "" {
		config.Scheduler.ScanInterval = "1m"
	}

	return &config, nil
}
