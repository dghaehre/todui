package main

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
}

type Item struct {
	// Due         string   `json:"due,omitempty"`
	Id          string   `json:"id"`
	ProjectId   string   `json:"project_id"`
	Content     string   `json:"content"`
	Description string   `json:"description"`
	Priority    int      `json:"priority"`
	ParentId    string   `json:"parent_id"`
	Labels      []string `json:"labels"`
	Checked     bool     `json:"checked"`
}

// "id": "2995104339",
// "user_id": "2671355",
// "project_id": "2203306141",
// "content": "Buy Milk",
// "description": "",
// "priority": 1,
// "due": null,
// "parent_id": null,
// "child_order": 1,
// "section_id": null,
// "day_order": -1,
// "collapsed": false,
// "labels": ["Food", "Shopping"],
// "added_by_uid": "2671355",
// "assigned_by_uid": "2671355",
// "responsible_uid": null,
// "checked": false,
// "is_deleted": false,
// "sync_id": null,
// "added_at": "2014-09-26T08:25:05.000000Z"
