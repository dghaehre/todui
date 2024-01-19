package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
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
	Top           key.Binding
	Bottom        key.Binding
	AllTasksTab   key.Binding
	CompletedTab  key.Binding
	TodayTab      key.Binding
	InboxTab      key.Binding
	Info          key.Binding
	Done          key.Binding
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
	inboxTab Tab = iota
	todayTab
	allTasksTab
	completedTab

	// NOTE: make sure to increment counter if we add a new page
	totalTab = 4
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
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "to the top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "to the bottom"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Info: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "task info"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		InboxTab: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "inbox tab"),
		),
		TodayTab: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "today tab"),
		),
		AllTasksTab: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "All tasks tab"),
		),
		CompletedTab: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "completed tab"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		Done: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "Mark as done"),
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

	cursorBackgroundStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("8"))

	dimTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	chosenTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))

	projectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14"))

	dueDateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5"))

	dueDateOverdueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("1"))

	labelsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5"))

	p1Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("1"))

	p2Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("3"))

	p3Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("4"))
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
	storage        Storage
	keys           keyMap
	totalWidth     int
	totalHeight    int
	listHeight     int
	debug          bool
	syncing        bool
	todos          []Todo
	filteredTodos  []Todo
	todayTodos     []Todo
	inboxTodos     []Todo
	completedTodos []Todo
	cursor         cursorPosition
	tab            Tab
	currentFilter  string
	showHelp       bool
	showInfo       bool
	textInput      textinput.Model
	inputField     inputField
	syncError      error
}

func NewModel(storage Storage, debug bool) model {
	ti := textinput.New()
	ti.CharLimit = 156
	ti.Cursor.SetMode(cursor.CursorStatic)
	ti.Placeholder = ""
	ti.Prompt = ""
	return model{
		storage:   storage,
		keys:      keys,
		debug:     debug,
		tab:       todayTab,
		showInfo:  true,
		textInput: ti,
		syncing:   true, // always try to sync on startup
		cursor: cursorPosition{
			index: 0,
		},
	}
}

func (m model) Init() tea.Cmd {
	return m.getLocalTodos
}

// Async functions

// Will eventually call fetchTodos
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

func (m model) markAsDone(todo Todo) func() tea.Msg {
	return func() tea.Msg {
		todos, err := m.storage.markAsDone(todo)
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

func (m model) editTask(data EditTaskData) func() tea.Msg {
	return func() tea.Msg {
		todos, err := m.storage.editTask(data)
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

// TODO:
// handle error if cursor is out of scope
// handle multiple pages
func (m model) getCurrentTodo() (Todo, error) {
	var todos []Todo
	switch m.tab {
	case todayTab:
		todos = m.todayTodos
	case inboxTab:
		todos = m.inboxTodos
	case completedTab:
		todos = m.completedTodos
	default:
		todos = m.filteredTodos
	}

	if m.cursor.index >= len(todos) {
		return Todo{}, fmt.Errorf("cursor out of scope")
	}
	todo := todos[m.cursor.index]
	return todo, nil
}

func editTaskInEditor(todo Todo, path string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // Always! 💪
	}
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return editorFinishedMsg{err}
		}
		todo, updateChildren, err := parseEditFile(path, todo)
		if err != nil {
			return editorFinishedMsg{err}
		}
		return EditTask{
			data: EditTaskData{
				todo:           todo,
				updateChildren: updateChildren,
			},
		}
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

	case EditTask:
		return m, m.editTask(msg.data)

	case LocalTodos:
		m.todos = msg.data
		filtered := filterContents(msg.data, m.currentFilter)
		sort.Sort(ByDueThenPriority(filtered))
		m.filteredTodos = filtered
		m.todayTodos = filterToday(m.filteredTodos)
		m.inboxTodos = filterInbox(m.filteredTodos)
		return m, m.fetchTodos

	case FetchedTodos:
		m.todos = msg.data
		filtered := filterContents(msg.data, m.currentFilter)
		sort.Sort(ByDueThenPriority(filtered))
		m.filteredTodos = filtered
		m.todayTodos = filterToday(m.filteredTodos)
		m.inboxTodos = filterInbox(m.filteredTodos)
		m.syncing = false
		return m, nil

	// Set window size
	case tea.WindowSizeMsg:
		m.totalHeight = msg.Height
		m.totalWidth = msg.Width
		m.listHeight = msg.Height - 8
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
					m.todayTodos = filterToday(m.filteredTodos)
					m.inboxTodos = filterInbox(m.filteredTodos)
				}
				m.textInput.SetValue("")
				m.textInput.Prompt = ""
				m.inputField.enabled = false
				m.refreshCursor()
				if m.inputField.command == inputFieldCommandNew {
					m.inputField.command = ""
					m.syncing = true
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
				m.todayTodos = filterToday(m.filteredTodos)
				m.inboxTodos = filterInbox(m.filteredTodos)
				m.moveCursor(tea.KeyMsg{})
				return m, nil
			}
		}
		// Handle text input..
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)

		if m.inputField.command == inputFieldCommandFilter {
			m.filteredTodos = filterContents(m.todos, m.textInput.Value())
			m.todayTodos = filterToday(m.filteredTodos)
			m.inboxTodos = filterInbox(m.filteredTodos)
		}

		return m, cmd
	}

	// Normal list view
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Edit):
			todo, err := m.getCurrentTodo()
			if err != nil {
				// handle error
			}
			path, err := createEditFile(todo)
			if err != nil {
				// TODO: handle error
			}
			m.syncing = true
			return m, editTaskInEditor(todo, path)
		case key.Matches(msg, m.keys.NewWithEditor):
			return m, newTaskInEditor(m)
		case key.Matches(msg, m.keys.Down, m.keys.Up, m.keys.Bottom, m.keys.Top):
			m.moveCursor(msg)
		case key.Matches(msg, m.keys.AllTasksTab), key.Matches(msg, m.keys.CompletedTab), key.Matches(msg, m.keys.TodayTab), key.Matches(msg, m.keys.InboxTab):
			m.changeTab(msg)
			m.refreshCursor()
		case key.Matches(msg, m.keys.Done):
			todo, err := m.getCurrentTodo()
			if err != nil {
				// handle error
			}
			m.syncing = true
			return m, m.markAsDone(todo)
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		case key.Matches(msg, m.keys.Info):
			m.showInfo = !m.showInfo
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
			m.syncing = true
			m.syncError = nil
			return m, m.fetchTodos
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
	mainList := m.getMainList()
	content := m.renderViewList(mainList)
	return m.debugView() + top + content + m.getEmptyLines(content+top) + m.bottomBar()
}

func (m model) debugView() string {
	if !m.debug {
		return ""
	}
	content := fmt.Sprintf("h: %d, w: %d, Cursor: %+v, Tab: %+v, todos: %d, todayTodos: %d, filteredTodos: %d, syncError: %+v",
		m.totalHeight, m.totalWidth, m.cursor, m.tab, len(m.todos), len(m.todayTodos), len(m.filteredTodos), m.syncError)
	style := lipgloss.NewStyle().
		Width(m.totalWidth).
		Align(lipgloss.Right)
	return style.Render(content) + "\n"
}

func (m model) getMainList() []Todo {
	switch m.tab {
	case allTasksTab:
		return m.filteredTodos
	case todayTab:
		return m.todayTodos
	case inboxTab:
		return m.inboxTodos
	case completedTab:
		return m.completedTodos
	default:
		return []Todo{}
	}
}

func (m model) tabToString(p Tab) string {
	switch p {
	case completedTab:
		return "Completed tasks"
	case allTasksTab:
		return "All tasks"
	case inboxTab:
		inboxCount := len(m.inboxTodos)
		if inboxCount > 0 {
			return "Inbox " + chosenTextStyle.Render(fmt.Sprintf("(%d)", inboxCount))
		}
		return "Inbox"
	case todayTab:
		// TODO: need completed tasks (today) to display correctly here..
		extra := ""
		style := dimTextStyle
		todayCount := len(m.todayTodos)
		if todayCount != 0 {
			extra = " " + fmt.Sprintf("(%d/%d)", 0, todayCount)
		}
		if m.tab == todayTab {
			style = p3Style
		}
		return "Today tasks" + style.Render(extra)
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
			s += chosenTextStyle.Render(m.tabToString(i))
		} else {
			s += dimTextStyle.Render(m.tabToString(i))
		}

		if i != (totalTab - 1) {
			s += " | "
		}
	}

	if m.syncing {
		s += "  " + dimTextStyle.Render("syncing...")
	}

	s += "\n"
	if m.currentFilter != "" {
		s += chosenTextStyle.Render("  filter: on")
	} else {
		s += dimTextStyle.Render("  filter: off")
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
		Width(m.totalWidth - w). // m.totalWidth might be 0
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

func (m model) renderInfo(todo Todo, height int) string {
	rows := [][]string{
		{"title", todo.Content},
		{"description", todo.Description},
		{"project", todo.ProjectDisplay(m.totalWidth - 30)},
		{"due", todo.DueDisplay(true)},
		{"priority", displayPrioriy(todo.Priority)},
		{"labels", strings.Join(todo.Labels, ", ")},
	}
	if len(todo.Children) > 0 {
		rows = append(rows, []string{"children", todo.Children[0].Content})
		for _, child := range todo.Children[1:] {
			rows = append(rows, []string{"", child.Content})
		}
	}
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("8"))).
		BorderRow(true).
		Height(height).
		Width(m.totalWidth / 2).
		Rows(rows...)
	return t.Render()
}

func (m model) renderViewList(todos []Todo) string {
	projectLength := projectNameSize(todos, 30)
	content := ""
	showing := 0
	info := ""
	width := m.totalWidth
	if m.showInfo {
		width = (m.totalWidth / 2)
	}
	listHeight := m.listHeight - 1
	for i, v := range todos {
		if m.cursor.index == i {
			content += "→ " + v.renderInList(width, projectLength)
			if m.showInfo {
				info = m.renderInfo(v, m.totalHeight) + "\n"
			}
		} else {
			content += "  " + v.renderInList(width, projectLength)
		}
		content += "\n"
		showing++
		if i >= listHeight {
			break
		}
	}
	content += fmt.Sprintf("\nshowing %d of %d", showing, len(todos))
	if m.showInfo {
		content = lipgloss.NewStyle().
			Width(m.totalWidth / 2).
			Render(content)

		info = lipgloss.NewStyle().
			Width(m.totalWidth / 2).
			Render(info)

		return lipgloss.JoinHorizontal(lipgloss.Top, content, info)
	}
	return content
}

// Add strikethrough if the task is completed/checked
func (t Todo) renderInList(w int, projectNameLength int) string {
	desc := defaultTextStyle.Render(withSize(t.Content, w-50))
	labels := ""
	for _, l := range t.Labels {
		labels += labelsStyle.Render(" @" + l)
	}
	project := t.ProjectDisplay(projectNameLength)
	due := t.DueDisplay(false)
	rec := "  "
	if t.Due.IsRecurring {
		rec = "↻ "
	}
	priority := displayPrioriy(t.Priority)
	children := ""
	totalChildren := len(t.Children)
	if totalChildren > 0 {
		children += dimTextStyle.Render(fmt.Sprintf(" (%d)", totalChildren))
	}
	result := project + " " + rec + " " + due + " " + priority + " " + desc + " " + labels + children
	if lipgloss.Width(result) > w {
		// Dont show labels if it doesnt fit
		return project + " " + rec + " " + due + " " + priority + " " + desc + " " + children
	}
	return result
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
		return []key.Binding{k.Sync, k.New, k.NewWithEditor, k.Edit, k.Done, k.Filter, k.Up, k.Down, k.Top, k.Bottom, k.AllTasksTab, k.CompletedTab, k.Help, k.Quit}
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
	case key.Matches(km, m.keys.InboxTab):
		m.tab = 0
	case key.Matches(km, m.keys.TodayTab):
		m.tab = 1
	case key.Matches(km, m.keys.AllTasksTab):
		m.tab = 2
	case key.Matches(km, m.keys.CompletedTab):
		m.tab = 3
	}
}

func (m *model) refreshCursor() {
	maxIndex := len(m.getMainList()) - 1
	if maxIndex < 0 { // No elements in list (doesnt matter then)
		return
	}
	if m.cursor.index > maxIndex {
		m.cursor.index = maxIndex
	}
}

func (m *model) moveCursor(km tea.KeyMsg) {
	maxIndex := len(m.getMainList()) - 1
	if maxIndex <= 0 {
		return
	}

	if m.cursor.index > maxIndex {
		m.cursor.index = maxIndex
		return
	}

	switch {
	case key.Matches(km, m.keys.Up):
		if m.cursor.index > 0 {
			m.cursor.index--
		}
	case key.Matches(km, m.keys.Top):
		m.cursor.index = 0
	case key.Matches(km, m.keys.Bottom):
		m.cursor.index = maxIndex
	case key.Matches(km, m.keys.Down):
		if !(m.cursor.index >= maxIndex) {
			m.cursor.index++
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
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
