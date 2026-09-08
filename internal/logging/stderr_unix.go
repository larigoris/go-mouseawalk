//go:build !windows

package logging

import (
	"fmt"
	"os"
	"syscall"
)

// redirectStderr направляет файловый дескриптор 2 в журнал.
//
// robotgo и gohook — обёртки над C, и об ошибках они сообщают записью прямо
// в stderr, минуя slog (например "Accessibility API is disabled!"). Без этого
// перехвата самая важная диагностика в файл журнала не попадает — особенно
// при запуске из Finder, где stderr вообще некуда выводить.
//
// Возвращает копию исходного stderr, чтобы записи slog по-прежнему были видны
// в терминале. Копию нужно вернуть в restoreStderr.
func redirectStderr(file *os.File) (*os.File, error) {
	saved, err := syscall.Dup(int(os.Stderr.Fd()))
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить stderr: %w", err)
	}

	if err := syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd())); err != nil {
		_ = syscall.Close(saved)
		return nil, fmt.Errorf("не удалось перенаправить stderr в журнал: %w", err)
	}

	return os.NewFile(uintptr(saved), "/dev/stderr"), nil
}

// restoreStderr возвращает дескриптор 2 на исходный поток.
func restoreStderr(console *os.File) {
	if console == nil || console == os.Stderr {
		return
	}

	_ = syscall.Dup2(int(console.Fd()), int(os.Stderr.Fd()))
	_ = console.Close()
}
