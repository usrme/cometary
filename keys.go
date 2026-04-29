package main

import "charm.land/bubbles/v2/key"

type customKeyMap struct {
	Cycle key.Binding
	Diff  key.Binding
}

var customKeys = customKeyMap{
	Cycle: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "cycle through commit messages or changed file paths"),
	),
	Diff: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "view staged changes"),
	),
}
