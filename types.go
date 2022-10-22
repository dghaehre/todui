package main

// Todoist project:
//     "id": 220474322,
//     "name": "Inbox",
//     "comment_count": 10,
//     "order": 1,
//     "color": 47,
//     "shared": false,
//     "sync_id": 0,
//     "favorite": false,
//     "inbox_project": true,
//     "url": "https://todoist.com/showProject?id=220474322"
// }

type Project struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	CommentCount int    `json:"comment_count"`
	Order        int    `json:"order"`
	Color        int    `json:"color"`
	Shared       bool   `json:"shared"`
	SyncId       int    `json:"sync_id"`
	Favorite     bool   `json:"favorite"`
	InboxProject bool   `json:"inbox_project"`
	Url          string `json:"url"`
}

type Todo struct {
	desc        string
	projectName string
	project     Project
}

type Storage struct {
	pendingTodos     []Todo
	completedTodos   []Todo
	pendingProjects  []Project
	completedProject []Project
}

// TODO: create interfaces that is needed to sync a project and/or a todo
