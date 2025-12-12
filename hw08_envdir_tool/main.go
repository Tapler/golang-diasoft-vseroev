package main

import (
	"fmt"
	"os"
)

func main() {
	// Проверяем количество аргументов
	// Формат: go-envdir /path/to/env/dir command arg1 arg2 ...
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "%v: Usage: go-envdir /path/to/env/dir command [args...]\n", ErrInvalidArguments)
		os.Exit(1)
	}

	envDir := os.Args[1]
	cmdArgs := os.Args[2:]

	// Читаем переменные окружения из директории
	env, err := ReadDir(envDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading environment directory: %v\n", err)
		os.Exit(1)
	}

	// Запускаем команду с новым окружением
	returnCode := RunCmd(cmdArgs, env)
	os.Exit(returnCode)
}
