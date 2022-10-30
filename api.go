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

type SyncResponse struct {
	Projects []Project `json:"projects"`
	Items    []Item    `json:"items"`
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

func (api API) getPendingTodos(ctx context.Context) (SyncResponse, error) {
	var syncResponse SyncResponse
	values := url.Values{
		"sync_token":     {"*"},
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
