package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeTempState creates a temporary state file with the given content and
// returns its path.
func writeTempState(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, stateFile)
	if content != "" {
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	return path
}

func TestWasRunToday_NotYet(t *testing.T) {
	path := writeTempState(t, "")
	got, err := wasRunToday(path, "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false (empty file), got true")
	}
}

func TestWasRunToday_NoFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), stateFile)
	got, err := wasRunToday(path, "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false (missing file), got true")
	}
}

func TestWasRunToday_RunToday(t *testing.T) {
	content := "myapp=" + time.Now().Format("2006-01-02") + "\n"
	path := writeTempState(t, content)
	got, err := wasRunToday(path, "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected true (ran today), got false")
	}
}

func TestWasRunToday_DifferentApp(t *testing.T) {
	content := "otherapp=" + time.Now().Format("2006-01-02") + "\n"
	path := writeTempState(t, content)
	got, err := wasRunToday(path, "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false (different app), got true")
	}
}

func TestWasRunToday_OldDate(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	content := "myapp=" + yesterday + "\n"
	path := writeTempState(t, content)
	got, err := wasRunToday(path, "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false (yesterday's date), got true")
	}
}

func TestRecordRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), stateFile)
	if err := recordRun(path, "myapp"); err != nil {
		t.Fatalf("recordRun: %v", err)
	}
	got, err := wasRunToday(path, "myapp")
	if err != nil {
		t.Fatalf("wasRunToday: %v", err)
	}
	if !got {
		t.Error("expected wasRunToday=true after recordRun")
	}
}

func TestRecordRun_Idempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), stateFile)
	for i := 0; i < 3; i++ {
		if err := recordRun(path, "myapp"); err != nil {
			t.Fatalf("recordRun iteration %d: %v", i, err)
		}
	}
	// wasRunToday should still return true even with duplicate entries
	got, err := wasRunToday(path, "myapp")
	if err != nil {
		t.Fatalf("wasRunToday: %v", err)
	}
	if !got {
		t.Error("expected wasRunToday=true")
	}
}

func TestRun_NoArgs(t *testing.T) {
	err := run([]string{})
	if err == nil {
		t.Error("expected error for no arguments")
	}
}

func TestRun_AlreadyRanToday(t *testing.T) {
	dir := t.TempDir()
	content := "true=" + time.Now().Format("2006-01-02") + "\n"
	statePath := filepath.Join(dir, stateFile)
	if err := os.WriteFile(statePath, []byte(content), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Override home directory so run() finds our state file.
	t.Setenv("HOME", dir)

	// "true" was recorded as run today; run() should skip execution without
	// error.
	err := run([]string{"true"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_ExecutesProgram(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// "true" is a standard Unix binary that always succeeds.
	err := run([]string{"true"})
	if err != nil {
		t.Errorf("unexpected error running 'true': %v", err)
	}

	// After execution the state file should record today's run.
	statePath := filepath.Join(dir, stateFile)
	got, err := wasRunToday(statePath, "true")
	if err != nil {
		t.Fatalf("wasRunToday: %v", err)
	}
	if !got {
		t.Error("expected 'true' to be recorded after first run")
	}
}

func TestToday(t *testing.T) {
	d := today()
	if len(d) != 10 {
		t.Errorf("today() returned %q, expected YYYY-MM-DD (len 10)", d)
	}
}
