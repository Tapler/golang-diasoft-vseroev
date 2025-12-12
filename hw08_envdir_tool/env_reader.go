package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

var (
	ErrInvalidArguments = errors.New("invalid arguments")
	ErrDirectoryRead    = errors.New("failed to read directory")
)

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDirectoryRead, err)
	}

	env := make(Environment)

	for _, entry := range entries {
		// Пропускаем директории
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Имя не должно содержать '='
		if strings.Contains(name, "=") {
			continue
		}

		filePath := filepath.Join(dir, name)

		// Читаем содержимое файла
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", name, err)
		}

		// Если файл пустой, помечаем переменную для удаления
		if len(content) == 0 {
			env[name] = EnvValue{
				Value:      "",
				NeedRemove: true,
			}
			continue
		}

		// Берем только первую строку
		lines := bytes.SplitN(content, []byte("\n"), 2)
		value := lines[0]

		// Заменяем терминальные нули на перевод строки
		value = bytes.ReplaceAll(value, []byte{0x00}, []byte("\n"))

		// Удаляем пробелы и табуляцию в конце
		valueStr := string(value)
		valueStr = strings.TrimRight(valueStr, " \t")

		env[name] = EnvValue{
			Value:      valueStr,
			NeedRemove: false,
		}
	}

	return env, nil
}
