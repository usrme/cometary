package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// runDiffPager displays staged diff in an external pager.
// --no-pager is used to force git to output raw (allowing --color=always to work),
// then piped to less -R to enable ANSI color handling while providing navigation.
// The explicit less invocation (without -X) uses the alternate screen buffer so
// the pager's output is cleared on exit, leaving Cometary's UI in place rather
// than being pushed down by leftover diff content (as happens with git's default
// pager when configured with -X).
func runDiffPager() tea.Cmd {
	cmd := exec.Command("sh", "-c", "git --no-pager diff --color=always --cached | less -R")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return tea.ExecProcess(cmd, nil)
}

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

func defaultBranch() string {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	ref := strings.TrimSpace(string(output))
	if idx := strings.LastIndex(ref, "/"); idx >= 0 {
		return ref[idx+1:]
	}
	return ref
}

func currentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(string(output))
	}
	return strings.TrimSpace(string(output)), nil
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
