package main

import (
	// "io/ioutil"
	// "io/fs"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	mockTodos []Todo = []Todo{
		{Id: "0", Content: "some test", ProjectName: "test"},
		{Id: "1", Content: "dog"},
		{Id: "2", Content: "enda en test", ProjectName: "chores"},
	}
)

func TestFilterContents(t *testing.T) {
	res := filterContents(mockTodos, "")
	if len(res) != len(mockTodos) {
		t.Errorf("filtered length should be %d", len(mockTodos))
	}

	if !strings.Contains("dette er en test", "test") {
		t.Errorf("hva faen")
	}

	if strings.Contains("dette er en dog", "test") {
		t.Errorf("hva faen")
	}

	res = filterContents(mockTodos, "test")
	if len(res) != 2 {
		t.Errorf("filtered length should be %d, not %d", 2, len(res))
	}

	res = filterContents(mockTodos, "enda")
	if len(res) != 1 {
		t.Errorf("filtered length should be %d, not %d", 2, len(res))
	}
}

func TestToTodos(t *testing.T) {
	items := []Item{{
		Id:          "1",
		ProjectId:   "1",
		Content:     "test",
		Description: "first test",
		Priority:    0,
		ParentId:    "",
		Checked:     false,
	}}

	projects := []Project{{
		Id:   "1",
		Name: "Inbox",
	}}

	todos := toTodos(items, projects)
	require.Equal(t, 1, len(todos))
}

func TestToTodosWithChildren(t *testing.T) {
	items := []Item{{
		Id:          "1",
		ProjectId:   "1",
		Content:     "parent item",
		Description: "",
		ParentId:    "",
	}, {
		Id:          "2",
		ProjectId:   "1",
		Content:     "first child",
		Description: "",
		ParentId:    "1",
	}, {
		Id:          "3",
		ProjectId:   "1",
		Content:     "second child",
		Description: "",
		ParentId:    "1",
	}}

	projects := []Project{{
		Id:   "1",
		Name: "Inbox",
	}}

	todos := toTodos(items, projects)
	require.Equal(t, 1, len(todos))
	require.Equal(t, 2, len(todos[0].Children))
}

// TODO
func TestParseTaskFile(t *testing.T) {
	// Testing empty task
	// path, err := createNewTaskFile()
	// require.NoError(t, err)
	// _, err = parseTaskFile(path)
	// require.Equal(t, "parse error: could not parse header", err.Error())
	// err = os.Remove(path)
	// require.NoError(t, err)

	// // Test oneliner
	// path, err = createNewTaskFile()
	// require.NoError(t, err)
	// ioutil.WriteFile(path, []byte("this is a test"), fs.ModePerm)
	// todo, err := parseTaskFile(path)
	// require.NoError(t, err)
	// require.Equal(t, "this is a test", todo.Content)
	// require.Equal(t, "", todo.Description)
	// err = os.Remove(path)
	// require.NoError(t, err)

	// Test multiline
	// path, err = createNewTaskFile()
	// require.NoError(t, err)
	// ioutil.WriteFile(path, []byte("this is a test\ndescription\nhei"), fs.ModePerm)
	// todo, err = parseTaskFile(path)
	// require.NoError(t, err)
	// require.Equal(t, "this is a test", todo.Content)
	// require.Equal(t, "description\nhei", todo.Description)
	// err = os.Remove(path)
	// require.NoError(t, err)
}

func TestEditAndParseTaskFile(t *testing.T) {
	todo := Todo{
		Id:          "123",
		ProjectName: "chores",
		ProjectId:   "456",
		Content:     "A test todo",
		Description: "from a test",
		Labels:      []string{},
		Children:    []Todo{},
	}

	path, err := createEditFile(todo)
	require.NoError(t, err)
	newTodo, _, err := parseEditFile(path, todo)
	require.NoError(t, err)
	require.EqualValues(t, todo, newTodo)

	err = os.Remove(path)
	require.NoError(t, err)
}

func TestProjectNameSize(t *testing.T) {
	todos := []Todo{
		{Id: "0", Content: "some test", ProjectName: "test"},
		{Id: "1", Content: "dog"},
		{Id: "2", Content: "enda en test", ProjectName: "chores"},
	}
	i := projectNameSize(todos, 30)
	require.Equal(t, 6, i)
}
