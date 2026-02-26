package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func main() {
	config := loadConfig()

	format := config.ShowStatsFormat
	if config.SessionStatAsSeconds {
		format = "seconds"
	}

	tracker, err := NewRuntimeTracker("")
	if err != nil {
		fail("error creating tracker: %s", err)
	}

	if len(os.Args) > 1 && os.Args[1] == "-s" {
		stats := tracker.GetStats()
		fmt.Println(formatStat("Daily", stats.Daily[stats.CurrentDay], config.ShowStatsFormat))
		fmt.Println(formatStat("Weekly", stats.Weekly[stats.CurrentWeek], config.ShowStatsFormat))
		fmt.Println(formatStat("Monthly", stats.Monthly[stats.CurrentMonth], config.ShowStatsFormat))
		fmt.Println(formatStat("Yearly", stats.Yearly[stats.CurrentYear], config.ShowStatsFormat))
		os.Exit(0)
	}

	if err := findGitDir(); err != nil {
		fail(err.Error())
	}

	stagedFiles, err := filesInStaging()
	if err != nil {
		fail(err.Error())
	}

	commitSearchTerm := ""
	if len(os.Args) > 1 && os.Args[1] == "-m" {
		commitSearchTerm = os.Args[2]
	}

	if config.StoreRuntime || config.ShowRuntime {
		tracker.Start()
	}

	m := newModel(config, stagedFiles, commitSearchTerm)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fail(err.Error())
	}

	fmt.Println("")
	if !m.Finished() {
		fail("terminated")
	}

	msg, withBody := m.CommitMessage()
	if err := commit(msg, withBody, config.SignOffCommits); err != nil {
		fail("error committing: %s", err)
	}

	if config.StoreRuntime || config.ShowRuntime {
		err := tracker.Stop()
		if err != nil {
			fail("error stopping tracker: %s", err)
		}

		if config.ShowRuntime && !config.ShowStats {
			stats := tracker.GetStats()
			fmt.Println()
			fmt.Println(formatStat("Session", stats.Session, format))
		}
	}

	if config.ShowStats {
		stats := tracker.GetStats()
		fmt.Println()
		fmt.Println(formatStat("Session", stats.Session, format))
		fmt.Println(formatStat("Daily", stats.Daily[stats.CurrentDay], config.ShowStatsFormat))
		fmt.Println(formatStat("Weekly", stats.Weekly[stats.CurrentWeek], config.ShowStatsFormat))
		fmt.Println(formatStat("Monthly", stats.Monthly[stats.CurrentMonth], config.ShowStatsFormat))
		fmt.Println(formatStat("Yearly", stats.Yearly[stats.CurrentYear], config.ShowStatsFormat))
	}
}

func formatStat(stat string, seconds float32, format string) string {
	var value float32
	switch format {
	case "minutes":
		value = seconds / 60.0
	case "hours":
		value = seconds / 3600.0
	default:
		value = seconds
	}
	return fmt.Sprintf(" > %s: %.2f %s", stat, value, format)
}

func fail(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
