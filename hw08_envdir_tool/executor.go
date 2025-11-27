package main

import (
	"errors"
	"os"
	"os/exec"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return -1
	}

	// Создаем команду
	// #nosec G204
	command := exec.Command(cmd[0], cmd[1:]...)

	// Получаем текущее окружение
	currentEnv := os.Environ()

	// Применяем изменения из env
	envMap := make(map[string]string)

	// Парсим текущее окружение в map
	for _, e := range currentEnv {
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				envMap[e[:i]] = e[i+1:]
				break
			}
		}
	}

	// Применяем изменения
	for name, value := range env {
		if value.NeedRemove {
			delete(envMap, name)
		} else {
			envMap[name] = value.Value
		}
	}

	// Собираем окружение обратно
	newEnv := make([]string, 0, len(envMap))
	for k, v := range envMap {
		newEnv = append(newEnv, k+"="+v)
	}

	command.Env = newEnv

	// Пробрасываем стандартные потоки
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// Запускаем команду
	err := command.Run()
	if err != nil {
		// Если команда завершилась с ошибкой, возвращаем её код выхода
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		// Если произошла другая ошибка
		return -1
	}

	return 0
}
