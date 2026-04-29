package main

import "charm.land/bubbles/v2/key"

type customKeyMap struct {
	Cycle  key.Binding
	DiffCh key.Binding
}

var customKeys = customKeyMap{
	Cycle: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "cycle through commit messages or changed file paths"),
	),
	DiffCh: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("Ctrl+P", "view staged changes"),
	),
}
