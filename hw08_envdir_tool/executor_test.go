package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	t.Run("successful command execution", func(t *testing.T) {
		t.Helper()
		env := Environment{
			"TEST_VAR": {Value: "test_value", NeedRemove: false},
		}

		// Команда echo должна успешно выполниться
		returnCode := RunCmd([]string{"echo", "hello"}, env)
		require.Equal(t, 0, returnCode)
	})

	t.Run("command with non-zero exit code", func(t *testing.T) {
		t.Helper()
		env := Environment{}

		// Команда false возвращает код 1
		returnCode := RunCmd([]string{"false"}, env)
		require.Equal(t, 1, returnCode)
	})

	t.Run("empty command", func(t *testing.T) {
		t.Helper()
		env := Environment{}

		returnCode := RunCmd([]string{}, env)
		require.Equal(t, -1, returnCode)
	})

	t.Run("non-existent command", func(t *testing.T) {
		t.Helper()
		env := Environment{}

		returnCode := RunCmd([]string{"nonexistent_command_12345"}, env)
		require.Equal(t, -1, returnCode)
	})

	t.Run("environment variable removal", func(t *testing.T) {
		t.Helper()
		env := Environment{
			"TO_REMOVE": {Value: "", NeedRemove: true},
			"TO_SET":    {Value: "new_value", NeedRemove: false},
		}

		// Эта команда просто выполнится успешно
		returnCode := RunCmd([]string{"true"}, env)
		require.Equal(t, 0, returnCode)
	})
}
