package main

import (
	"context"
	"database/sql"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func NewDB(path string) (DB, error) {
	err := creatFileIfNotExist(path)
	if err != nil {
		return DB{}, err
	}
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return DB{}, err
	}
	err = conn.Ping()
	if err != nil {
		return DB{}, err
	}
	db := DB{
		conn: conn,
	}
	err = db.setup()
	return db, err
}

func (db DB) Close() error {
	return db.conn.Close()
}

// TODO: Handle migrations etc..
func (db DB) setup() error {
	_, err := db.conn.Exec(`
create table if not exists synctoken (
 id integer primary key,
 token text not null
);

create table if not exists completed (
 id integer primary key,
 content text,
 project_id text,
 completed_at text
);

create table if not exists project (
 id integer primary key,
 is_archived bit,
 is_deleted bit,
 name text not null,
 parent_id integer
);

create table if not exists item (
 id integer primary key,
 project_id integer,
 content text,
 description text,
 priority integer,
 parent_id integer,
 checked bit,
 labels text,
 due_is_recurring bit,
 due_date text,
 due_string text,
 due_timezone text,
 due_lang text
)`)
	return err
}

func (db DB) InsertFromSync(ctx context.Context, res SyncResponse) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	query := `replace into synctoken (id, token) values (@id, @token)`
	_, err = tx.ExecContext(ctx, query, sql.Named("id", 0), sql.Named("token", res.SyncToken))
	if err != nil {
		tx.Rollback()
		return err
	}
	err = insertItems(ctx, tx, res.Items)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = insertProjects(ctx, tx, res.Projects)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func insertItems(ctx context.Context, tx *sql.Tx, items []Item) error {
	query := `replace into item (id, project_id, content, description, priority, parent_id, checked, due_is_recurring, due_date, due_string, due_timezone, due_lang, labels) values (@id, @projectid, @content, @description, @priority, @parentid, @checked, @due_is_recurring, @due_date, @due_string, @due_timezone, @due_lang, @labels)`
	for _, item := range items {
		_, err := tx.ExecContext(ctx, query,
			sql.Named("id", item.Id),
			sql.Named("projectid", item.ProjectId), // TODO: handle null!
			sql.Named("content", item.Content),
			sql.Named("description", item.Description),
			sql.Named("priority", item.Priority),
			sql.Named("parentid", item.ParentId),
			sql.Named("checked", item.Checked),
			sql.Named("due_is_recurring", item.Due.IsRecurring),
			sql.Named("due_date", item.Due.Date),
			sql.Named("due_string", item.Due.String),
			sql.Named("due_timezone", item.Due.Timezone),
			sql.Named("due_lang", item.Due.Lang),
			sql.Named("labels", strings.Join(item.Labels, ",")),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertProjects(ctx context.Context, tx *sql.Tx, projects []Project) error {
	query := `replace into project (id, name) values (@id, @name)`
	for _, project := range projects {
		_, err := tx.ExecContext(ctx, query,
			sql.Named("id", project.Id),
			sql.Named("name", project.Name),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db DB) getToken(ctx context.Context) (string, error) {
	query := `select token from synctoken where id = 0`
	token := "*"
	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return token, err
	}
	for rows.Next() {
		err = rows.Scan(&token)
		if err != nil {
			return token, err
		}
	}
	return token, nil
}

// TODO: use one single query and return: []Todo
func (db DB) getPending(ctx context.Context) (res PendingResponse, err error) {
	items, err := db.getPendingItems(ctx)
	if err != nil {
		return res, err
	}
	projects, err := db.getProjects(ctx)
	if err != nil {
		return res, err
	}
	res.Items = items
	res.Projects = projects
	return res, err
}

func (db DB) getPendingItems(ctx context.Context) ([]Item, error) {
	var items = make([]Item, 0)
	query := `select id, project_id, content, description, priority, parent_id, due_string, due_date, due_lang, due_is_recurring, due_timezone, labels from item where checked = false`
	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return items, err
	}
	for rows.Next() {
		var item Item
		var labels string
		err = rows.Scan(&item.Id,
			&item.ProjectId,
			&item.Content,
			&item.Description,
			&item.Priority,
			&item.ParentId,
			&item.Due.String,
			&item.Due.Date,
			&item.Due.Lang,
			&item.Due.IsRecurring,
			&item.Due.Timezone,
			&labels,
		)
		if err != nil {
			return items, err
		}
		for _, l := range strings.Split(labels, ",") {
			if strings.TrimSpace(l) != "" {
				item.Labels = append(item.Labels, l)
			}
		}
		items = append(items, item)
	}
	return items, nil
}

// TODO: also fetch from completed table?
func (db DB) getCompletedItems(ctx context.Context) ([]CompletedItem, error) {
	var items = make([]CompletedItem, 0)
	query := `select id, project_id, content, description item where checked = true`
	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return items, err
	}
	for rows.Next() {
		var item CompletedItem
		err = rows.Scan(&item.Id,
			&item.ProjectId,
			&item.Content,
		)
		if err != nil {
			return items, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (db DB) getProjects(ctx context.Context) ([]Project, error) {
	var projects = make([]Project, 0)
	query := `select id, name from project`
	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return projects, err
	}
	for rows.Next() {
		var p Project
		err = rows.Scan(&p.Id, &p.Name)
		if err != nil {
			return projects, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}
