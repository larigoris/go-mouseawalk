// Package logging настраивает файловый журнал GOwalk.
//
// Каждый запуск пишет отдельный файл в каталог, специфичный для ОС:
//
//	macOS    ~/Library/Logs/GOwalk
//	Linux    $XDG_STATE_HOME/GOwalk/logs  (или ~/.local/state/GOwalk/logs)
//	Windows  %LOCALAPPDATA%\GOwalk\Logs
//
// Каталог можно переопределить переменной окружения GOWALK_LOG_DIR.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

const (
	// appName — имя каталога журналов.
	appName = "GOwalk"
	// fileNameLayout — метка времени в имени файла; сортируется лексикографически.
	fileNameLayout = "20060102-150405"
	// keepFiles — сколько последних журналов оставлять при очистке.
	keepFiles = 10
)

// Session — журнал одного запуска программы.
type Session struct {
	Logger *slog.Logger
	Path   string

	file    *os.File
	console *os.File
	start   time.Time
}

// Dir возвращает каталог журналов для текущей ОС.
func Dir() (string, error) {
	if custom := os.Getenv("GOWALK_LOG_DIR"); custom != "" {
		return custom, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("не удалось определить домашний каталог: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Logs", appName), nil
	case "windows":
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(base, appName, "Logs"), nil
	default:
		base := os.Getenv("XDG_STATE_HOME")
		if base == "" {
			base = filepath.Join(home, ".local", "state")
		}
		return filepath.Join(base, appName, "logs"), nil
	}
}

// Start создаёт каталог журналов, открывает файл под текущий запуск и пишет
// в него заголовок сессии. Записи дублируются в stderr.
func Start(verbose bool) (*Session, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("не удалось создать каталог журналов %s: %w", dir, err)
	}

	start := time.Now()
	path := filepath.Join(dir, "gowalk-"+start.Format(fileNameLayout)+".log")

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл журнала %s: %w", path, err)
	}

	// Вывод C-библиотек (robotgo/gohook) уводим в журнал, а консольную копию
	// записей slog продолжаем писать в исходный stderr.
	console, err := redirectStderr(file)
	if err != nil {
		console = os.Stderr
	}

	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(io.MultiWriter(file, console), &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			// Полная дата уже есть в заголовке сессии и в имени файла,
			// в строках достаточно времени с миллисекундами.
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format("15:04:05.000"))
			}
			return a
		},
	})

	s := &Session{
		Logger:  slog.New(handler),
		Path:    path,
		file:    file,
		console: console,
		start:   start,
	}

	s.Logger.Info("=== запуск GOwalk ===",
		"время_запуска", start.Format("2006-01-02 15:04:05 MST"),
		"журнал", path,
		"pid", os.Getpid(),
		"os", runtime.GOOS,
		"arch", runtime.GOARCH,
		"go", runtime.Version(),
	)
	if exe, err := os.Executable(); err == nil {
		s.Logger.Info("исполняемый файл", "путь", exe)
	}
	if wd, err := os.Getwd(); err == nil {
		s.Logger.Debug("рабочий каталог", "путь", wd)
	}

	cleanup(dir, s.Logger)
	return s, nil
}

// Finish записывает итог запуска и закрывает файл журнала.
// reason == nil означает штатное завершение.
func (s *Session) Finish(reason error) {
	if s == nil || s.file == nil {
		return
	}

	cause := "штатное завершение"
	if reason != nil {
		cause = reason.Error()
	}

	s.Logger.Info("=== остановка GOwalk ===",
		"причина", cause,
		"время_остановки", time.Now().Format("2006-01-02 15:04:05 MST"),
		"время_работы", time.Since(s.start).Round(time.Millisecond).String(),
	)

	restoreStderr(s.console)
	s.console = nil

	_ = s.file.Close()
	s.file = nil
}

// cleanup оставляет в каталоге только keepFiles последних журналов.
func cleanup(dir string, log *slog.Logger) {
	files, err := filepath.Glob(filepath.Join(dir, "gowalk-*.log"))
	if err != nil || len(files) <= keepFiles {
		return
	}

	// Имена содержат метку времени в формате, сортируемом как строка.
	sort.Strings(files)
	for _, old := range files[:len(files)-keepFiles] {
		if err := os.Remove(old); err != nil {
			log.Debug("не удалось удалить старый журнал", "файл", old, "ошибка", err)
		}
	}
}
