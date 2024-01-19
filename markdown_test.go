package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarkdownSimpleParserEdit(t *testing.T) {
	todo := Todo{
		Content:     "Thish is a test todo",
		Description: "some description",
		Priority:    1,
		Labels:      []string{"test"},
		Checked:     false,
		Children:    []Todo{},
		Due: Due{
			Date: "2021-01-01",
		},
	}

	file, err := createEditFile(todo)
	require.NoError(t, err)

	// Do no changes

	b, err := os.ReadFile(file)
	require.NoError(t, err)
	t.Log(string(b))

	parsedTodo, _, err := parseEditFile(file, todo)
	require.NoError(t, err)
	require.Equal(t, todo, parsedTodo)

	parsed, _, err := parseEditFile(file, todo)
	require.NoError(t, err)
	require.Equal(t, todo, parsed)
}

func TestMarkdownEmptyParserEdit(t *testing.T) {
	// Test with empty todo
	todo2 := Todo{
		Content:  "Thish is a test todo",
		Labels:   []string{},
		Children: []Todo{},
	}

	file2, err := createEditFile(todo2)
	require.NoError(t, err)

	// Do no changes

	b, err := os.ReadFile(file2)
	require.NoError(t, err)
	t.Log(string(b))

	parsedTodo2, _, err := parseEditFile(file2, todo2)
	require.NoError(t, err)
	require.Equal(t, todo2, parsedTodo2)
}

func TestMarkdownWithChildren(t *testing.T) {
	withoutDesc := Todo{
		Content:     "Thish is a test todo",
		Description: "",
		Priority:    1,
		Labels:      []string{"test"},
		Checked:     false,
		Children: []Todo{
			{
				Id:      "123",
				Content: "child 1",
				Checked: true,
			},
			{
				Id:      "456",
				Content: "child 2",
				Checked: false,
			},
		},
	}
	file, err := createEditFile(withoutDesc)
	require.NoError(t, err)

	// Do no changes

	parsedTodo, _, err := parseEditFile(file, withoutDesc)
	require.NoError(t, err)
	require.Equal(t, withoutDesc, parsedTodo)
	require.Equal(t, len(withoutDesc.Children), len(parsedTodo.Children))
	require.Equal(t, withoutDesc.Description, parsedTodo.Description)
	require.Equal(t, withoutDesc.Children[0].Content, parsedTodo.Children[0].Content)
	// t.Log(parsedTodo.Children[0].Content)
	// t.Fail()

	/////////////////////////////////

	withDesc := Todo{
		Content:     "Thish is a test todo",
		Description: "",
		Priority:    1,
		Labels:      []string{"test"},
		Checked:     false,
		Children: []Todo{
			{
				Content: "child 1",
				Checked: true,
			},
			{
				Content: "child 2",
				Checked: false,
			},
		},
	}
	file2, err := createEditFile(withDesc)
	require.NoError(t, err)

	// Do no changes

	parsedTodo2, _, err := parseEditFile(file2, withDesc)
	require.NoError(t, err)
	require.Equal(t, withDesc, parsedTodo2)
	require.Equal(t, len(withDesc.Children), len(parsedTodo2.Children))
	require.Equal(t, withDesc.Description, parsedTodo2.Description)
	require.Equal(t, withDesc.Children[0].Content, parsedTodo2.Children[0].Content)
}

func TestMarkdownChangeChild(t *testing.T) {
	todo := Todo{
		Content:     "Thish is a test todo",
		Description: "some text",
		Priority:    1,
		Labels:      []string{"test"},
		Checked:     false,
		Children: []Todo{
			{
				Id:      "123",
				Content: "child 1",
				Checked: true,
			},
		},
	}
	path, err := createEditFile(todo)
	require.NoError(t, err)

	b, err := os.ReadFile(path)
	require.NoError(t, err)

	modifiedContent := strings.Replace(string(b), "[ ] child 1", "[ ] something else", 1)
	err = os.WriteFile(path, []byte(modifiedContent), 0644)
	require.NoError(t, err)

	parsedTodo, updatedChildren, err := parseEditFile(path, todo)
	require.NoError(t, err)
	require.Equal(t, len(todo.Children), len(parsedTodo.Children))
	require.Equal(t, "something else", parsedTodo.Children[0].Content)
	require.Equal(t, 1, len(updatedChildren))
	require.Equal(t, "something else", updatedChildren[0].Content)
	require.Equal(t, "123", updatedChildren[0].Org.Id)
}

func TestMarkdownNewChild(t *testing.T) {
	todo := Todo{
		Content:     "Thish is a test todo",
		Description: "some text",
		Priority:    1,
		Labels:      []string{"test"},
		Checked:     false,
		Children: []Todo{
			{
				Content: "child 1",
				Checked: false,
			},
		},
	}
	path, err := createEditFile(todo)
	require.NoError(t, err)

	b, err := os.ReadFile(path)
	require.NoError(t, err)

	modifiedContent := string(b) + "- [ ] child 2\n"
	err = os.WriteFile(path, []byte(modifiedContent), 0644)
	require.NoError(t, err)

	parsedTodo, updatedChildren, err := parseEditFile(path, todo)
	require.NoError(t, err)
	require.Equal(t, len(todo.Children), len(parsedTodo.Children))
	require.Equal(t, 1, len(updatedChildren))
	require.Equal(t, UpdateStatusNew, updatedChildren[0].UpdateStatus)
	require.Equal(t, "child 2", updatedChildren[0].Content)
	require.False(t, updatedChildren[0].Checked)
}
