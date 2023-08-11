package main

import "time"

type FetchedTodos struct {
	data []Todo
}

type LocalTodos struct {
	data []Todo
}

type NewTask struct {
	data Todo
}

type EditTask struct {
	data Todo
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

func (t Todo) DueToday() bool {
	return t.Due.Date == time.Now().Format("2006-01-02")
}

func (t Todo) DueDisplay() string {
	if t.Due.Date == "" {
		return ""
	}
	parsed, err := time.Parse("2006-01-02", t.Due.Date)
	if err != nil {
		return t.Due.Date
	}
	if parsed.Before(time.Now().AddDate(0, 0, 7)) {
		if parsed.Day() == time.Now().Day() {
			return "Today"
		}
		return parsed.Weekday().String()
	}
	return t.Due.Date
}

type Due struct {
	String      string `json:"string"`
	Date        string `json:"date"`
	Lang        string `json:"lang"`
	IsRecurring bool   `json:"is_recurring"`
	Timezone    string `json:"timezone,omitempty"`
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
