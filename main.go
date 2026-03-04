package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const stateFile = ".onceaday"

// today returns the current date as YYYY-MM-DD.
func today() string {
	return time.Now().Format("2006-01-02")
}

// stateFilePath returns the path to the state file in the user's home directory.
func stateFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, stateFile), nil
}

// wasRunToday reports whether program was already recorded as run today in the
// given state file.
func wasRunToday(path, program string) (bool, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer f.Close()

	target := program + "=" + today()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if scanner.Text() == target {
			return true, nil
		}
	}
	return false, scanner.Err()
}

// recordRun appends an entry for program with today's date to the state file.
func recordRun(path, program string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s=%s\n", program, today())
	return err
}

// runProgram executes program with the given args, wiring its stdio to the
// current process.
func runProgram(program string, args []string) error {
	cmd := exec.Command(program, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: onceaday <program> [args...]")
	}

	program := args[0]
	programArgs := args[1:]

	path, err := stateFilePath()
	if err != nil {
		return err
	}

	done, err := wasRunToday(path, program)
	if err != nil {
		return fmt.Errorf("reading state file: %w", err)
	}
	if done {
		return nil
	}

	runErr := runProgram(program, programArgs)

	if err := recordRun(path, program); err != nil {
		return fmt.Errorf("recording run: %w", err)
	}

	if runErr != nil {
		return fmt.Errorf("running %s: %w", program, runErr)
	}
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "onceaday:", err)
		os.Exit(1)
	}
}
