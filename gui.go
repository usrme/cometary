package main

import (
	"fmt"
	"io"
	"os/exec"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"golang.org/x/exp/maps"
)

const (
	defaultWidth = 40
	listHeight   = 15
)

var (
	titleTextStyle        lipgloss.Style
	titleStyle            lipgloss.Style
	itemStyle             lipgloss.Style
	characterCountColors  compat.AdaptiveColor
	overflowCharColor     compat.AdaptiveColor
	selectedItemColors    compat.AdaptiveColor
	selectedItemStyle     lipgloss.Style
	selectedItemPadded    lipgloss.Style
	itemDescriptionStyle  lipgloss.Style
	listStyles            list.Styles
	paginationStyle       lipgloss.Style
	helpStyle             lipgloss.Style
	quitTextStyle         lipgloss.Style
	versionStyle          func(...string) string
	defaultPromptStyles   textinput.Styles
	overflowPromptStyles  textinput.Styles
	defaultLimitStyle     lipgloss.Style
	overflowLimitStyle    lipgloss.Style
	selectedItemIndicator string
	scopeInputText        = "What is the scope?"
	msgInputText          = "What is the commit message?"
	bodyInputText         = "Do you need to specify a body/footer?"
	constrainInput        bool
	totalInputCharLimit   int
)

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(prefix)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Title())

	var output string
	if index == m.Index() {
		output = selectedItemPadded.Render(selectedItemIndicator + str)
	} else {
		output = itemStyle.Render(str)
	}
	output += itemDescriptionStyle.PaddingLeft(15 - len(str)).Render(i.Description())

	_, _ = fmt.Fprint(w, output)
}

type (
	stagedFilesMsg    []string
	commitMessagesMsg []string
)

type model struct {
	chosenPrefix           bool
	chosenScope            bool
	chosenMsg              bool
	chosenBody             bool
	specifyBody            bool
	prefix                 string
	prefixDescription      string
	scope                  string
	msg                    string
	prefixList             list.Model
	msgInput               textinput.Model
	scopeInput             textinput.Model
	ynInput                textinput.Model
	constrainInput         bool
	totalInputCharLimit    int
	overflowCharLimit      bool
	previousInputTexts     string
	typed                  int
	quitting               bool
	stagedFiles            []string
	scopeCompletionOrder   string
	stagedFilePathSegments []string
	scopeInputIndex        int
	commitSearchTerm       string
	findAllCommitMessages  bool
	commitMessages         []string
	messageInputIndex      int
	branchName             string
	scopeBranchFirst       bool
}

func newModel(c *config, stagedFiles []string, commitSearchTerm string, branchName string) *model {
	prefixes := convertPrefixes(c.Prefixes)
	prefixList := list.New(prefixes, itemDelegate{}, defaultWidth, listHeight)
	prefixList.Title = "What are you committing?"
	prefixList.SetShowStatusBar(false)
	prefixList.SetFilteringEnabled(false)
	prefixList.Styles.Title = titleTextStyle
	prefixList.Styles.PaginationStyle = paginationStyle
	prefixList.Styles.HelpStyle = helpStyle
	prefixList.NewStatusMessage(versionStyle(pkgVersion()))

	scopeInput := textinput.New()
	scopeInput.Placeholder = "Scope"
	scopeInput.Prompt = selectedItemIndicator

	commitInput := textinput.New()
	commitInput.Placeholder = "Commit message"
	commitInput.Prompt = selectedItemIndicator

	bodyConfirmation := textinput.New()
	bodyConfirmation.Placeholder = "y/N"
	bodyConfirmation.CharLimit = 1
	bodyConfirmation.SetWidth(20)
	bodyConfirmation.Prompt = selectedItemIndicator

	scopeInput.SetStyles(defaultPromptStyles)
	commitInput.SetStyles(defaultPromptStyles)
	bodyConfirmation.SetStyles(defaultPromptStyles)

	if c == nil || c.ScopeInputCharLimit == 0 {
		scopeInput.CharLimit = 16
		scopeInput.SetWidth(20)
	} else {
		scopeInput.CharLimit = c.ScopeInputCharLimit
		scopeInput.SetWidth(c.ScopeInputCharLimit)
	}

	if c == nil || c.CommitInputCharLimit == 0 {
		commitInput.CharLimit = 100
		commitInput.SetWidth(50)
	} else {
		commitInput.CharLimit = c.CommitInputCharLimit
		commitInput.SetWidth(c.CommitInputCharLimit)
	}

	if c != nil && c.OverflowCharLimit {
		scopeInput.CharLimit = 9999
		commitInput.CharLimit = 9999
	}

	if c == nil || c.TotalInputCharLimit == 0 {
		constrainInput = false
	} else {
		constrainInput = true
		totalInputCharLimit = c.TotalInputCharLimit
	}

	bindings := []key.Binding{
		customKeys.Cycle,
		customKeys.Diff,
	}
	prefixList.AdditionalShortHelpKeys = func() []key.Binding { return bindings }
	prefixList.AdditionalFullHelpKeys = func() []key.Binding { return bindings }

	return &model{
		prefixList:            prefixList,
		scopeInput:            scopeInput,
		msgInput:              commitInput,
		ynInput:               bodyConfirmation,
		constrainInput:        constrainInput,
		totalInputCharLimit:   totalInputCharLimit,
		overflowCharLimit:     c.OverflowCharLimit,
		stagedFiles:           stagedFiles,
		scopeCompletionOrder:  c.ScopeCompletionOrder,
		commitSearchTerm:      commitSearchTerm,
		findAllCommitMessages: c.FindAllCommitMessages,
		branchName:            branchName,
		scopeBranchFirst:      c.ScopeBranchFirst,
	}
}

func convertPrefixes(prefixes []prefix) []list.Item {
	var output []list.Item
	for _, prefix := range prefixes {
		output = append(output, prefix)
	}
	return output
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		formUniquePaths(m.stagedFiles, m.scopeCompletionOrder, m.branchName, m.scopeBranchFirst),
		findCommitMessages(m.commitSearchTerm, m.findAllCommitMessages, m.branchName, m.scopeBranchFirst),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.PasteMsg:
		switch {
		case !m.chosenPrefix:
			return m, nil
		case !m.chosenScope:
			m.scopeInput.SetValue(m.scopeInput.Value() + msg.Content)
			m.scopeInput.CursorEnd()
		case !m.chosenMsg:
			m.msgInput.SetValue(m.msgInput.Value() + msg.Content)
			m.msgInput.CursorEnd()
		}
		return m, nil
	case tea.KeyPressMsg:
		if key.Matches(msg, customKeys.Diff) {
			return m, runDiffPager()
		}
		switch {
		case msg.String() == "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case !m.chosenPrefix:
			return m.updatePrefixList(msg)
		case !m.chosenScope:
			return m.updateScopeInput(msg)
		case !m.chosenMsg:
			return m.updateMsgInput(msg)
		case !m.chosenBody:
			return m.updateYNInput(msg)
		default:
			return m, tea.Quit
		}
	case stagedFilesMsg:
		m.stagedFilePathSegments = msg
		return m, nil
	case commitMessagesMsg:
		m.commitMessages = msg
		return m, nil
	}
	return m, nil
}

func (m *model) Finished() bool {
	return m.chosenBody
}

func (m *model) CommitMessage() (string, bool) {
	prefix := m.prefix
	if m.scope != "" {
		prefix = fmt.Sprintf("%s(%s)", prefix, m.scope)
	}
	return fmt.Sprintf("%s: %s", prefix, m.msg), m.specifyBody
}

func (m *model) continueWithSelectedItem() {
	i, ok := m.prefixList.SelectedItem().(prefix)
	if ok {
		m.prefix = i.Title()
		m.prefixDescription = i.Description()
		m.chosenPrefix = true
		m.previousInputTexts = fmt.Sprintf(
			"\n%s %s\n",
			m.prefixList.Title,
			selectedItemStyle.Render(fmt.Sprintf("%s: %s", m.prefix, m.prefixDescription)),
		)
		m.typed = len(m.prefix) + len("(): ")
		m.scopeInput.Focus()
	}
}

func (m *model) updatePrefixList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.prefixList.SetWidth(msg.Width)
		return m, nil

	case tea.KeyPressMsg:
		switch keypress := msg.String(); keypress {

		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			var index int
			if keypress == "0" && len(m.prefixList.Items()) == 10 {
				// zero-based indexing, so index 9 equals element 10
				index = 9
			} else if keypress == "0" && len(m.prefixList.Items()) < 10 {
				// keep selected item where it was at
				return m, nil
			} else {
				index, _ = strconv.Atoi(keypress)
				index = index - 1
			}
			m.prefixList.Select(index)
			m.continueWithSelectedItem()

		case "enter":
			m.continueWithSelectedItem()
		}
	}

	var cmd tea.Cmd
	m.prefixList, cmd = m.prefixList.Update(msg)
	return m, cmd
}

func (m *model) updateScopeInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyEnter:
			m.chosenScope = true
			m.scope = m.scopeInput.Value()
			m.typed += len(m.scope)
			m.previousInputTexts = fmt.Sprintf(
				"%s%s %s\n",
				m.previousInputTexts,
				scopeInputText,
				selectedItemStyle.Render(m.scope),
			)
			m.msgInput.Focus()
		case tea.KeyTab:
			m.scopeInput.SetValue(m.stagedFilePathSegments[m.scopeInputIndex])
			if m.scopeInputIndex+1 == len(m.stagedFilePathSegments) {
				m.scopeInputIndex = 0
				return m, nil
			}
			m.scopeInputIndex += 1
			m.scopeInput.CursorEnd()
		case tea.KeyEsc:
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.scopeInput, cmd = m.scopeInput.Update(msg)
	return m, cmd
}

func (m *model) updateMsgInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyEnter:
			m.chosenMsg = true
			m.msg = m.msgInput.Value()
			m.typed += len(m.msg)
			m.previousInputTexts = fmt.Sprintf(
				"%s%s %s\n",
				m.previousInputTexts,
				msgInputText,
				selectedItemStyle.Render(m.msg),
			)
			m.ynInput.Focus()
		case tea.KeyTab:
			if len(m.commitMessages) > 0 {
				m.msgInput.SetValue(m.commitMessages[m.messageInputIndex])
				if m.messageInputIndex+1 == len(m.commitMessages) {
					m.messageInputIndex = 0
					return m, nil
				}
				m.messageInputIndex += 1
				m.msgInput.CursorEnd()
			}
		case tea.KeyEsc:
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.msgInput, cmd = m.msgInput.Update(msg)
	return m, cmd
}

func (m *model) updateYNInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyEnter:
			m.chosenMsg = true
			switch strings.ToLower(m.ynInput.Value()) {
			case "y":
				m.specifyBody = true
			}
			m.chosenBody = true
			m.previousInputTexts = fmt.Sprintf(
				"%s%s %s\n",
				m.previousInputTexts,
				bodyInputText,
				selectedItemStyle.Render(strconv.FormatBool(m.specifyBody)),
			)
			return m, tea.Quit
		case tea.KeyEsc:
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.ynInput, cmd = m.ynInput.Update(msg)
	return m, cmd
}

func renderCurrentLimit(m *model, charLimit int, input string) string {
	overflow, _ := getInputColors(m, charLimit, input)
	style := defaultLimitStyle
	if overflow {
		style = overflowLimitStyle
	}
	return style.Render(fmt.Sprintf("[%s/%d]", getInputCount(m, charLimit, input), getInputLimit(m, charLimit)))
}

func getInputColors(m *model, charLimit int, input string) (bool, compat.AdaptiveColor) {
	limit := getInputLimit(m, charLimit)
	inputLength := getInputLength(m, input)

	color := characterCountColors
	overflow := m.overflowCharLimit && inputLength > limit
	if overflow {
		color = overflowCharColor
	}
	return overflow, color
}

func getInputLimit(m *model, charLimit int) int {
	if m.constrainInput {
		return m.totalInputCharLimit
	}
	return charLimit
}

func getInputLength(m *model, input string) int {
	if m.constrainInput {
		return len(m.prefix) + len("(): ") + len(input) + len(m.scope)
	}
	return len(input)
}

func getInputCount(m *model, charLimit int, input string) string {
	limit := getInputLimit(m, charLimit)
	inputLength := getInputLength(m, input)
	padWidth := len(strconv.Itoa(limit))
	return fmt.Sprintf("%0*d", padWidth, inputLength)
}

func (m *model) View() tea.View {
	lengthExceedMessage := "Number of characters equals total input limit. Value will be left blank"

	switch {
	case !m.chosenPrefix:
		return tea.NewView("\n" + m.prefixList.View())
	case !m.chosenScope:
		limit := renderCurrentLimit(m, m.scopeInput.CharLimit, m.scopeInput.Value())

		if m.constrainInput && !m.overflowCharLimit {
			m.scopeInput.CharLimit = m.totalInputCharLimit - m.typed
			if m.scopeInput.CharLimit == 0 {
				m.scopeInput.Placeholder = lengthExceedMessage
				m.scopeInput.EchoMode = textinput.EchoNone
				m.scopeInput.SetValue("")
			}
		}

		overflow, _ := getInputColors(m, m.scopeInput.CharLimit, m.scopeInput.Value())
		if overflow {
			m.scopeInput.SetStyles(overflowPromptStyles)
		} else {
			m.scopeInput.SetStyles(defaultPromptStyles)
		}

		return tea.NewView(titleStyle.Render(fmt.Sprintf(
			"%s%s (Enter to skip / Esc to cancel) %s\n%s",
			m.previousInputTexts,
			scopeInputText,
			limit,
			m.scopeInput.View(),
		)))
	case !m.chosenMsg:
		limit := renderCurrentLimit(m, m.msgInput.CharLimit, m.msgInput.Value())

		if m.constrainInput && !m.overflowCharLimit {
			m.msgInput.CharLimit = m.totalInputCharLimit - m.typed
			if m.msgInput.CharLimit == 0 {
				m.msgInput.Placeholder = lengthExceedMessage
				m.msgInput.EchoMode = textinput.EchoNone
				m.msgInput.SetValue("")
			}
		}

		overflow, _ := getInputColors(m, m.msgInput.CharLimit, m.msgInput.Value())
		if overflow {
			m.msgInput.SetStyles(overflowPromptStyles)
		} else {
			m.msgInput.SetStyles(defaultPromptStyles)
		}

		return tea.NewView(titleStyle.Render(fmt.Sprintf(
			"%s%s (Esc to cancel) %s\n%s",
			m.previousInputTexts,
			msgInputText,
			limit,
			m.msgInput.View(),
		)))
	case !m.chosenBody:
		return tea.NewView(titleStyle.Render(fmt.Sprintf(
			"%s%s (Esc to cancel)\n%s",
			m.previousInputTexts,
			bodyInputText,
			m.ynInput.View(),
		)))
	case m.quitting:
		return tea.NewView(quitTextStyle.Render("Aborted.\n"))
	default:
		return tea.NewView(titleStyle.Render(fmt.Sprintf(
			"%s\n---\n",
			m.previousInputTexts,
		)))
	}
}

func branchCompletionEntries(branchName string) []string {
	if branchName == "" || branchName == defaultBranch() {
		return nil
	}
	return []string{strings.ToLower(branchName), strings.ToUpper(branchName)}
}

func formUniquePaths(stagedFiles []string, scopeCompletionOrder string, branchName string, scopeBranchFirst bool) tea.Cmd {
	return func() tea.Msg {
		uniqueMap := make(map[string]bool)
		var joinedPaths []string
		for _, p := range stagedFiles {
			if uniqueMap[p] {
				continue
			}
			s := strings.Split(p, "/")
			for j, q := range s {
				if j == len(s)-1 {
					uniqueMap[q] = true
				}
				joinedPaths = append(joinedPaths, q)
				uniqueMap[strings.Join(joinedPaths, "/")] = true
			}
			joinedPaths = joinedPaths[:0]
		}

		uniquePaths := maps.Keys(uniqueMap)
		ascending := scopeCompletionOrder == "ascending"
		sort.Slice(uniquePaths, func(i, j int) bool {
			if ascending {
				return len(uniquePaths[i]) < len(uniquePaths[j])
			}
			return len(uniquePaths[i]) > len(uniquePaths[j])
		})
		if scopeBranchFirst {
			entries := branchCompletionEntries(branchName)
			if entries != nil {
				uniquePaths = append(entries, uniquePaths...)
			}
		}
		return stagedFilesMsg(uniquePaths)
	}
}

func findCommitMessages(grep string, findAll bool, branchName string, scopeBranchFirst bool) tea.Cmd {
	return func() tea.Msg {
		if grep == "" {
			if scopeBranchFirst {
				return commitMessagesMsg(branchCompletionEntries(branchName))
			}
			return commitMessagesMsg([]string{})
		}
		cmd := exec.Command("git", "log", "--oneline", "--pretty=format:%s", "--grep="+grep)
		output, err := cmd.CombinedOutput()
		if err != nil {
			if scopeBranchFirst {
				return commitMessagesMsg(branchCompletionEntries(branchName))
			}
			return commitMessagesMsg([]string{})
		}

		messages := strings.Split(strings.TrimSpace(string(output)), "\n")
		uniqueMap := make(map[string]bool)
		var msg string
		for _, m := range messages {
			msg = m
			if !findAll {
				// Given conventional commit adherence, the semicolon can be assumed
				// to be a safe enough delimiter upon which to separate prefix, an
				// optional scope, and the message
				s := strings.Split(m, ":")
				// If m does not contain colon then it's not a valid conventional commit
				if len(s) == 1 {
					continue
				}
				msg = strings.TrimSpace(s[1])
			}
			if _, ok := uniqueMap[msg]; ok {
				continue
			}
			uniqueMap[msg] = true
		}

		var result []string
		if scopeBranchFirst {
			entries := branchCompletionEntries(branchName)
			if entries != nil {
				result = append(result, entries...)
			}
		}
		for m := range uniqueMap {
			result = append(result, m)
		}
		return commitMessagesMsg(result)
	}
}

var pkgVersion = sync.OnceValue(func() string {
	if version != "" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	var vcsRev, vcsModified string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			vcsRev = s.Value
		case "vcs.modified":
			vcsModified = s.Value
		}
	}
	if vcsRev != "" {
		if vcsModified == "true" {
			return vcsRev + " (modified)"
		}
		return vcsRev
	}
	return "unknown"
})
