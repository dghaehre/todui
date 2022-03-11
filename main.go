package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/////////////////
// Model
/////////////////

// Keys
type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Help  key.Binding
	Quit  key.Binding
}

type InputMode = string
type Mode = string
type View = string
type Page = uint8

const (
	allTasksPage Page = iota
	projectsPage
)

var (
	insertInputMode InputMode = "insert"
	normalInputMode InputMode = "normal"
	defaultMode     Mode      = "default"

	defaultView View = "default"

	keys = keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		)}

	defaultTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("12"))

	dimTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	chosenTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))
)

type Todo struct {
	desc string
}

type cursorPosition struct {
	view  View
	index int
}

type model struct {
	keys        keyMap
	inputMode   InputMode
	mode        Mode
	totalWidth  int
	totalHeight int
	debug       bool
	todos       []Todo
	cursor      cursorPosition
	page        Page
}

var mockTodos []Todo = []Todo{
	{desc: "a test"},
	{desc: "another test"},
}

func NewModel(debug bool) model {
	return model{
		keys:      keys,
		mode:      defaultMode,
		inputMode: normalInputMode,
		debug:     debug,
		page:      allTasksPage,
		cursor: cursorPosition{
			view:  defaultView,
			index: 0,
		},
		todos: mockTodos,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

/////////////
// Update
////////////

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Set window size
	case tea.WindowSizeMsg:
		m.totalHeight = msg.Height
		m.totalWidth = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Down), key.Matches(msg, m.keys.Up):
			m.moveCursor(msg)
		case key.Matches(msg, m.keys.Left), key.Matches(msg, m.keys.Right):
			m.changePage(msg)
		case key.Matches(msg, m.keys.Help):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

/////////////
// View
////////////

func (m model) View() string {
	top := m.topBar()
	content := m.defaultList()
	return m.debugView() + top + content + m.getEmptyLines(content+top) + m.bottomBar()
}

func (m model) debugView() string {
	if !m.debug {
		return ""
	}
	content := fmt.Sprintf("STATE: Height: %d, Width: %d, InputMode: %+v, Mode: %+v, Cursor: %+v, Page: %+v",
		m.totalHeight, m.totalWidth, m.inputMode, m.mode, m.cursor, m.page)
	style := lipgloss.NewStyle().
		Width(m.totalWidth).
		Align(lipgloss.Right)
	return style.Render(content) + "\n"
}

func (m model) topBar() string {
	style := lipgloss.NewStyle().
		Width(m.totalWidth).
		// Foreground(lipgloss.Color("13")).
		Align(lipgloss.Left)

	var s string

	// all tasks
	if m.page == allTasksPage {
		s += chosenTextStyle.Render("by task")
	} else {
		s += dimTextStyle.Render("by task")
	}

	s += " | "

	// projects
	if m.page == projectsPage {
		s += chosenTextStyle.Render("projects")
	} else {
		s += dimTextStyle.Render("projects")
	}
	return style.Render(s) + "\n" + "\n"
}

func (m model) bottomBar() string {
	style := lipgloss.NewStyle().
		Width(m.totalWidth).
		Foreground(lipgloss.Color("12")).
		Align(lipgloss.Right)

	var s string
	ks := m.getValidKeys(m.keys)
	for i, v := range ks {
		h := v.Help()
		s += h.Key + ": " + h.Desc
		if (i + 1) != len(ks) {
			s += "   "
		}
	}
	return style.Render(s)
}

func (m model) defaultList() string {
	// We currently assume that the cursor position has view default
	content := ""
	style := lipgloss.NewStyle().
		Width(m.totalWidth).
		Align(lipgloss.Left)

	for i, v := range m.todos {
		if m.cursor.index == i {
			content += "→ " + defaultTextStyle.Render(v.desc)
		} else {
			content += "  " + defaultTextStyle.Render(v.desc)
		}
		content += "\n"
	}
	return style.Render(content)
}

/////////////
// Utils
////////////

// Returns the valid command keys to be used, based on current state
func (m model) getValidKeys(k keyMap) []key.Binding {
	switch m.mode {
	case defaultMode:
		return []key.Binding{k.Up, k.Down, k.Help, k.Quit}
	}
	return []key.Binding{k.Quit} // Cannot and should not happen
}

func (m model) getEmptyLines(content string) string {
	s := ""
	statusBar := 1
	debug := 0
	if m.debug {
		debug = 1
	}
	lines := m.totalHeight - strings.Count(content, "\n") - statusBar - debug
	for i := 0; i < lines; i++ {
		s += "\n"
	}
	return s
}

func (m *model) changePage(km tea.KeyMsg) {
	switch {
	case key.Matches(km, m.keys.Left):
		if m.page > 0 {
			m.page--
		}
	case key.Matches(km, m.keys.Right):
		if !(m.page == 1) {
			m.page++
		}
	}
}

func (m *model) moveCursor(km tea.KeyMsg) {
	switch {
	case key.Matches(km, m.keys.Up):
		if m.cursor.index > 0 {
			m.cursor.index--
		}
	case key.Matches(km, m.keys.Down):
		if !(m.cursor.index+1 == len(m.todos)) {
			m.cursor.index++
		}
	}
}

/////////////
// Main
////////////

func main() {
	// TODO: read debug flag
	p := tea.NewProgram(NewModel(true))
	if err := p.Start(); err != nil {
		fmt.Println("There has been an error")
		os.Exit(1)
	}
}
