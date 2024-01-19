package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
)

type API struct {
	token  string
	client http.Client
}

func NewAPI(tokenPath string) (API, error) {
	t, err := os.ReadFile(tokenPath)
	if err != nil {
		return API{}, err
	}

	return API{
		token: strings.TrimSpace(fmt.Sprintf("%s", t)),
	}, nil
}

type QuickAddResponse struct {
	Id string `json:"id"`
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

// returns the id of the new todo
func (api API) quickAdd(ctx context.Context, content string) (string, error) {
	values := url.Values{
		"text": {content},
	}
	s := values.Encode()
	body := strings.NewReader(s)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.todoist.com/sync/v9/quick/add", body)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+api.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("quickadd returned %d status code", res.StatusCode)
	}
	var quickAddResponse QuickAddResponse
	err = json.NewDecoder(res.Body).Decode(&quickAddResponse)
	if err != nil {
		return "", err
	}
	return quickAddResponse.Id, nil
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

func (api API) makeChild(ctx context.Context, parentId, childId string) error {
	v := fmt.Sprintf(`[
    {
        "type": "item_move",
        "uuid": "%s",
        "args": {
            "id": "%s", 
            "parent_id": "%s"
        }
    }]`, uuid.New().String(), childId, parentId)
	values := url.Values{
		"commands": {v},
	}
	s := values.Encode()
	body := strings.NewReader(s)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.todoist.com/sync/v9/sync", body)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+api.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return nil
}

// NOTE: if makechild fails we are in a wierd state...
func (api API) newChild(ctx context.Context, parentId string, content string) error {
	id, err := api.quickAdd(ctx, content)
	if err != nil {
		return err
	}
	return api.makeChild(ctx, parentId, id)
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
	DueString   string   `json:"due_string,omitempty"`
}

// TODO: somehow support editing of children
func (api API) editTask(ctx context.Context, todo Todo) error {
	url := fmt.Sprintf("https://api.todoist.com/rest/v2/tasks/%s", todo.Id)
	t := EditRequest{
		Content:     todo.Content,
		Description: todo.Description,
		Due:         todo.Due.Date,
		Labels:      todo.Labels,
		Priority:    todo.Priority,
		DueString:   todo.Due.ChangeString,
	}
	if t.DueString != "" { // If change string is set, then we dont want to set due. To avoid conflicts
		t.Due = ""
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
