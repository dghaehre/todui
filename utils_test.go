package main

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

var (
	mockTodos []Todo = []Todo{
		{Id: "0", Content: "some test"},
		{Id: "1", Content: "dog"},
		{Id: "2", Content: "enda en test"},
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

	res = filterContents(mockTodos, "test")
	if len(res) != 1 {
		t.Errorf("filtered length should be %d, not %d", 1, len(res))
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
