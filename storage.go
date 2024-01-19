package main

import (
	"context"
	"fmt"
	"time"
)

type Storage struct {
	api API
	db  DB
}

type PendingResponse struct {
	Projects []Project `json:"projects"`
	Items    []Item    `json:"items"`
}

type SyncResponse struct {
	Projects  []Project `json:"projects"`
	Items     []Item    `json:"items"`
	SyncToken string    `json:"sync_token"`
}

func newContext() (context.Context, func()) {
	return context.WithTimeout(context.Background(), time.Second*8)
}

// NOTE: kinda ugly now, but works ish.
// fetches from api with sync token, and then does a "refetch" from db to get all relevant todos.
// This is to not have to map over existing todos, but just update the new list with everything.
func (s Storage) fetchTodos() ([]Todo, error) {
	ctx, cancel := newContext()
	defer cancel()
	token, err := s.db.getToken(ctx)
	if err != nil {
		return nil, err
	}
	res, err := s.api.getPending(ctx, token)
	if err != nil {
		return nil, err
	}
	err = s.db.InsertFromSync(ctx, res)
	if err != nil {
		return nil, err
	}
	localRes, err := s.db.getPending(ctx)
	if err != nil {
		return nil, err
	}
	return toTodos(localRes.Items, localRes.Projects), nil
}

func (s Storage) localTodos() ([]Todo, error) {
	ctx, cancel := newContext()
	defer cancel()
	res, err := s.db.getPending(ctx)
	if err != nil {
		return nil, err
	}
	return toTodos(res.Items, res.Projects), nil
}

func (s Storage) newTask(todo Todo) ([]Todo, error) {
	ctx, cancel := newContext()
	defer cancel()
	err := s.api.newTask(ctx, todo)
	if err != nil {
		return nil, err
	}
	return s.fetchTodos()
}

func (s Storage) editTask(data EditTaskData) ([]Todo, error) {
	ctx, cancel := newContext()
	defer cancel()
	err := s.api.editTask(ctx, data.todo)
	if err != nil {
		return nil, err
	}
	// NOTE: could possibly do this in parallel.
	for _, child := range data.updateChildren {
		switch child.UpdateStatus {
		case UpdateStatusModified:
			err = s.api.editTask(ctx, child.Org)
			if child.Checked {
				err = s.api.markAsDone(ctx, child.Org)
			}
		case UpdateStatusDeleted:
			// err = s.api.deleteTask(ctx, child.Org)
			return nil, fmt.Errorf("delete is not implemented")
		case UpdateStatusNew:
			err = s.api.newChild(ctx, data.todo.Id, child.Content)
		}
	}
	return s.fetchTodos()
}

func (s Storage) quickAdd(content string) ([]Todo, error) {
	ctx, cancel := newContext()
	defer cancel()
	_, err := s.api.quickAdd(ctx, content)
	if err != nil {
		return nil, err
	}
	return s.fetchTodos()
}

func (s Storage) markAsDone(todo Todo) ([]Todo, error) {
	ctx, cancel := newContext()
	defer cancel()
	err := s.api.markAsDone(ctx, todo)
	if err != nil {
		return nil, err
	}
	return s.fetchTodos()
}
