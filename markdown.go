package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// https://github.com/yuin/goldmark-meta

var markdown = goldmark.New(
	goldmark.WithExtensions(
		meta.New(
			meta.WithStoresInDocument(),
		),
	),
)

func createNewTaskFile() (string, error) {
	path := fmt.Sprintf("/tmp/%d.md", time.Now().Unix())
	var b bytes.Buffer
	fmt.Fprintf(&b, "---")
	fmt.Fprintf(&b, "labels: ")
	fmt.Fprintf(&b, "project: Inbox")
	fmt.Fprintf(&b, "---")
	fmt.Fprintf(&b, "# ")
	err := os.WriteFile(path, b.Bytes(), 0644)
	return path, err
}

func renderPriority(p int) string {
	switch p {
	case 1:
		return "p4"
	case 2:
		return "p3"
	case 3:
		return "p2"
	case 4:
		return "p1"
	}
	return ""
}

func parsePriority(p string) int {
	switch p {
	case "p4":
		return 1
	case "p3":
		return 2
	case "p2":
		return 3
	case "p1":
		return 4
	}
	return 1
}

// TODO: some validation might be needed here..
func parseEditFile(path string, todo Todo) (Todo, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return todo, err
	}

	document := markdown.Parser().Parse(text.NewReader(b))
	metaData := document.OwnerDocument().Meta()
	document.OwnerDocument().Dump(b, 0)

	switch v := metaData["due"].(type) {
	case string:
		todo.Due.Date = v
	}

	switch v := metaData["priority"].(type) {
	case string:
		todo.Priority = parsePriority(v)
	}

	labels := make([]string, 0)
	switch v := metaData["labels"].(type) {
	case []interface{}:
		for _, l := range v {
			labels = append(labels, l.(string))
		}
	}
	todo.Labels = labels

	todo.Content = string(document.FirstChild().Text(b))
	switch node := document.FirstChild().NextSibling().(type) {
	case ast.Node:
		todo.Description = strings.TrimSpace(string(node.Text(b)))
	}

	return todo, nil
}

func createEditFile(todo Todo) (string, error) {
	path := fmt.Sprintf("/tmp/%s.md", todo.Id)
	var b bytes.Buffer
	fmt.Fprintf(&b, "---\n")
	fmt.Fprintf(&b, "due: %s\n", todo.Due.Date)
	fmt.Fprintf(&b, "priority: %s\n", renderPriority(todo.Priority))
	fmt.Fprintf(&b, "labels:\n")
	for _, t := range todo.Labels {
		fmt.Fprintf(&b, " - %s\n", t)
	}
	fmt.Fprintf(&b, "---\n\n")
	fmt.Fprintf(&b, "# %s\n\n", todo.Content)
	fmt.Fprintf(&b, "%s", todo.Description)
	err := os.WriteFile(path, b.Bytes(), 0644)
	return path, err
}

// TODO
func parseTaskFile(path string) (todo Todo, err error) {
	b, err := os.ReadFile(path)
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
