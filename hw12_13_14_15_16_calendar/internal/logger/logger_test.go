package logger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel LogLevel
	}{
		{
			name:          "DEBUG level",
			level:         "DEBUG",
			expectedLevel: LevelDebug,
		},
		{
			name:          "INFO level",
			level:         "INFO",
			expectedLevel: LevelInfo,
		},
		{
			name:          "WARN level",
			level:         "WARN",
			expectedLevel: LevelWarn,
		},
		{
			name:          "ERROR level",
			level:         "ERROR",
			expectedLevel: LevelError,
		},
		{
			name:          "lowercase debug",
			level:         "debug",
			expectedLevel: LevelDebug,
		},
		{
			name:          "mixed case Info",
			level:         "Info",
			expectedLevel: LevelInfo,
		},
		{
			name:          "unknown level defaults to INFO",
			level:         "UNKNOWN",
			expectedLevel: LevelInfo,
		},
		{
			name:          "empty level defaults to INFO",
			level:         "",
			expectedLevel: LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level)
			require.NotNil(t, logger)
			require.Equal(t, tt.expectedLevel, logger.level)
		})
	}
}

func TestLogger_AllLevels(t *testing.T) {
	t.Run("DEBUG level logs all messages", func(t *testing.T) {
		var buf bytes.Buffer
		logger := New("DEBUG")
		logger.logger.SetOutput(&buf)
		logger.logger.SetFlags(0)

		logger.Debug("debug msg")
		logger.Info("info msg")
		logger.Warn("warn msg")
		logger.Error("error msg")

		output := buf.String()
		require.Contains(t, output, "[DEBUG] debug msg")
		require.Contains(t, output, "[INFO] info msg")
		require.Contains(t, output, "[WARN] warn msg")
		require.Contains(t, output, "[ERROR] error msg")
	})

	t.Run("INFO level logs info, warn, error", func(t *testing.T) {
		var buf bytes.Buffer
		logger := New("INFO")
		logger.logger.SetOutput(&buf)
		logger.logger.SetFlags(0)

		logger.Debug("debug msg")
		logger.Info("info msg")
		logger.Warn("warn msg")
		logger.Error("error msg")

		output := buf.String()
		require.NotContains(t, output, "[DEBUG]")
		require.Contains(t, output, "[INFO] info msg")
		require.Contains(t, output, "[WARN] warn msg")
		require.Contains(t, output, "[ERROR] error msg")
	})

	t.Run("WARN level logs warn and error", func(t *testing.T) {
		var buf bytes.Buffer
		logger := New("WARN")
		logger.logger.SetOutput(&buf)
		logger.logger.SetFlags(0)

		logger.Debug("debug msg")
		logger.Info("info msg")
		logger.Warn("warn msg")
		logger.Error("error msg")

		output := buf.String()
		require.NotContains(t, output, "[DEBUG]")
		require.NotContains(t, output, "[INFO]")
		require.Contains(t, output, "[WARN] warn msg")
		require.Contains(t, output, "[ERROR] error msg")
	})

	t.Run("ERROR level logs only error", func(t *testing.T) {
		var buf bytes.Buffer
		logger := New("ERROR")
		logger.logger.SetOutput(&buf)
		logger.logger.SetFlags(0)

		logger.Debug("debug msg")
		logger.Info("info msg")
		logger.Warn("warn msg")
		logger.Error("error msg")

		output := buf.String()
		require.NotContains(t, output, "[DEBUG]")
		require.NotContains(t, output, "[INFO]")
		require.NotContains(t, output, "[WARN]")
		require.Contains(t, output, "[ERROR] error msg")
	})
}

func TestLogger_MessageFormatting(t *testing.T) {
	t.Run("messages with special characters", func(t *testing.T) {
		var buf bytes.Buffer
		logger := New("INFO")
		logger.logger.SetOutput(&buf)
		logger.logger.SetFlags(0)

		specialMsg := "message with \n newline and \t tab"
		logger.Info(specialMsg)

		output := buf.String()
		require.Contains(t, output, specialMsg)
	})

	t.Run("empty message", func(t *testing.T) {
		var buf bytes.Buffer
		logger := New("INFO")
		logger.logger.SetOutput(&buf)
		logger.logger.SetFlags(0)

		logger.Info("")

		output := buf.String()
		require.Contains(t, output, "[INFO]")
	})

	t.Run("long message", func(t *testing.T) {
		var buf bytes.Buffer
		logger := New("INFO")
		logger.logger.SetOutput(&buf)
		logger.logger.SetFlags(0)

		longMsg := strings.Repeat("a", 1000)
		logger.Info(longMsg)

		output := buf.String()
		require.Contains(t, output, "[INFO]")
		require.Contains(t, output, longMsg)
	})
}
