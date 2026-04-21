package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger   LoggerConf   `toml:"logger"`
	HTTP     HTTPConf     `toml:"http"`
	Database DatabaseConf `toml:"database"`
}

type LoggerConf struct {
	Level string `toml:"level"`
}

type HTTPConf struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
}

type DatabaseConf struct {
	Type string `toml:"type"` // "memory" или "sql"
	DSN  string `toml:"dsn"`  // Строка подключения для SQL
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

	// Установка значений по умолчанию
	if config.Logger.Level == "" {
		config.Logger.Level = "INFO"
	}
	if config.HTTP.Host == "" {
		config.HTTP.Host = "localhost"
	}
	if config.HTTP.Port == "" {
		config.HTTP.Port = "8080"
	}
	if config.Database.Type == "" {
		config.Database.Type = "memory"
	}

	return &config, nil
}
