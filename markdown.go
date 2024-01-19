package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
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
func parseEditFile(path string, todo Todo) (Todo, []UpdateChild, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return todo, nil, err
	}

	document := markdown.Parser().Parse(text.NewReader(b))
	metaData := document.OwnerDocument().Meta()

	switch v := metaData["due"].(type) {
	case string:
		todo.Due.Date = v
	}

	switch v := metaData["due_string"].(type) {
	case string:
		if v != "" {
			todo.Due.ChangeString = v
		}
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

	title := document.FirstChild()
	todo.Content = string(title.Text(b))

	var node ast.Node
	children := make([]Todo, 0)
	updateChildren := make([]UpdateChild, 0)
	for node = title.NextSibling(); node != nil; node = node.NextSibling() {
		switch n := node.(type) {
		case *ast.List:
			children, updateChildren, err = parseUpdatedChildren(todo.Children, n, b)
			if err != nil {
				return todo, nil, err
			}
		case ast.Node:
			todo.Description = strings.TrimSpace(string(n.Text(b)))
		}
	}
	todo.Children = children
	return todo, updateChildren, nil
}

// TODO: Add support for deleting children.
// Might have to do something like: - [D] child
//
// TODO: Add support for ordering children.
//
// What if I want to rearrange the order of the children?
// https://developer.todoist.com/sync/v9/#move-an-item
func parseUpdatedChildren(children []Todo, node *ast.List, source []byte) ([]Todo, []UpdateChild, error) {
	parsedChildren := make([]Todo, 0)
	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch n.Kind() {
		case ast.KindList:
			return ast.WalkContinue, nil
		case ast.KindListItem:
			if entering { // The list item we want. We dont support nested sub lists yet.
				t := strings.TrimSpace(string(n.Text(source)))
				checked := false
				t = strings.TrimPrefix(t, "[ ] ")
				if strings.HasPrefix(t, "[X]") {
					checked = true
					t = strings.TrimPrefix(t, "[X] ")
				}
				parsedChildren = append(parsedChildren, Todo{
					Content: t,
					Checked: checked,
				})
			}
			return ast.WalkContinue, nil
		}
		return ast.WalkSkipChildren, nil
	})
	if err != nil {
		return children, nil, err
	}

	if len(parsedChildren) < len(children) {
		return children, nil, fmt.Errorf("parse error: parsed children is less than original children")
	}

	// Can I figure out the order without having to add id's?
	updateChildren := make([]UpdateChild, 0)
	for i, t := range parsedChildren {
		if i >= len(children) {
			updateChildren = append(updateChildren, UpdateChild{
				Content:      t.Content,
				Checked:      t.Checked,
				UpdateStatus: UpdateStatusNew,
			})
		} else {
			if children[i].Content != t.Content || children[i].Checked != t.Checked {
				children[i].Content = t.Content
				children[i].Checked = t.Checked
				updateChildren = append(updateChildren, UpdateChild{
					Org:          children[i],
					Content:      t.Content,
					Checked:      t.Checked,
					UpdateStatus: UpdateStatusModified,
				})
			}
		}
	}
	return children, updateChildren, nil
}

func createEditFile(todo Todo) (string, error) {
	path := fmt.Sprintf("/tmp/%s.md", todo.Id)
	var b bytes.Buffer
	fmt.Fprintf(&b, "---\n")
	fmt.Fprintf(&b, "due: %s\n", todo.Due.Date)
	fmt.Fprintf(&b, "priority: %s\n", renderPriority(todo.Priority))
	fmt.Fprintf(&b, "# Use due_string to set a new date with normal language\n")
	fmt.Fprintf(&b, "due_string:\n")
	fmt.Fprintf(&b, "labels:\n")
	for _, t := range todo.Labels {
		fmt.Fprintf(&b, " - %s\n", t)
	}
	fmt.Fprintf(&b, "---\n\n")
	fmt.Fprintf(&b, "# %s\n\n", todo.Content)
	fmt.Fprintf(&b, "%s", todo.Description)
	if len(todo.Children) > 0 {
		fmt.Fprintf(&b, "\n\n")
		// NOTE: maybe put labels here? I guess projects doesnt make sense (should be the same as the parent)
		for _, t := range todo.Children {
			fmt.Fprintf(&b, "- [ ] %s\n", t.Content)
		}
	}
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
