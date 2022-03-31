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
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	NextPage key.Binding
	PrevPage key.Binding
	New      key.Binding
	Help     key.Binding
	Quit     key.Binding
}

type InputMode = string
type Mode = string
type View = string
type Page = uint8

const (
	allTasksPage Page = iota
	projectsPage
	inboxPage
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
		NextPage: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]", "next page"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[", "previous page"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
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

	projectStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("5"))

)

type cursorPosition struct {
	view  View
	index int
}

type model struct {
	keys           keyMap
	inputMode      InputMode
	mode           Mode
	totalWidth     int
	totalHeight    int
	debug          bool
	todos          []Todo
	storage        Storage
	cursor         cursorPosition
	page           Page
	currentProject string
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
		storage: NewStorage(),
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
		case key.Matches(msg, m.keys.PrevPage), key.Matches(msg, m.keys.NextPage):
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
	content := m.getMainList()
	return m.debugView() + top + content + m.getEmptyLines(content+top) + m.bottomBar()
}

func (m model) debugView() string {
	if !m.debug {
		return ""
	}
	content := fmt.Sprintf("STATE: Height: %d, Width: %d, InputMode: %+v, Mode: %+v, Cursor: %+v, Page: %+v, Project: %+v",
		m.totalHeight, m.totalWidth, m.inputMode, m.mode, m.cursor, m.page, m.currentProject)
	style := lipgloss.NewStyle().
		Width(m.totalWidth).
		Align(lipgloss.Right)
	return style.Render(content) + "\n"
}

func (m model) getMainList() string {
	switch m.page {
	case projectsPage:
		return m.projectsList()
	case allTasksPage:
		return m.defaultList()
	default:
		return ""
	}
}

func (m model) topBar() string {
	style := lipgloss.NewStyle().
		Width(m.totalWidth).
		// Foreground(lipgloss.Color("13")).
		Align(lipgloss.Left)

	var s string

	if m.page == allTasksPage {
		s += chosenTextStyle.Render("All tasks")
	} else {
		s += dimTextStyle.Render("All tasks")
	}

	s += " | "

	// projects
	if m.page == projectsPage {
		s += chosenTextStyle.Render("By project")
	} else {
		s += dimTextStyle.Render("By project")
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

	for i, v := range m.storage.getPendingTodos(m.currentProject) {
		if m.cursor.index == i {
			content += "→ " + v.renderInList()
		} else {
			content += "  " + v.renderInList()
		}
		content += "\n"
	}
	return style.Render(content)
}

func (t Todo) renderInList() string {
	desc := defaultTextStyle.Render(t.desc)
	project := ""
	if t.project.Name != "" {
		project = projectStyle.Render("#" + t.project.Name)
	}
	return desc + " " + project
}

func (m model) projectsList() string {
	// We currently assume that the cursor position has view default
	content := ""
	style := lipgloss.NewStyle().
		Width(m.totalWidth).
		Align(lipgloss.Left)

	for i, p := range m.storage.getPendingProjects() {
		if m.cursor.index == i {
			content += "→ " + projectStyle.Render(p.Name)
		} else {
			content += "  " + projectStyle.Render(p.Name)
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
		return []key.Binding{k.New, k.Up, k.Down, k.PrevPage, k.NextPage, k.Help, k.Quit}
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
	case key.Matches(km, m.keys.PrevPage):
		if m.page <= 1 {
			m.currentProject = ""
		}
		if m.page > 0 {
			m.page--
		}
	case key.Matches(km, m.keys.NextPage):
		if m.page == 0 {
			m.setCurrentProject()
			m.page++
		}
	}
}

// TODO: fiks
func (m *model) setCurrentProject() {
	m.currentProject = m.storage.getTodoProject(m.cursor.index)
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
