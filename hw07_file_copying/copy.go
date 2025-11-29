package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	// Проверяем отрицательные значения
	if offset < 0 || limit < 0 {
		return fmt.Errorf("offset and limit must be non-negative")
	}

	// Открываем исходный файл для чтения
	srcFile, err := os.OpenFile(fromPath, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Получаем информацию о файле
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Проверяем, что файл является поддерживаемым
	if !srcInfo.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	// Проверяем, что offset не превышает размер файла
	if offset > srcInfo.Size() {
		return ErrOffsetExceedsFileSize
	}

	// Перемещаемся к нужному offset
	if offset > 0 {
		_, err = srcFile.Seek(offset, io.SeekStart)
		if err != nil {
			return fmt.Errorf("failed to seek: %w", err)
		}
	}

	// Вычисляем количество байт для копирования
	bytesToCopy := srcInfo.Size() - offset
	if limit > 0 && limit < bytesToCopy {
		bytesToCopy = limit
	}

	// Создаем файл назначения с правами доступа 0644 (rw-r--r--)
	dstFile, err := os.OpenFile(toPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Создаем прогресс-бар
	bar := pb.Full.Start64(bytesToCopy)
	defer bar.Finish()

	// Создаем reader с прогресс-баром
	barReader := bar.NewProxyReader(srcFile)

	// Копируем данные
	_, err = io.CopyN(dstFile, barReader, bytesToCopy)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to copy: %w", err)
	}

	return nil
}
