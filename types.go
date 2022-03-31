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
	Id           int
	Name         string
	CommentCount int
	Order        int
	Color        int
	Shared       bool
	InboxProject bool
	Url          string
}

type Todo struct {
	desc        string
	projectName string
	project     Project
}

type Storage struct {
	pendingTodos   []Todo
	completedTodos []Todo
  pendingProjects []Project
}


// TODO: create interfaces that is needed to sync a project and/or a todo
