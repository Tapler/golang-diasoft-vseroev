package main

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

func TestCopy(t *testing.T) {
	tests := []struct {
		name        string
		offset      int64
		limit       int64
		expectedOut string
		wantErr     bool
		errType     error
	}{
		{
			name:        "copy full file (offset=0, limit=0)",
			offset:      0,
			limit:       0,
			expectedOut: "testdata/out_offset0_limit0.txt",
			wantErr:     false,
		},
		{
			name:        "copy with limit 10",
			offset:      0,
			limit:       10,
			expectedOut: "testdata/out_offset0_limit10.txt",
			wantErr:     false,
		},
		{
			name:        "copy with limit 1000",
			offset:      0,
			limit:       1000,
			expectedOut: "testdata/out_offset0_limit1000.txt",
			wantErr:     false,
		},
		{
			name:        "copy with limit exceeding file size",
			offset:      0,
			limit:       10000,
			expectedOut: "testdata/out_offset0_limit10000.txt",
			wantErr:     false,
		},
		{
			name:        "copy with offset 100 and limit 1000",
			offset:      100,
			limit:       1000,
			expectedOut: "testdata/out_offset100_limit1000.txt",
			wantErr:     false,
		},
		{
			name:        "copy with offset 6000 and limit 1000",
			offset:      6000,
			limit:       1000,
			expectedOut: "testdata/out_offset6000_limit1000.txt",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем временный файл для вывода
			tmpFile, err := os.CreateTemp("", "copy_test_*.txt")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			tmpPath := tmpFile.Name()
			tmpFile.Close()
			defer os.Remove(tmpPath)

			// Выполняем копирование
			err = Copy("testdata/input.txt", tmpPath, tt.offset, tt.limit)

			// Проверяем ошибку
			if (err != nil) != tt.wantErr {
				t.Errorf("Copy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("Copy() error = %v, expected error type %v", err, tt.errType)
				}
				return
			}

			// Сравниваем содержимое файлов
			assertFilesEqual(t, tmpPath, tt.expectedOut)
		})
	}
}

func TestCopyErrors(t *testing.T) {
	t.Run("offset exceeds file size", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "copy_test_*.txt")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath)

		err = Copy("testdata/input.txt", tmpPath, 100000, 0)
		if !errors.Is(err, ErrOffsetExceedsFileSize) {
			t.Errorf("expected ErrOffsetExceedsFileSize, got %v", err)
		}
	})

	t.Run("offset equals file size", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "copy_test_*.txt")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath)

		// Получаем размер файла
		srcInfo, err := os.Stat("testdata/input.txt")
		if err != nil {
			t.Fatalf("failed to stat input file: %v", err)
		}

		// При offset == размеру файла должно скопироваться 0 байт без ошибки
		err = Copy("testdata/input.txt", tmpPath, srcInfo.Size(), 0)
		if err != nil {
			t.Errorf("expected no error for offset equal to file size, got %v", err)
		}

		// Проверяем, что файл результата пустой
		resultContent, err := os.ReadFile(tmpPath)
		if err != nil {
			t.Fatalf("failed to read result file: %v", err)
		}
		if len(resultContent) != 0 {
			t.Errorf("expected empty result file, got %d bytes", len(resultContent))
		}
	})

	t.Run("source file does not exist", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "copy_test_*.txt")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath)

		err = Copy("testdata/nonexistent.txt", tmpPath, 0, 0)
		if err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
	})

	t.Run("negative offset", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "copy_test_*.txt")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath)

		err = Copy("testdata/input.txt", tmpPath, -1, 0)
		if err == nil {
			t.Error("expected error for negative offset, got nil")
		}
	})

	t.Run("negative limit", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "copy_test_*.txt")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath)

		err = Copy("testdata/input.txt", tmpPath, 0, -1)
		if err == nil {
			t.Error("expected error for negative limit, got nil")
		}
	})

	t.Run("unsupported file type (directory)", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "copy_test_*.txt")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath)

		// Пытаемся скопировать директорию как файл
		err = Copy("testdata", tmpPath, 0, 0)
		if !errors.Is(err, ErrUnsupportedFile) {
			t.Errorf("expected ErrUnsupportedFile, got %v", err)
		}
	})
}

func assertFilesEqual(t *testing.T, gotPath, expectedPath string) {
	t.Helper()

	gotContent, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("failed to read result file: %v", err)
	}

	expectedContent, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read expected file: %v", err)
	}

	if !bytes.Equal(gotContent, expectedContent) {
		t.Errorf("file content mismatch: got %d bytes, expected %d bytes", len(gotContent), len(expectedContent))
	}
}
