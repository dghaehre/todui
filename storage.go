package main

import (
	"context"
	"time"
)

// conventions:
// fetch is used for api use
// local is used for sqlite use

type Storage struct {
	api API
	db  DB
}

// TODO: use sync token and populate database
func (s Storage) fetchTodos() ([]Todo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	res, err := s.api.getPendingTodos(ctx)
	if err != nil {
		return nil, err
	}
	return toTodos(res.Items, res.Projects), nil
}

func (s Storage) localTodos() ([]Todo, error) {
	var todos = make([]Todo, 0, 0)
	return todos, nil
}
