package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrCopyToSameFile        = errors.New("copy to the same file")
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
	ErrNegativeOffset        = errors.New("negative offset")
	ErrNegativeLimit         = errors.New("negative limit")
	ErrCopyFailed            = errors.New("data copy failed")
)

func Copy(fromPath, toPath string, offset, limit int64) (retErr error) {
	// Проверка путей на корректность и клонов
	same, err := isSameFile(fromPath, toPath)
	if err != nil {
		return fmt.Errorf("path verification failed: %w", err)
	}
	if same {
		return fmt.Errorf("%w: %q", ErrCopyToSameFile, fromPath)
	}

	// Првоерка флагов
	if offset < 0 {
		return fmt.Errorf("%w: %d", ErrNegativeOffset, offset)
	}
	if limit < 0 {
		return fmt.Errorf("%w: %d", ErrNegativeLimit, limit)
	}

	// Открываем файл для чтения
	fromFile, err := os.Open(fromPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if err := fromFile.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("failed to close source file: %w", err)
		}
	}()

	// Получаем информацию о файле
	fromFileInfo, err := fromFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Проверяем, что это обычный файл
	if !fromFileInfo.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	fileSize := fromFileInfo.Size()

	// Проверяем корректность смещения
	if offset > fileSize {
		return ErrOffsetExceedsFileSize
	}

	// Устанавливаем позицию чтения
	if _, err = fromFile.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("seek error: %w", err)
	}

	// Вычисляем сколько будем читать
	bytesToCopy := fileSize - offset
	if limit > 0 && bytesToCopy > limit {
		bytesToCopy = limit
	}

	// Создаем файл для записи
	toFile, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		closeErr := toFile.Close()

		if closeErr != nil && retErr == nil {
			retErr = fmt.Errorf("failed to close destination file: %w", closeErr)
		}

		if errors.Is(retErr, ErrCopyFailed) {
			// Удаляем файл в случае ошибки копирования
			removeErr := os.Remove(toPath)
			if removeErr != nil && !os.IsNotExist(removeErr) {
				// Если не удалось удалить, добавляем эту информацию к основной ошибке
				retErr = fmt.Errorf("%w (and failed to remove partial file: %w)", retErr, removeErr)
			}
		}
	}()

	// Создаем progress bar
	bar := pb.Full.Start64(bytesToCopy)
	defer bar.Finish()

	// Копируем данные с прогрессом
	_, err = io.Copy(toFile, bar.NewProxyReader(io.LimitReader(fromFile, bytesToCopy)))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCopyFailed, err)
	}

	// Устанавливаем права доступа
	if err := os.Chmod(toPath, fromFileInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

func isSameFile(path1, path2 string) (bool, error) {
	// Нормализуем пути (убираем ./ и ../)
	absPath1, err := filepath.Abs(path1)
	if err != nil {
		return false, fmt.Errorf("failed to resolve path %q: %w", path1, err)
	}

	absPath2, err := filepath.Abs(path2)
	if err != nil {
		return false, fmt.Errorf("failed to resolve path %q: %w", path2, err)
	}

	// Быстрая проверка: если пути идентичны после нормализации
	if absPath1 == absPath2 {
		return true, nil
	}

	// Проверяем существование первого файла
	info1, err := os.Stat(absPath1)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Файл не существует - значит не совпадают
		}
		return false, fmt.Errorf("failed to access %q: %w", absPath1, err)
	}

	// Проверяем существование второго файла
	info2, err := os.Stat(absPath2)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Файл не существует - значит не совпадают
		}
		return false, fmt.Errorf("failed to access %q: %w", absPath2, err)
	}

	// Сравниваем файлы (учитывает hard links и симлинки)
	return os.SameFile(info1, info2), nil
}
