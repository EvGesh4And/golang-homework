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
	ErrNegativeOffset        = errors.New("negative offset")
	ErrNegativeLimit         = errors.New("negative limit")
	ErrCopyToSameFile        = errors.New("copy to the same file")
)

func Copy(fromPath, toPath string, offset, limit int64) (retErr error) {

	if fromPath == toPath {
		return fmt.Errorf("%w: %q", ErrCopyToSameFile, fromPath)
	}
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
		// Сначала пробуем синхронизировать данные на диск
		syncErr := toFile.Sync()
		closeErr := toFile.Close()

		if (syncErr != nil || closeErr != nil) && retErr == nil {
			retErr = fmt.Errorf("failed to sync/close destination file (sync: %w, close: %w)", syncErr, closeErr)
		}
	}()

	// Создаем progress bar
	bar := pb.Full.Start64(bytesToCopy)
	defer bar.Finish()

	// Копируем данные с прогрессом
	_, err = io.Copy(toFile, bar.NewProxyReader(io.LimitReader(fromFile, bytesToCopy)))
	if err != nil {
		// Удаляем файл в случае ошибки копирования
		removeErr := os.Remove(toPath)
		if removeErr != nil && !os.IsNotExist(removeErr) {
			// Если не удалось удалить, добавляем эту информацию к основной ошибке
			return fmt.Errorf("failed: %w (and failed to remove partial file: %w)", err, removeErr)
		}
		return fmt.Errorf("failed: %w", err)
	}

	// Устанавливаем права доступа
	if err := os.Chmod(toPath, fromFileInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}
