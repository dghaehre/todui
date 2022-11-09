package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
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
			return ioutil.WriteFile(path, []byte(""), 0644)
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
		Children:    []Todo{},
	}
}

// tries to set item onto child, returns ok if succesfull
func setChild(projects []Project, todo *Todo, item Item) bool {
	if todo.Id == item.ParentId {
		todo.Children = append(todo.Children, toTodo(item, projects))
		return true
	}
	for i := 0; i < len(todo.Children); i++ {
		if setChild(projects, &todo.Children[i], item) {
			return true
		}
	}
	return false
}

func toTodos(items []Item, projects []Project) []Todo {
	var todos = make([]Todo, 0, len(items))
	var children = make(map[string]Item)
	// Create all root nodes
	for _, item := range items {
		if item.ParentId == "" {
			todos = append(todos, toTodo(item, projects))
		} else {
			children[item.ProjectId] = item
		}
	}

	for _, c := range children {
		for i := 0; i < len(todos); i++ {
			ok := setChild(projects, &todos[i], c)
			if ok {
				delete(children, c.Id)
			}
		}
	}
	// What do we do if we have children left over..?
	return todos
}

// in markdown
func createEditFile(m model) (string, error) {
	todo := m.filteredTodos[m.cursor.index]
	path := fmt.Sprintf("/tmp/%s.md", todo.Id)
	var b bytes.Buffer
	fmt.Fprintf(&b, "# %s", todo.Content)
	fmt.Fprint(&b, "\n\n")
	fmt.Fprintf(&b, "%s", todo.Description)
	err := ioutil.WriteFile(path, b.Bytes(), 0644)
	return path, err
}

func createNewTaskFile() (string, error) {
	path := fmt.Sprintf("/tmp/%d.md", time.Now().Unix())
	var b bytes.Buffer
	fmt.Fprintf(&b, "---")
  fmt.Fprintf(&b, "labels: ")
  fmt.Fprintf(&b, "project: Inbox")
	fmt.Fprintf(&b, "---")
	fmt.Fprintf(&b, "# ")
	err := ioutil.WriteFile(path, b.Bytes(), 0644)
	return path, err
}

// TODO
func parseTaskFile(path string) (todo Todo, err error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return todo, err
	}
	lines := strings.SplitN(string(b), "\n", 2)
	todo.Content = strings.TrimPrefix(lines[0], "#")
	if len(todo.Content) < 2 {
		return todo, fmt.Errorf("parse error: could not parse header")
	}
	if len(lines) == 2 {
		todo.Description = lines[1]
	}
	return todo, nil
}
