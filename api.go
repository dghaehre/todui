package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type API struct {
	token  string
	client http.Client
}

func NewAPI(tokenPath string) (API, error) {
	t, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return API{}, err
	}

	return API{
		token: strings.TrimSpace(fmt.Sprintf("%s", t)),
	}, nil
}

func (api API) getPending(ctx context.Context, token string) (SyncResponse, error) {
	var syncResponse SyncResponse
	values := url.Values{
		"sync_token":     {token},
		"resource_types": {"[\"items\", \"projects\"]"},
	}
	s := values.Encode()
	body := strings.NewReader(s)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.todoist.com/sync/v9/sync", body)
	if err != nil {
		return syncResponse, err
	}
	req.Header.Add("Authorization", "Bearer "+api.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return syncResponse, err
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&syncResponse)
	return syncResponse, err
}

func (api API) quickAdd(ctx context.Context, content string) error {
	values := url.Values{
		"text": {content},
	}
	s := values.Encode()
	body := strings.NewReader(s)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.todoist.com/sync/v9/quick/add", body)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+api.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("quickadd returned %d status code", res.StatusCode)
	}
	return nil
}

func (api API) markAsDone(ctx context.Context, todo Todo) error {
	url := fmt.Sprintf("https://api.todoist.com/rest/v2/tasks/%s/close", todo.Id)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(""))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+api.token)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 204 {
		return fmt.Errorf("Could not mark task as done. API returned %d status code, expected 204", res.StatusCode)
	}
	return nil
}

func (api API) newTask(ctx context.Context, todo Todo) error {
	return fmt.Errorf("new task: not implemented yet")
}

// TODO: use sync api
// NOT usign the sync API
//
// Currently only sending content and description
type EditRequest struct {
	Content     string   `json:"content,omitempty"`
	Description string   `json:"description,omitempty"`
	Due         string   `json:"due_date,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Priority    int      `json:"priority,omitempty"`
}

func (api API) editTask(ctx context.Context, todo Todo) error {
	url := fmt.Sprintf("https://api.todoist.com/rest/v2/tasks/%s", todo.Id)
	t := EditRequest{
		Content:     todo.Content,
		Description: todo.Description,
		Due:         todo.Due.Date,
		Labels:      todo.Labels,
	}
	body, err := json.Marshal(t)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+api.token)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("Could not edit task. API returned %d status code", res.StatusCode)
	}
	return nil
}
