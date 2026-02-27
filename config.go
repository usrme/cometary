package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type prefix struct {
	T string `json:"title"`
	D string `json:"description"`
}

type config struct {
	Prefixes              []prefix `json:"prefixes"`
	SignOffCommits        bool     `json:"signOffCommits"`
	ScopeInputCharLimit   int      `json:"scopeInputCharLimit"`
	CommitInputCharLimit  int      `json:"commitInputCharLimit"`
	TotalInputCharLimit   int      `json:"totalInputCharLimit"`
	OverflowCharLimit     bool     `json:"overflowCharLimit"`
	ScopeCompletionOrder  string   `json:"scopeCompletionOrder"`
	FindAllCommitMessages bool     `json:"findAllCommitMessages"`
	StoreRuntime          bool     `json:"storeRuntime"`
	ShowRuntime           bool     `json:"showRuntime"`
	ShowStats             bool     `json:"showStats"`
	ShowStatsFormat       string   `json:"showStatsFormat"`
	SessionStatAsSeconds  bool     `json:"sessionStatAsSeconds"`
	ColorScheme           string   `json:"colorScheme,omitempty"`
}

func (i prefix) Title() string       { return i.T }
func (i prefix) Description() string { return i.D }
func (i prefix) FilterValue() string { return i.T }

var defaultPrefixes = []prefix{
	{
		T: "feat",
		D: "Introduces a new feature",
	},
	{
		T: "fix",
		D: "Patches a bug",
	},
	{
		T: "docs",
		D: "Documentation changes only",
	},
	{
		T: "test",
		D: "Adding missing tests or correcting existing tests",
	},
	{
		T: "build",
		D: "Changes that affect the build system",
	},
	{
		T: "ci",
		D: "Changes to CI configuration files and scripts",
	},
	{
		T: "perf",
		D: "A code change that improves performance",
	},
	{
		T: "refactor",
		D: "A code change that neither fixes a bug nor adds a feature",
	},
	{
		T: "revert",
		D: "Reverts a previous change",
	},
	{
		T: "style",
		D: "Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)",
	},
	{
		T: "chore",
		D: "A minor change which does not fit into any other category",
	},
}

const applicationName = "cometary"

func loadConfig() (*config, string) {
	nonXdgConfigFile := ".comet.json"

	// Check for configuration file local to current directory
	if _, err := os.Stat(nonXdgConfigFile); err == nil {
		return loadConfigFile(nonXdgConfigFile)
	}

	// Check for configuration file local to user's home directory
	if home, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(home, nonXdgConfigFile)
		if _, err := os.Stat(path); err == nil {
			return loadConfigFile(path)
		}
	}

	// Check for configuration file according to XDG Base Directory Specification
	if cfgDir, err := GetConfigDir(); err == nil {
		path := filepath.Join(cfgDir, "config.json")
		if _, err := os.Stat(path); err == nil {
			return loadConfigFile(path)
		}
	}

	return newConfig(), ""
}

func newConfig() *config {
	return &config{
		Prefixes:              defaultPrefixes,
		SignOffCommits:        false,
		ScopeInputCharLimit:   16,
		CommitInputCharLimit:  100,
		TotalInputCharLimit:   0,
		OverflowCharLimit:     false,
		ScopeCompletionOrder:  "descending",
		FindAllCommitMessages: false,
		StoreRuntime:          false,
		ShowRuntime:           false,
		ShowStats:             false,
		ShowStatsFormat:       "seconds",
		SessionStatAsSeconds:  true,
	}
}

func GetConfigDir() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")

	if configDir == "" || configDir[0:1] != "/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".config", applicationName), nil
	}

	return filepath.Join(configDir, applicationName), nil
}

func loadConfigFile(path string) (*config, string) {
	var c config
	data, err := os.ReadFile(path)
	if err != nil {
		return &c, path
	}

	if err := json.Unmarshal(data, &c); err != nil {
		return &c, path
	}

	return &c, path
}
