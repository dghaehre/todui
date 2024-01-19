package main

import (
	"strings"
	"time"
)

type FetchedTodos struct {
	data []Todo
}

type LocalTodos struct {
	data []Todo
}

type NewTask struct {
	data Todo
}

type UpdateStatus int

const (
	UpdateStatusModified UpdateStatus = iota
	UpdateStatusDeleted
	UpdateStatusNew
)

type UpdateChild struct {
	Org     Todo
	Checked bool
	Content string
	UpdateStatus
}

type EditTaskData struct {
	todo           Todo
	updateChildren []UpdateChild
}

// TODO: add children as well somewhow
type EditTask struct {
	data EditTaskData
}

// ^
// Should probably rename these structs.
// They currently work as following:
// - LocalTodos will in the update function trigger fetchTodos
// - FetchedTodos will not trigger any other function

type SyncError struct {
	err error
}

type Project struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	CommentCount int    `json:"comment_count"`
	Order        int    `json:"order"`
	Color        string `json:"color"`
	Shared       bool   `json:"shared"`
	SyncId       string `json:"sync_id"`
	Favorite     bool   `json:"favorite"`
	InboxProject bool   `json:"inbox_project"`
	Url          string `json:"url"`
}

// Todo might need to be an interface.. because CompletedItem looks very different..
type Todo struct {
	Id          string
	ProjectId   string
	ProjectName string
	Content     string
	Description string
	Priority    int
	Labels      []string
	Checked     bool
	Children    []Todo
	Due         Due
}

func (t Todo) DueTodayOrBefore() bool {
	due, err := time.Parse("2006-01-02", t.Due.Date)
	if err != nil {
		return false
	}
	return !due.After(time.Now())
}

func (t Todo) DueDisplay(withExtraInfo bool) string {
	width := 10
	if withExtraInfo {
		width = 30
	}
	if t.Due.Date == "" {
		return dueDateStyle.Width(width).Render("")
	}
	result := ""
	overdue := false
	dateTimeOnly := strings.Split(t.Due.Date, "T")[0]
	parsed, err := time.Parse("2006-01-02", dateTimeOnly)
	if err != nil {
		return dueDateStyle.Width(10).Render(t.Due.Date)
	}
	if parsed.Before(time.Now().AddDate(0, 0, 7)) {
		today := time.Now().Day()
		switch parsed.Day() {
		case today:
			result = "Today"
		case today + 1:
			result = "Tomorrow"
		case today - 1:
			result = "Yesterday"
			overdue = true
		default:
			result = parsed.Weekday().String()
			overdue = parsed.Before(time.Now())
		}
	} else {
		result = parsed.Format("02/01/2006")
		overdue = parsed.Before(time.Now())
	}

	if withExtraInfo {
		result += ", " + t.Due.String
	}

	if overdue {
		return dueDateOverdueStyle.Width(width).Render(result)
	} else {
		return dueDateStyle.Width(width).Render(result)
	}
}

func (t Todo) ProjectDisplay(projectNameLength int) string {
	project := " "
	if t.ProjectName != "" {
		project = "#" + t.ProjectName
	}
	return projectStyle.Width(projectNameLength + 1).Render(project)
}

type Due struct {
	ChangeString string // Used for editing a due date with natural language
	String       string `json:"string"`
	Date         string `json:"date"`
	Lang         string `json:"lang"`
	IsRecurring  bool   `json:"is_recurring"`
	Timezone     string `json:"timezone,omitempty"`
}

type Item struct {
	Id          string   `json:"id"`
	ProjectId   string   `json:"project_id"`
	Content     string   `json:"content"`
	Description string   `json:"description"`
	Priority    int      `json:"priority"`
	ParentId    string   `json:"parent_id"`
	Labels      []string `json:"labels"`
	Checked     bool     `json:"checked"`
	Due         Due      `json:"due"`
}

type CompletedItem struct {
	Id            string `json:"id"`
	ProjectId     string `json:"project_id"`
	Content       string `json:"content"`
	MetaData      string `json:"meta_data"`
	CompletedDate string `json:"completed_date"`
}
