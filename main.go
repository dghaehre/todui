package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
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
	Up            key.Binding
	Down          key.Binding
	Left          key.Binding
	Right         key.Binding
	AllTasksTab   key.Binding
	CompletedTab  key.Binding
	Filter        key.Binding
	SetInput      key.Binding
	ClearInput    key.Binding
	ExitInput     key.Binding
	New           key.Binding
	NewWithEditor key.Binding
	CreateNewTask key.Binding
	Edit          key.Binding
	Sync          key.Binding
	Help          key.Binding
	Quit          key.Binding
}

type InputFieldCommand = string
type Command = string
type View = string
type Tab = int

const (
	allTasksTab Tab = iota
	completedTab

	// NOTE: make sure to increment counter if we add a new page
	totalTab = 2
)

var (
	inputFieldCommandNew    InputFieldCommand = "new"
	inputFieldCommandFilter InputFieldCommand = "filter"

	fetchedTodos Command = "fetchedTodos"

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
			key.WithHelp("/", "filter"),
		),
		AllTasksTab: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "All tasks tab"),
		),
		CompletedTab: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "completed tab"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		NewWithEditor: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "new in editor"),
		),
		CreateNewTask: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "set"),
		),
		SetInput: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "set"),
		),
		ClearInput: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "clear"),
		),
		ExitInput: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "exit"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
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
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		)}

	defaultTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15"))

	dimTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	chosenTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))

	projectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14"))

	labelsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5"))
)

type cursorPosition struct {
	view  View
	index int
}

type inputField struct {
	command InputFieldCommand
	content string
	enabled bool
}

type editorFinishedMsg struct{ err error }

type model struct {
	storage       Storage
	keys          keyMap
	totalWidth    int
	totalHeight   int
	listHeight    int
	debug         bool
	sync          bool
	todos         []Todo
	filteredTodos []Todo
	cursor        cursorPosition
	tab           Tab
	currentFilter string
	showHelp      bool
	textInput     textinput.Model
	inputField    inputField
	syncError     error
}

func NewModel(storage Storage, debug bool) model {
	ti := textinput.New()
	ti.CharLimit = 156
	ti.SetCursorMode(textinput.CursorStatic)
	ti.Placeholder = ""
	ti.Prompt = ""
	return model{
		storage:   storage,
		keys:      keys,
		debug:     debug,
		tab:       allTasksTab,
		textInput: ti,
		cursor: cursorPosition{
			index: 0,
		},
	}
}

func (m model) Init() tea.Cmd {
	return m.getLocalTodos
}

// Async functions

func (m model) getLocalTodos() tea.Msg {
	res, err := m.storage.localTodos()
	if err != nil {
		return SyncError{ // TODO: new error
			err: err,
		}
	}
	return LocalTodos{
		data: res,
	}
}

func (m model) fetchTodos() tea.Msg {
	todos, err := m.storage.fetchTodos()
	if err != nil {
		return SyncError{ // TODO: new error
			err: err,
		}
	}
	return FetchedTodos{
		data: todos,
	}
}

func (m model) quickAdd(content string) func() tea.Msg {
	return func() tea.Msg {
		todos, err := m.storage.quickAdd(content)
		if err != nil {
			return SyncError{ // TODO: new error
				err: err,
			}
		}
		return FetchedTodos{
			data: todos,
		}
	}
}

func (m model) newTask(todo Todo) func() tea.Msg {
	return func() tea.Msg {
		todos, err := m.storage.newTask(todo)
		if err != nil {
			return SyncError{ // TODO: new error
				err: err,
			}
		}
		return FetchedTodos{
			data: todos,
		}
	}
}
func newTaskInEditor(m model) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	path, err := createNewTaskFile()
	if err != nil {
		return nil
	}
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return editorFinishedMsg{err}
		}
		todo, err := parseTaskFile(path)
		if err != nil {
			return editorFinishedMsg{err}
		}
		return NewTask{
			data: todo,
		}
	})
}

func editTaskInEditor(m model) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	path, err := createEditFile(m)
	if err != nil {
		return nil
	}
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

/////////////
// Update
////////////

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case SyncError:
		m.syncError = msg.err
		return m, nil

	case NewTask:
		return m, m.newTask(msg.data)

	case LocalTodos:
		m.todos = msg.data
		m.filteredTodos = filterContents(msg.data, m.currentFilter)
		return m, m.fetchTodos

	case FetchedTodos:
		m.todos = msg.data
		m.filteredTodos = filterContents(msg.data, m.currentFilter)
		return m, nil

	// Set window size
	case tea.WindowSizeMsg:
		m.totalHeight = msg.Height
		m.totalWidth = msg.Width
		m.listHeight = 20
		return m, nil
	}

	if m.inputField.enabled {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {

			case "enter":
				value := m.textInput.Value()
				if m.inputField.command == inputFieldCommandFilter {
					m.currentFilter = value
					m.filteredTodos = filterContents(m.todos, value)
				}
				m.textInput.SetValue("")
				m.textInput.Prompt = ""
				m.inputField.enabled = false
				m.moveCursor(tea.KeyMsg{})
				if m.inputField.command == inputFieldCommandNew {
					m.inputField.command = ""
					return m, m.quickAdd(value)
				}
				m.inputField.command = ""
				return m, nil
			case "ctrl+c", "esc": // TODO: use keys instead (keymatches)
				m.textInput.SetValue("")
				m.textInput.Prompt = ""
				m.inputField.enabled = false

				// Delete current filter
				m.filteredTodos = m.todos
				m.currentFilter = ""
				m.moveCursor(tea.KeyMsg{})
				return m, nil
			}
		}
		// Handle text input..
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)

		if m.inputField.command == inputFieldCommandFilter {
			m.filteredTodos = filterContents(m.todos, m.textInput.Value())
		}

		return m, cmd
	}

	// Normal list view
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Edit):
			return m, editTaskInEditor(m)
		case key.Matches(msg, m.keys.NewWithEditor):
			return m, newTaskInEditor(m)
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
	content := fmt.Sprintf("h: %d, w: %d, Cursor: %+v, Tab: %+v, todos: %d, filteredTodos: %d, syncError: %+v",
		m.totalHeight, m.totalWidth, m.cursor, m.tab, len(m.todos), len(m.filteredTodos), m.syncError)
	style := lipgloss.NewStyle().
		Width(m.totalWidth).
		Align(lipgloss.Right)
	return style.Render(content) + "\n"
}

func (m model) getMainList() string {
	switch m.tab {
	case allTasksTab:
		return m.allTasksList()
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
	pageStyle := lipgloss.NewStyle().
		Width(m.totalWidth)

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

	tabStyle := lipgloss.NewStyle().
		Height(2).
		Width(m.totalWidth / 2).
		Align(lipgloss.Left)

	content := tabStyle.Render(s) + m.showError()
	return pageStyle.Render(content) + "\n" + "\n"
}

// TODO: create a notifcation popup ish thing.
func (m model) showError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")).
		Width(m.totalWidth / 3).
		Height(1).
		Align(lipgloss.Right)
	var e string
	if m.syncError != nil {
		e += strings.TrimSpace(fmt.Sprintf("%s", m.syncError))
	}
	return errorStyle.Render(e)
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
	ks := m.getValidKeys(m.keys)
	for i, v := range ks {
		h := v.Help()
		s += h.Key + ": " + h.Desc
		if (i + 1) != len(ks) {
			s += "   "
		}
	}
	return input + style.Render(s)
}

func (m model) allTasksList() string {
	// We currently assume that the cursor position has view default
	content := ""
	style := lipgloss.NewStyle().
		Width(m.totalWidth).
		Align(lipgloss.Left)

	for i, v := range m.filteredTodos {
		if m.cursor.index == i {
			content += "→ " + v.renderInList(m.totalWidth)
		} else {
			content += "  " + v.renderInList(m.totalWidth)
		}
		content += "\n"
		if i == m.listHeight {
			break
		}
	}
	return style.Render(content)
}

func (t Todo) renderInList(w int) string {
	desc := defaultTextStyle.Render(withSize(t.Content, w-50))

	labels := ""
	for _, l := range t.Labels {
		labels += labelsStyle.Render(" @" + l)
	}

	// TODO: remove emojis
	project := ""
	if t.ProjectName != "" {
		project = projectStyle.Render(" #" + t.ProjectName)
	}

	children := ""
	totalChildren := len(t.Children)
	if totalChildren > 0 {
		completedChildren := totalChildren // TODO
		children += dimTextStyle.Render(fmt.Sprintf(" (%d/%d)", completedChildren, totalChildren))
	}
	return desc + " " + project + labels + children
}

/////////////
// Utils
////////////

// Returns the valid command keys to be used, based on current state
func (m model) getValidKeys(k keyMap) []key.Binding {
	if m.inputField.enabled {
		return []key.Binding{k.SetInput, k.ClearInput, k.ExitInput}
	}
	if m.showHelp {
		return []key.Binding{k.Sync, k.New, k.NewWithEditor, k.Edit, k.Filter, k.Up, k.Down, k.AllTasksTab, k.CompletedTab, k.Help, k.Quit}
	}
	return []key.Binding{k.Help, k.Quit}
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

func (m *model) moveCursor(km tea.KeyMsg) {
	bottom := m.listHeight
	if len(m.filteredTodos) < bottom {
		bottom = len(m.filteredTodos) - 1
	}
	switch {
	case key.Matches(km, m.keys.Up):
		if m.cursor.index > 0 {
			m.cursor.index--
		}
	case key.Matches(km, m.keys.Down):
		if !(m.cursor.index >= bottom) {
			m.cursor.index++
		}
	default:
		if m.cursor.index > bottom {
			m.cursor.index = bottom
		}
	}
}

/////////////
// Main
////////////

func main() {

	// TODO: only open "tui" app when no args are given
	// TODO: create a sync command for completed tasks

	var tokenPath string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	defaultTokenPath := homeDir + "/.todoist.token"
	flag.StringVar(&tokenPath, "t", defaultTokenPath, "Path to todoist API token.")

	var dbPath string
	flag.StringVar(&dbPath, "d", homeDir+"/.cache/todui.db", "Path to local db.")

	debug := flag.Bool("debug", false, "Run tui in debug mode")

	sync := flag.Bool("sync", false, "do a full sync to local db")

	flag.Parse()

	db, err := NewDB(dbPath)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	defer db.Close()

	api, err := NewAPI(tokenPath)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	storage := Storage{
		api: api,
		db:  db,
	}

	if *sync == true {
		// run sync command
		fmt.Println("TODO")
		return
	}

	model := NewModel(storage, *debug)

	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
