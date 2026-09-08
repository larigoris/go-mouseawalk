package logging

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// readLatestLog возвращает содержимое единственного/последнего журнала в dir.
func readLatestLog(t *testing.T, dir string) string {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(dir, "gowalk-*.log"))
	if err != nil || len(files) == 0 {
		t.Fatalf("в %s не создано ни одного журнала (err=%v)", dir, err)
	}

	data, err := os.ReadFile(files[len(files)-1])
	if err != nil {
		t.Fatalf("не удалось прочитать журнал: %v", err)
	}
	return string(data)
}

func TestDirUsesEnvOverride(t *testing.T) {
	want := filepath.Join(t.TempDir(), "мой-каталог")
	t.Setenv("GOWALK_LOG_DIR", want)

	got, err := Dir()
	if err != nil {
		t.Fatalf("Dir() вернул ошибку: %v", err)
	}
	if got != want {
		t.Errorf("Dir() = %q, ожидалось %q", got, want)
	}
}

func TestDirDefaultIsPlatformSpecific(t *testing.T) {
	t.Setenv("GOWALK_LOG_DIR", "")

	got, err := Dir()
	if err != nil {
		t.Fatalf("Dir() вернул ошибку: %v", err)
	}

	var want string
	switch runtime.GOOS {
	case "darwin":
		want = filepath.Join("Library", "Logs", appName)
	case "windows":
		want = filepath.Join(appName, "Logs")
	default:
		want = filepath.Join(appName, "logs")
	}

	if !strings.Contains(got, want) {
		t.Errorf("Dir() = %q, ожидался путь, содержащий %q", got, want)
	}
}

func TestStartCreatesLogWithSessionHeader(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOWALK_LOG_DIR", dir)

	session, err := Start(false)
	if err != nil {
		t.Fatalf("Start() вернул ошибку: %v", err)
	}
	session.Finish(nil)

	body := readLatestLog(t, dir)
	for _, want := range []string{"запуск GOwalk", "время_запуска", "pid", runtime.GOOS, runtime.GOARCH} {
		if !strings.Contains(body, want) {
			t.Errorf("в заголовке журнала нет %q\nсодержимое:\n%s", want, body)
		}
	}
}

func TestStartCreatesMissingDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "нет", "такого", "пути")
	t.Setenv("GOWALK_LOG_DIR", dir)

	session, err := Start(false)
	if err != nil {
		t.Fatalf("Start() не создал каталог: %v", err)
	}
	session.Finish(nil)

	if _, err := os.Stat(dir); err != nil {
		t.Errorf("каталог %s не создан: %v", dir, err)
	}
}

func TestFinishRecordsReason(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOWALK_LOG_DIR", dir)

	session, err := Start(false)
	if err != nil {
		t.Fatalf("Start() вернул ошибку: %v", err)
	}
	session.Finish(errors.New("тестовая причина остановки"))

	body := readLatestLog(t, dir)
	for _, want := range []string{"остановка GOwalk", "тестовая причина остановки", "время_работы"} {
		if !strings.Contains(body, want) {
			t.Errorf("в записи об остановке нет %q\nсодержимое:\n%s", want, body)
		}
	}
}

func TestFinishWithoutReasonIsCleanShutdown(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOWALK_LOG_DIR", dir)

	session, err := Start(false)
	if err != nil {
		t.Fatalf("Start() вернул ошибку: %v", err)
	}
	session.Finish(nil)

	if body := readLatestLog(t, dir); !strings.Contains(body, "штатное завершение") {
		t.Errorf("ожидалось 'штатное завершение'\nсодержимое:\n%s", body)
	}
}

func TestFinishIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOWALK_LOG_DIR", dir)

	session, err := Start(false)
	if err != nil {
		t.Fatalf("Start() вернул ошибку: %v", err)
	}

	session.Finish(nil)
	session.Finish(nil) // повторный вызов не должен паниковать
}

func TestCleanupKeepsOnlyRecentLogs(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOWALK_LOG_DIR", dir)

	// Заведомо старые журналы: их должно стать меньше после запуска.
	old := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < keepFiles+5; i++ {
		name := "gowalk-" + old.Add(time.Duration(i)*time.Minute).Format(fileNameLayout) + ".log"
		if err := os.WriteFile(filepath.Join(dir, name), []byte("старый\n"), 0o644); err != nil {
			t.Fatalf("не удалось создать тестовый журнал: %v", err)
		}
	}

	session, err := Start(false)
	if err != nil {
		t.Fatalf("Start() вернул ошибку: %v", err)
	}
	session.Finish(nil)

	files, err := filepath.Glob(filepath.Join(dir, "gowalk-*.log"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) != keepFiles {
		t.Errorf("осталось %d журналов, ожидалось %d", len(files), keepFiles)
	}

	// Журнал текущего запуска обязан уцелеть.
	if _, err := os.Stat(session.Path); err != nil {
		t.Errorf("журнал текущего запуска удалён: %v", err)
	}
}
