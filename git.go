package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func filesInStaging() ([]string, error) {
	cmd := exec.Command("git", "diff", "--no-ext-diff", "--cached", "--name-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}, errors.New(string(output))
	}
	lines := strings.TrimSpace(string(output))
	if lines == "" {
		return []string{}, fmt.Errorf("no files added to staging area")
	}
	return strings.Split(lines, "\n"), nil
}

func findGitDir() error {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}
	return nil
}

func commit(msg string, body bool, signOff bool) error {
	gitArgs := os.Args[1:]
	if len(os.Args) > 1 && os.Args[1] == "-m" {
		gitArgs = os.Args[3:]
	}
	args := append([]string{
		"commit", "-m", msg,
	}, gitArgs...)
	if body {
		args = append(args, "-e")
	}
	if signOff {
		args = append(args, "-s")
	}
	cmd := exec.Command("git", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
