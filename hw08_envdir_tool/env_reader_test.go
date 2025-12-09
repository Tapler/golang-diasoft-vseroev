package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadDir(t *testing.T) {
	t.Run("read testdata env directory", testReadTestdataEnv)
	t.Run("non-existent directory", testNonExistentDirectory)
	t.Run("empty directory", testEmptyDirectory)
	t.Run("file with equals sign in name", testFileWithEqualsSign)
	t.Run("file with null bytes", testFileWithNullBytes)
	t.Run("file with trailing spaces and tabs", testFileWithTrailingSpaces)
	t.Run("subdirectories are ignored", testSubdirectoriesIgnored)
}

func testReadTestdataEnv(t *testing.T) {
	t.Helper()
	env, err := ReadDir("testdata/env")
	require.NoError(t, err)
	require.NotNil(t, env)

	// Проверяем BAR - первая строка "bar", вторая строка игнорируется
	require.Contains(t, env, "BAR")
	require.Equal(t, "bar", env["BAR"].Value)
	require.False(t, env["BAR"].NeedRemove)

	// Проверяем FOO - содержит пробелы в начале и нулевой байт в первой строке
	require.Contains(t, env, "FOO")
	// Файл FOO содержит "   foo\x00with new line", нулевой байт заменяется на \n
	require.Equal(t, "   foo\nwith new line", env["FOO"].Value)
	require.False(t, env["FOO"].NeedRemove)

	// Проверяем HELLO
	require.Contains(t, env, "HELLO")
	require.Equal(t, "\"hello\"", env["HELLO"].Value)
	require.False(t, env["HELLO"].NeedRemove)

	// Проверяем UNSET - пустой файл (0 байт)
	require.Contains(t, env, "UNSET")
	require.Empty(t, env["UNSET"].Value)
	require.True(t, env["UNSET"].NeedRemove)

	// Проверяем EMPTY - файл с одним пробелом и переводом строки
	require.Contains(t, env, "EMPTY")
	require.Empty(t, env["EMPTY"].Value)
	require.False(t, env["EMPTY"].NeedRemove)
}

func testNonExistentDirectory(t *testing.T) {
	t.Helper()
	_, err := ReadDir("testdata/nonexistent")
	require.Error(t, err)
}

func testEmptyDirectory(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()

	env, err := ReadDir(tmpDir)
	require.NoError(t, err)
	require.Empty(t, env)
}

func testFileWithEqualsSign(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "VAR=VALUE"), []byte("test"), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "NORMAL"), []byte("value"), 0o644)
	require.NoError(t, err)

	env, err := ReadDir(tmpDir)
	require.NoError(t, err)

	require.NotContains(t, env, "VAR=VALUE")
	require.Contains(t, env, "NORMAL")
}

func testFileWithNullBytes(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()

	content := []byte("test\x00value\x00end")
	err := os.WriteFile(filepath.Join(tmpDir, "NULLS"), content, 0o644)
	require.NoError(t, err)

	env, err := ReadDir(tmpDir)
	require.NoError(t, err)

	require.Contains(t, env, "NULLS")
	require.Equal(t, "test\nvalue\nend", env["NULLS"].Value)
}

func testFileWithTrailingSpaces(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()

	content := []byte("value  \t \t  ")
	err := os.WriteFile(filepath.Join(tmpDir, "SPACES"), content, 0o644)
	require.NoError(t, err)

	env, err := ReadDir(tmpDir)
	require.NoError(t, err)

	require.Contains(t, env, "SPACES")
	require.Equal(t, "value", env["SPACES"].Value)
}

func testSubdirectoriesIgnored(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()

	err := os.Mkdir(filepath.Join(tmpDir, "subdir"), 0o755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "VAR"), []byte("value"), 0o644)
	require.NoError(t, err)

	env, err := ReadDir(tmpDir)
	require.NoError(t, err)

	require.NotContains(t, env, "subdir")
	require.Contains(t, env, "VAR")
}
