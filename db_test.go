package main

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSetup(t *testing.T) {
	_ = newTestDB(t)
}

func TestInsert(t *testing.T) {
	db := newTestDB(t)
	syncResponse := SyncResponse{
		SyncToken: "testing",
		Items: []Item{{
			Id:          "1",
			ProjectId:   "1",
			Content:     "test",
			Description: "first test",
			Priority:    0,
			ParentId:    "",
			Checked:     false,
			Due: Due{
				IsRecurring: true,
				Date:        "",
				String:      "",
				Timezone:    "",
				Lang:        "",
			},
		}},
		Projects: []Project{{
			Id:   "1",
			Name: "Inbox",
		}},
	}
	err := db.InsertFromSync(context.Background(), syncResponse)
	require.NoError(t, err)

	res, err := db.getPending(context.Background())
	require.NoError(t, err)
	require.Equal(t, len(syncResponse.Items), len(res.Items))
	require.Equal(t, syncResponse.Items[0].Description, res.Items[0].Description)
	require.Equal(t, syncResponse.Items[0].Content, res.Items[0].Content)
	require.Equal(t, syncResponse.Items[0].Id, res.Items[0].Id)
	require.Equal(t, syncResponse.Items[0].Due.IsRecurring, res.Items[0].Due.IsRecurring)
	require.Equal(t, syncResponse.Items[0].Checked, res.Items[0].Checked)
	require.Equal(t, syncResponse.Projects[0].Name, res.Projects[0].Name)
}

func newTestDB(t *testing.T) DB {
	now := time.Now()
	dbName := fmt.Sprintf("testoutput/test-%d.db", now.Unix())
	db, err := NewDB(dbName)
	require.NoError(t, err)
	return db
}
