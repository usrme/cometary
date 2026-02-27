package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

type colors struct {
	TitleTextStyle       styleConfig `json:"titleTextStyle"`
	TitleStyle           styleConfig `json:"titleStyle"`
	ItemStyle            styleConfig `json:"itemStyle"`
	CharacterCountColors colorConfig `json:"characterCountColors"`
	OverflowCharColor    colorConfig `json:"overflowCharColor"`
	SelectedItemColors   colorConfig `json:"selectedItemColors"`
	SelectedItemStyle    styleConfig `json:"selectedItemStyle"`
	SelectedItemPadded   styleConfig `json:"selectedItemPadded"`
	ItemDescriptionStyle styleConfig `json:"itemDescriptionStyle"`
	PaginationStyle      styleConfig `json:"paginationStyle"`
	HelpStyle            styleConfig `json:"helpStyle"`
	QuitTextStyle        styleConfig `json:"quitTextStyle"`
	VersionStyle         colorConfig `json:"versionStyle"`
}

type styleConfig struct {
	Margin        *int  `json:"margin,omitempty"`
	MarginLeft    *int  `json:"marginLeft,omitempty"`
	MarginRight   *int  `json:"marginRight,omitempty"`
	MarginTop     *int  `json:"marginTop,omitempty"`
	MarginBottom  *int  `json:"marginBottom,omitempty"`
	Padding       *int  `json:"padding,omitempty"`
	PaddingLeft   *int  `json:"paddingLeft,omitempty"`
	PaddingBottom *int  `json:"paddingBottom,omitempty"`
	Faint         *bool `json:"faint,omitempty"`
}

type colorConfig struct {
	Light string `json:"light,omitempty"`
	Dark  string `json:"dark,omitempty"`
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func defaultColors() colors {
	return colors{
		TitleTextStyle: styleConfig{
			MarginLeft: intPtr(2),
		},
		TitleStyle: styleConfig{
			MarginLeft: intPtr(2),
		},
		ItemStyle: styleConfig{
			PaddingLeft: intPtr(4),
		},
		CharacterCountColors: colorConfig{
			Light: "#8dacb6",
			Dark:  "240",
		},
		OverflowCharColor: colorConfig{
			Light: "#d08770",
			Dark:  "#d08770",
		},
		SelectedItemColors: colorConfig{
			Light: "#d08770",
			Dark:  "#a3be8c",
		},
		SelectedItemStyle: styleConfig{
			PaddingLeft: intPtr(2),
		},
		SelectedItemPadded: styleConfig{
			PaddingLeft: intPtr(2),
		},
		ItemDescriptionStyle: styleConfig{
			PaddingLeft: intPtr(2),
			Faint:       boolPtr(true),
		},
		PaginationStyle: styleConfig{
			PaddingLeft: intPtr(4),
		},
		HelpStyle: styleConfig{
			PaddingLeft:   intPtr(4),
			PaddingBottom: intPtr(1),
		},
		QuitTextStyle: styleConfig{
			Margin:       intPtr(1),
			MarginTop:    intPtr(0),
			MarginBottom: intPtr(2),
			MarginLeft:   intPtr(4),
		},
		VersionStyle: colorConfig{
			Light: "#9b9b9b",
			Dark:  "#5c5c5c",
		},
	}
}

func applyColors(c colors) {
	listStyles = list.DefaultStyles(true)
	titleTextStyle = lipgloss.NewStyle()
	if c.TitleStyle.MarginLeft != nil {
		titleStyle = lipgloss.NewStyle().MarginLeft(*c.TitleStyle.MarginLeft)
	} else {
		titleStyle = lipgloss.NewStyle()
	}
	if c.ItemStyle.PaddingLeft != nil {
		itemStyle = lipgloss.NewStyle().PaddingLeft(*c.ItemStyle.PaddingLeft)
	} else {
		itemStyle = lipgloss.NewStyle()
	}
	characterCountColors = compat.AdaptiveColor{
		Light: lipgloss.Color(c.CharacterCountColors.Light),
		Dark:  lipgloss.Color(c.CharacterCountColors.Dark),
	}
	overflowCharColor = compat.AdaptiveColor{
		Light: lipgloss.Color(c.OverflowCharColor.Light),
		Dark:  lipgloss.Color(c.OverflowCharColor.Dark),
	}
	selectedItemColors = compat.AdaptiveColor{
		Light: lipgloss.Color(c.SelectedItemColors.Light),
		Dark:  lipgloss.Color(c.SelectedItemColors.Dark),
	}
	selectedItemStyle = lipgloss.NewStyle().Foreground(selectedItemColors)
	if c.SelectedItemPadded.PaddingLeft != nil {
		selectedItemPadded = lipgloss.NewStyle().Foreground(selectedItemColors).PaddingLeft(*c.SelectedItemPadded.PaddingLeft)
	} else {
		selectedItemPadded = lipgloss.NewStyle().Foreground(selectedItemColors)
	}
	itemDescriptionStyle = lipgloss.NewStyle()
	if c.ItemDescriptionStyle.PaddingLeft != nil {
		itemDescriptionStyle = itemDescriptionStyle.PaddingLeft(*c.ItemDescriptionStyle.PaddingLeft)
	}
	if c.ItemDescriptionStyle.Faint != nil {
		itemDescriptionStyle = itemDescriptionStyle.Faint(*c.ItemDescriptionStyle.Faint)
	}
	paginationStyle = listStyles.PaginationStyle
	if c.PaginationStyle.PaddingLeft != nil {
		paginationStyle = paginationStyle.PaddingLeft(*c.PaginationStyle.PaddingLeft)
	}
	helpStyle = listStyles.HelpStyle
	if c.HelpStyle.PaddingLeft != nil {
		helpStyle = helpStyle.PaddingLeft(*c.HelpStyle.PaddingLeft)
	}
	if c.HelpStyle.PaddingBottom != nil {
		helpStyle = helpStyle.PaddingBottom(*c.HelpStyle.PaddingBottom)
	}
	quitTextStyle = lipgloss.NewStyle()
	if c.QuitTextStyle.Margin != nil || c.QuitTextStyle.MarginTop != nil || c.QuitTextStyle.MarginBottom != nil || c.QuitTextStyle.MarginLeft != nil {
		quitTextStyle = lipgloss.NewStyle().Margin(
			orDefault(c.QuitTextStyle.Margin, 0),
			orDefault(c.QuitTextStyle.MarginTop, 0),
			orDefault(c.QuitTextStyle.MarginBottom, 0),
			orDefault(c.QuitTextStyle.MarginLeft, 0),
		)
	}
	versionStyle = lipgloss.NewStyle().
		Foreground(compat.AdaptiveColor{
			Light: lipgloss.Color(c.VersionStyle.Light),
			Dark:  lipgloss.Color(c.VersionStyle.Dark),
		}).Render
}

func orDefault[T any](ptr *T, defaultVal T) T {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func loadColorsFromFile(path string) (colors, bool) {
	var c colors
	data, err := os.ReadFile(path)
	if err != nil {
		return c, false
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return c, false
	}
	return c, true
}

func loadColorScheme(configPath string, cfg *config) {
	if cfg.ColorScheme == "" {
		return
	}

	schemePath := ""

	dir := filepath.Dir(configPath)
	if dir != "." {
		schemePath = filepath.Join(dir, cfg.ColorScheme)
		if _, err := os.Stat(schemePath); err != nil {
			schemePath = ""
		}
	}

	if schemePath == "" {
		schemePath = filepath.Join(".", cfg.ColorScheme)
		if _, err := os.Stat(schemePath); err != nil {
			return
		}
	}

	loaded, ok := loadColorsFromFile(schemePath)
	if !ok {
		return
	}

	merged := mergeColors(defaultColors(), loaded)
	applyColors(merged)
}

func mergeColors(defaults, loaded colors) colors {
	if loaded.TitleStyle.MarginLeft != nil || loaded.TitleStyle.Margin != nil {
		defaults.TitleStyle = loaded.TitleStyle
	}
	if loaded.ItemStyle.PaddingLeft != nil || loaded.ItemStyle.Padding != nil {
		defaults.ItemStyle = loaded.ItemStyle
	}
	if loaded.CharacterCountColors.Light != "" || loaded.CharacterCountColors.Dark != "" {
		defaults.CharacterCountColors = loaded.CharacterCountColors
	}
	if loaded.OverflowCharColor.Light != "" || loaded.OverflowCharColor.Dark != "" {
		defaults.OverflowCharColor = loaded.OverflowCharColor
	}
	if loaded.SelectedItemColors.Light != "" || loaded.SelectedItemColors.Dark != "" {
		defaults.SelectedItemColors = loaded.SelectedItemColors
	}
	if loaded.SelectedItemStyle.PaddingLeft != nil || loaded.SelectedItemStyle.Padding != nil {
		defaults.SelectedItemStyle = loaded.SelectedItemStyle
	}
	if loaded.SelectedItemPadded.PaddingLeft != nil || loaded.SelectedItemPadded.Padding != nil {
		defaults.SelectedItemPadded = loaded.SelectedItemPadded
	}
	if loaded.ItemDescriptionStyle.PaddingLeft != nil || loaded.ItemDescriptionStyle.Padding != nil || loaded.ItemDescriptionStyle.Faint != nil {
		defaults.ItemDescriptionStyle = loaded.ItemDescriptionStyle
	}
	if loaded.PaginationStyle.PaddingLeft != nil || loaded.PaginationStyle.Padding != nil {
		defaults.PaginationStyle = loaded.PaginationStyle
	}
	if loaded.HelpStyle.PaddingLeft != nil || loaded.HelpStyle.Padding != nil || loaded.HelpStyle.PaddingBottom != nil {
		defaults.HelpStyle = loaded.HelpStyle
	}
	if loaded.QuitTextStyle.Margin != nil || loaded.QuitTextStyle.MarginLeft != nil || loaded.QuitTextStyle.MarginTop != nil || loaded.QuitTextStyle.MarginBottom != nil {
		defaults.QuitTextStyle = loaded.QuitTextStyle
	}
	if loaded.VersionStyle.Light != "" || loaded.VersionStyle.Dark != "" {
		defaults.VersionStyle = loaded.VersionStyle
	}
	return defaults
}
