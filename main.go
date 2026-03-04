package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func stateDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	dir := filepath.Join(home, ".onceaday")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create state directory: %w", err)
	}
	return dir, nil
}

func stateFile(dir, key string) string {
	h := sha256.Sum256([]byte(key))
	return filepath.Join(dir, fmt.Sprintf("%x", h))
}

func hasRunToday(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	today := time.Now().Format("2006-01-02")
	return strings.TrimSpace(string(data)) == today
}

func recordRun(path string) error {
	today := time.Now().Format("2006-01-02")
	return os.WriteFile(path, []byte(today), 0600)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: onceaday <command> [args...]")
		os.Exit(1)
	}

	dir, err := stateDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	key := strings.Join(args, " ")
	sf := stateFile(dir, key)

	if hasRunToday(sf) {
		fmt.Fprintf(os.Stderr, "already ran today: %s\n", key)
		os.Exit(0)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := recordRun(sf); err != nil {
		fmt.Fprintln(os.Stderr, "warning: could not record run:", err)
	}
}
