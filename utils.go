package main

import (
	"errors"
	"os"
	"strings"
)

func Contains[T comparable](list []T, x T) bool {
	for _, v := range list {
		if v == x {
			return true
		}
	}
	return false
}

func creatFileIfNotExist(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return os.WriteFile(path, []byte(""), 0644)
		}
	}
	return err
}

type Filter struct {
	content     []string
	projects    []string
	labels      []string
	notProjects []string
	notLabels   []string
}

func parseFilter(s string) (Filter, bool) {
	var f Filter
	empty := true
	for _, w := range strings.Fields(s) {
		empty = false
		if strings.Index(w, "#") == 0 {
			f.projects = append(f.projects, strings.TrimPrefix(w, "#"))
			continue
		}
		if strings.Index(w, "@") == 0 {
			f.labels = append(f.labels, strings.TrimPrefix(w, "@"))
			continue
		}
		f.content = append(f.content, w)
	}
	return f, empty
}

// THE filter function
// TODO: might need some recursion here as well
func filterContents(list []Todo, filter string) []Todo {
	f, emptyFilter := parseFilter(filter)
	if emptyFilter {
		return list
	}
	var newList = make([]Todo, 0, len(list))
	for _, t := range list {
		for _, project := range f.projects {
			if t.ProjectName == project {
				newList = append(newList, t)
				continue
			}
		}
		for _, label := range f.labels {
			if Contains(t.Labels, label) {
				newList = append(newList, t)
				continue
			}
		}
		for _, w := range f.content {
			if strings.Contains(strings.ToLower(t.Content), strings.ToLower(w)) {
				newList = append(newList, t)
				continue
			}
		}
	}
	return newList
}

func filterToday(list []Todo) []Todo {
	var newList = make([]Todo, 0)
	for _, t := range list {
		if t.DueTodayOrBefore() {
			newList = append(newList, t)
		}
	}
	return newList
}

func filterInbox(list []Todo) []Todo {
	var newList = make([]Todo, 0)
	for _, t := range list {
		if t.ProjectName == "Inbox" {
			newList = append(newList, t)
		}
	}
	return newList
}

// Get length of the longest project name in the list
func projectNameSize(list []Todo, max int) int {
	res := 0
	for _, v := range list {
		l := len(v.ProjectName)
		if l > res {
			if l >= max {
				return max
			}
			res = l
		}
	}
	return res
}

func withSize(s string, i int) string {
	ss := strings.TrimRight(s, " ")
	if len(ss) <= (i - 3) {
		return s
	}
	total := s[:(i - 3)]
	return total + "..."
}

func getProjectName(projects []Project, id string) string {
	for _, p := range projects {
		if p.Id == id {
			return p.Name
		}
	}
	return ""
}

func toTodo(item Item, projects []Project) Todo {
	return Todo{
		Id:          item.Id,
		ProjectName: getProjectName(projects, item.ProjectId),
		ProjectId:   item.ProjectId,
		Content:     item.Content,
		Description: item.Description,
		Priority:    item.Priority,
		Labels:      item.Labels,
		Checked:     item.Checked,
		Due:         item.Due,
		Children:    []Todo{},
	}
}

// TODO: Support children of children
// This simple implementation only supports one level of children
// All other children will be lost..
func toTodos(items []Item, projects []Project) []Todo {
	todos := make([]Todo, 0)
	children := make([]Item, 0)
	// Create all root Todos
	for _, item := range items {
		if item.ParentId == "" {
			todos = append(todos, toTodo(item, projects))
		} else {
			children = append(children, item)
		}
	}

	for _, c := range children {
		for i := 0; i < len(todos); i++ {
			if todos[i].Id == c.ParentId {
				todos[i].Children = append(todos[i].Children, toTodo(c, projects))
			}
		}
	}
	return todos
}

func displayPrioriy(p int) string {
	switch p {
	case 4:
		return p1Style.Render("p1")
	case 3:
		return p2Style.Render("p2")
	case 2:
		return p3Style.Render("p3")
	default:
		return "  "
	}
}
