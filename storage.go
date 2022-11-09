package main

import (
	"context"
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
	return context.WithTimeout(context.Background(), time.Second*5)
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

func (s Storage) quickAdd(content string) ([]Todo, error) {
	ctx, cancel := newContext()
	defer cancel()
	err := s.api.quickAdd(ctx, content)
	if err != nil {
		return nil, err
	}
	return s.fetchTodos()
}
