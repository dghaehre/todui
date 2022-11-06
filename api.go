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
