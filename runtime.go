package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Stats holds the runtime statistics for different time periods.
type Stats struct {
	Session      float32            `json:"session"`
	Daily        map[string]float32 `json:"daily"`
	CurrentDay   string             `json:"currentDay"`
	Weekly       map[string]float32 `json:"weekly"`
	CurrentWeek  string             `json:"currentWeek"`
	Monthly      map[string]float32 `json:"monthly"`
	CurrentMonth string             `json:"currentMonth"`
	Yearly       map[string]float32 `json:"yearly"`
	CurrentYear  string             `json:"currentYear"`
	LastUpdate   string             `json:"lastUpdate"`
}

// RuntimeTracker handles program runtime tracking.
type RuntimeTracker struct {
	statsFilePath string
	startTime     time.Time
	stats         Stats
}

// NewRuntimeTracker creates a new RuntimeTracker instance.
func NewRuntimeTracker(filename string) (*RuntimeTracker, error) {
	cfgDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	if filename == "" {
		filename = "stats.json"
	}
	path := filepath.Join(cfgDir, filename)

	rt := &RuntimeTracker{
		statsFilePath: path,
		stats: Stats{
			Session: 0.0,
			Daily:   make(map[string]float32),
			Weekly:  make(map[string]float32),
			Monthly: make(map[string]float32),
			Yearly:  make(map[string]float32),
		},
	}

	if _, err := os.Stat(rt.statsFilePath); errors.Is(err, os.ErrNotExist) {
		rt.saveStats()
	}

	if err := rt.loadStats(); err != nil {
		return nil, fmt.Errorf("failed to load stats: %w", err)
	}

	return rt, nil
}

// loadStats loads existing statistics.
func (rt *RuntimeTracker) loadStats() error {
	data, err := os.ReadFile(rt.statsFilePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	return json.Unmarshal(data, &rt.stats)
}

// saveStats saves current statistics.
func (rt *RuntimeTracker) saveStats() error {
	if err := os.MkdirAll(filepath.Dir(rt.statsFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(rt.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	return os.WriteFile(rt.statsFilePath, data, 0644)
}

// Start begins tracking runtime.
func (rt *RuntimeTracker) Start() {
	rt.startTime = time.Now()
}

// Stop ends tracking runtime and updates statistics.
func (rt *RuntimeTracker) Stop() error {
	if rt.startTime.IsZero() {
		return fmt.Errorf("tracking was not started")
	}

	endTime := time.Now()
	runtime := float32(endTime.Sub(rt.startTime).Seconds())

	// Format time periods
	date := endTime.Format("2006-01-02")

	// Get ISO week number
	y, week := endTime.ISOWeek()
	weekStr := fmt.Sprintf("%d-W%02d", y, week)

	month := endTime.Format("2006-01")
	year := endTime.Format("2006")

	// Update statistics
	rt.stats.Session = runtime
	rt.stats.Daily[date] += runtime
	rt.stats.CurrentDay = date
	rt.stats.Weekly[weekStr] += runtime
	rt.stats.CurrentWeek = weekStr
	rt.stats.Monthly[month] += runtime
	rt.stats.CurrentMonth = month
	rt.stats.Yearly[year] += runtime
	rt.stats.CurrentYear = year
	rt.stats.LastUpdate = endTime.Format(time.RFC3339)

	// Save updated stats
	if err := rt.saveStats(); err != nil {
		return fmt.Errorf("failed to save stats: %w", err)
	}

	// Reset start time
	rt.startTime = time.Time{}

	return nil
}

// GetStats returns the current statistics
func (rt *RuntimeTracker) GetStats() Stats {
	return rt.stats
}
