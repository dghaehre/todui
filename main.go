package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/////////////////
// Model
/////////////////

// Keys
type keyMap struct {
	Up           key.Binding
	Down         key.Binding
	Left         key.Binding
	Right        key.Binding
	AllTasksTab  key.Binding
	CompletedTab key.Binding
	Filter       key.Binding
	New          key.Binding
	Sync         key.Binding
	Help         key.Binding
	Quit         key.Binding
}

type InputMode = string
type InputFieldCommand = string
type Command = string
type Mode = string
type View = string
type Tab = int

const (
	allTasksTab Tab = iota
	completedTab

	// NOTE: make sure to increment counter if we add a new page
	totalTab = 2
)

var (
	insertInputMode InputMode = "insert"
	normalInputMode InputMode = "normal"
	defaultMode     Mode      = "default"

	inputFieldCommandNew    InputFieldCommand = "new"
	inputFieldCommandFilter InputFieldCommand = "filter"

	fetchedTodos Command = "fetchedTodos"

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
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "Update filter"),
		),
		AllTasksTab: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "All tasks tab"),
		),
		CompletedTab: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "Completed tab"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		Sync: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sync"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
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

type FetchedTodos struct {
	data []Todo
}

type cursorPosition struct {
	view  View
	index int
}

type inputField struct {
	command InputFieldCommand
	content string
	enabled bool
}

type model struct {
	keys           keyMap
	inputMode      InputMode
	mode           Mode
	totalWidth     int
	totalHeight    int
	debug          bool
	sync           bool
	todos          []Todo
	todosInView    []Todo
	storage        Storage
	cursor         cursorPosition
	tab            Tab
	currentFilter  string
	showHelp       bool
	currentProject string
	textInput      textinput.Model
	inputField     inputField
}

func NewModel(debug bool) model {
	ti := textinput.New()
	ti.CharLimit = 156
	ti.SetCursorMode(textinput.CursorStatic)
	ti.Placeholder = ""
	ti.Prompt = ""
	return model{
		keys:      keys,
		mode:      defaultMode,
		inputMode: normalInputMode,
		debug:     debug,
		tab:       allTasksTab,
		textInput: ti,
		cursor: cursorPosition{
			view:  defaultView,
			index: 0,
		},
		storage: NewStorage(),
	}
}

func (m model) Init() tea.Cmd {
	return m.fetchTodos
}

func (m model) fetchTodos() tea.Msg {
	todos := m.storage.getPendingTodos("")
	return FetchedTodos{
		data: todos,
	}
}

/////////////
// Update
////////////

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case FetchedTodos:
		m.todos = msg.data
		m.todosInView = filterContents(msg.data, m.currentFilter)
		return m, nil

	// Set window size
	case tea.WindowSizeMsg:
		m.totalHeight = msg.Height
		m.totalWidth = msg.Width
		return m, nil
	}

	if m.inputField.enabled {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {

			// TODO: handle
			case "enter":
				m.currentFilter = m.textInput.Value()
				m.textInput.SetValue("")
				m.textInput.Prompt = ""
				m.inputField.enabled = false
				return m, nil
			case "ctrl+c":
				m.textInput.SetValue("")
				m.textInput.Prompt = ""
				m.inputField.enabled = false

				// Delete current filter
				m.todosInView = m.todos
				m.currentFilter = ""
				return m, nil
			}
		}
		// Handle text input..
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)

		if m.inputField.command == inputFieldCommandFilter {
			m.todosInView = filterContents(m.todos, m.textInput.Value())
		}

		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Down), key.Matches(msg, m.keys.Up):
			m.moveCursor(msg)
		case key.Matches(msg, m.keys.AllTasksTab), key.Matches(msg, m.keys.CompletedTab):
			m.changeTab(msg)
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		case key.Matches(msg, m.keys.Filter):
			m.textInput.Focus()
			m.textInput.SetValue(m.currentFilter)
			m.textInput.Placeholder = ""
			// m.textInput.TextStyle = defaultTextStyle
			m.textInput.Prompt = "/"
			m.inputField.enabled = true
			m.inputField.command = inputFieldCommandFilter
			return m, nil
		case key.Matches(msg, m.keys.New):
			m.textInput.Focus()
			m.textInput.SetValue("")
			m.textInput.Placeholder = ""
			m.textInput.Prompt = "new task: "
			m.inputField.enabled = true
			m.inputField.command = inputFieldCommandNew
			return m, nil
		case key.Matches(msg, m.keys.Sync):
			m.sync = true
			return m, nil
			// m.changeTab(msg)
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
	content := fmt.Sprintf("Height: %d, Width: %d, InputMode: %+v, Mode: %+v, Cursor: %+v, Tab: %+v, Project: %+v, todos: %d, todosInView: %d",
		m.totalHeight, m.totalWidth, m.inputMode, m.mode, m.cursor, m.tab, m.currentProject, len(m.todos), len(m.todosInView))
	style := lipgloss.NewStyle().
		Width(m.totalWidth).
		Align(lipgloss.Right)
	return style.Render(content) + "\n"
}

func (m model) getMainList() string {
	switch m.tab {
	case allTasksTab:
		return m.defaultList()
	default:
		return "TODO"
	}
}

func tabToString(p Tab) string {
	switch p {
	case completedTab:
		return "Completed tasks"
	case allTasksTab:
		return "All tasks"
	}
	return ""
}

func (m model) topBar() string {
	style := lipgloss.NewStyle().
		Width(m.totalWidth).
		Align(lipgloss.Left)

	var s string
	s += "  "

	for i := 0; i < totalTab; i++ {
		if m.tab == i {
			s += chosenTextStyle.Render(tabToString(i))
		} else {
			s += dimTextStyle.Render(tabToString(i))
		}

		if i != (totalTab - 1) {
			s += " | "
		}
	}
	return style.Render(s) + "\n" + "\n"
}

func (m model) bottomBar() string {
	var input string
	if m.inputField.enabled {
		inputStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Align(lipgloss.Left)
		filterStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Align(lipgloss.Left)
		if m.inputField.command == inputFieldCommandFilter {
			input = filterStyle.Render(m.textInput.View())
		} else {
			input = inputStyle.Render(m.textInput.View())
		}
	}
	w, _ := lipgloss.Size(input)

	style := lipgloss.NewStyle().
		Width(m.totalWidth - w).
		Foreground(lipgloss.Color("12")).
		Align(lipgloss.Right)

	var s string
	if m.showHelp {
		ks := m.getValidKeys(m.keys)
		for i, v := range ks {
			h := v.Help()
			s += h.Key + ": " + h.Desc
			if (i + 1) != len(ks) {
				s += "   "
			}
		}
	} else {
		h := m.keys.Help.Help()
		s += h.Key + ": " + h.Desc
		s += "   "
		q := m.keys.Quit.Help()
		s += q.Key + ": " + q.Desc
	}
	return input + style.Render(s)
}

func (m model) defaultList() string {
	// We currently assume that the cursor position has view default
	content := ""
	style := lipgloss.NewStyle().
		Width(m.totalWidth).
		Align(lipgloss.Left)

	for i, v := range m.todosInView {
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

// TODO: remove
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
		return []key.Binding{k.Sync, k.New, k.Filter, k.Up, k.Down, k.AllTasksTab, k.CompletedTab, k.Help, k.Quit}
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

func (m *model) changeTab(km tea.KeyMsg) {
	switch {
	case key.Matches(km, m.keys.AllTasksTab):
		m.tab = 0
	case key.Matches(km, m.keys.CompletedTab):
		m.tab = 1
	}
}

// TODO: remove
// NOTE: hva gjør denne egentlig?
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
		// TODO: change bottom pased on page
		bottom := len(m.todosInView)
		if !(m.cursor.index+1 == bottom) {
			m.cursor.index++
		}
	}
}

/////////////
// Main
////////////

func main() {

	// TODO: read debug flag
	// TODO: only open "tui" app when no args are given
	// TODO: create a sync command for completed tasks

	var tokenPath string
	flag.StringVar(&tokenPath, "t", "~/.todoist.token", "Path to todoist API token.")
	var dbPath string
	flag.StringVar(&dbPath, "d", "~/.cache/todui.db", "Path to local db.")
	sync := flag.Bool("sync", false, "do a full sync to local db")
	flag.Parse()

	// db, err := newDB()
	// api := newAPI()

	if *sync == true {
		// run sync command
		fmt.Println("TODO")
		return
	}

	model := NewModel(true)

	// Launch tui
	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
