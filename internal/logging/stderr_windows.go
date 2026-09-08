//go:build windows

package logging

import "os"

// redirectStderr на Windows не используется: тамошняя сборка не тянет
// C-библиотеки, пишущие в stderr напрямую.
func redirectStderr(_ *os.File) (*os.File, error) {
	return os.Stderr, nil
}

// restoreStderr на Windows делать нечего.
func restoreStderr(_ *os.File) {}
