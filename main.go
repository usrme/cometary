package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func main() {
	config, configPath := loadConfig()
	applyColors(defaultColors())
	loadColorScheme(configPath, config)

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
		printStatsTable([]statRow{
			{"Daily", stats.Daily[stats.CurrentDay]},
			{"Weekly", stats.Weekly[stats.CurrentWeek]},
			{"Monthly", stats.Monthly[stats.CurrentMonth]},
			{"Yearly", stats.Yearly[stats.CurrentYear]},
		}, config.ShowStatsFormat)
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
			printStatsTable([]statRow{
				{"Session", stats.Session},
			}, format)
		}
	}

	if config.ShowStats {
		stats := tracker.GetStats()
		fmt.Println()
		printStatsTable([]statRow{
			{"Session", stats.Session},
			{"Daily", stats.Daily[stats.CurrentDay]},
			{"Weekly", stats.Weekly[stats.CurrentWeek]},
			{"Monthly", stats.Monthly[stats.CurrentMonth]},
			{"Yearly", stats.Yearly[stats.CurrentYear]},
		}, config.ShowStatsFormat)
	}
}

type statRow struct {
	Label   string
	Seconds float32
}

func printStatsTable(rows []statRow, format string) {
	// Compute display values.
	type displayRow struct {
		label string
		value string
		unit  string
	}
	unit := format
	if unit == "" {
		unit = "seconds"
	}

	display := make([]displayRow, len(rows))
	labelWidth := 0
	valueWidth := 0
	for i, r := range rows {
		var v float32
		switch format {
		case "minutes":
			v = r.Seconds / 60.0
		case "hours":
			v = r.Seconds / 3600.0
		default:
			v = r.Seconds
		}
		display[i] = displayRow{
			label: r.Label,
			value: fmt.Sprintf("%.2f", v),
			unit:  unit,
		}
		if len(r.Label) > labelWidth {
			labelWidth = len(r.Label)
		}
		if len(display[i].value) > valueWidth {
			valueWidth = len(display[i].value)
		}
	}

	// Derive hbar width from a concrete formatted row so padding is exact.
	unitWidth := len(unit)
	sampleRow := fmt.Sprintf("  %-*s  %*s  %-*s  ",
		labelWidth, "", valueWidth, "", unitWidth, unit)
	hbar := strings.Repeat("─", len(sampleRow))

	fmt.Printf(" ╭%s╮\n", hbar)
	for _, d := range display {
		fmt.Printf(" │  %-*s  %*s  %-*s  │\n",
			labelWidth, d.label,
			valueWidth, d.value,
			unitWidth, d.unit,
		)

	}
	fmt.Printf(" ╰%s╯\n", hbar)
}

func fail(format string, args ...any) {
	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
