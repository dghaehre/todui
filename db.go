package main

import (
	"context"
	"database/sql"
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
 checked bit
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
	query := `replace into item (id, project_id, content, description, priority, parent_id, checked) values (@id, @projectid, @content, @description, @priority, @parentid, @checked)`
	for _, item := range items {
		_, err := tx.ExecContext(ctx, query,
			sql.Named("id", item.Id),
			sql.Named("projectid", item.ProjectId), // TODO: handle null!
			sql.Named("content", item.Content),
			sql.Named("description", item.Description),
			sql.Named("priority", item.Priority),
			sql.Named("parentid", item.ParentId),
			sql.Named("checked", item.Checked),
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
	query := `select id, project_id, content, description, priority, parent_id from item where checked = false`
	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return items, err
	}
	for rows.Next() {
		var item Item
		err = rows.Scan(&item.Id, &item.ProjectId, &item.Content, &item.Description, &item.Priority, &item.ParentId)
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
