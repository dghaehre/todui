package main

// TODO: what needs to be storage, and what is "logic"?
// And maybe try to make this usesable for todoist?
//
// Maybe I should just make this todoist compatible for the get go..?

func NewStorage() Storage {
	return Storage{
		pendingTodos: mockTodos,
    pendingProjects: mockProjects,
	}
}

var (
	mockProjects []Project = []Project{
		{Id: 0, Name: "work"},
		{Id: 1, Name: "home"},
	}

	mockTodos []Todo = []Todo{
		{desc: "a test", project: mockProjects[1]},
		{desc: "get a dog", project: mockProjects[0]},
		{desc: "another test"},
	}
)

func (s *Storage) getPendingProjects() []Project {
	// for _, v := range s.pendingTodos {
	// 	if !Contains(projects, v.project) {
	// 		projects = append(projects, v.project)
	// 	}
	// }
	return s.pendingProjects
}

// project param might be empty, if so return all todos
func (s *Storage) getPendingTodos(project string) []Todo {
	if project == "" {
		return s.pendingTodos
	}

	var list []Todo
	for _, v := range s.pendingTodos {
		if v.projectName == project {
			list = append(list, v)
		}
	}
	return list
}

func (s *Storage) getTodoProject(index int) string {
	t := s.pendingTodos[index]
	return t.projectName
}
