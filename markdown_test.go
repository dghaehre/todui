package main

import (
	"os"
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

	parsedTodo, err := parseEditFile(file, todo)
	require.NoError(t, err)
	require.Equal(t, todo, parsedTodo)

	parsed, err := parseEditFile(file, todo)
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

	parsedTodo2, err := parseEditFile(file2, todo2)
	require.NoError(t, err)
	require.Equal(t, todo2, parsedTodo2)
}
