package Db

import (
	"context"
	"database/sql"
	"log"
)

type DbAccess struct {
	Db             *sql.DB
	PgsqlNameServe string
}

func (Db *DbAccess) Connect() error {
	var err error
	Db.Db, err = sql.Open("pgx", Db.PgsqlNameServe)
	if err != nil {
		return err
	}
	return err
}

func (Db *DbAccess) Close() {
	err := Db.Db.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func (Db *DbAccess) Check(ctx context.Context, id string) string {
	rows, err := Db.Db.QueryContext(ctx, "SELECT refresh FROM token WHERE id = $1", id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var token string
	if err = rows.Scan(&token); err != nil {
		log.Fatal(err)
	}

	return token
}

func (Db *DbAccess) Add(ctx context.Context, guid, refresh string, id int64) {
	if _, err := Db.Db.ExecContext(ctx, "INSERT INTO token (guid, refresh, id) VALUES ($1,$2,$3,$4)", guid, refresh, id); err != nil {
	}
}

func (Db *DbAccess) Refresh(ctx context.Context, oldHashToken, newHashToken string) {
	if _, err := Db.Db.ExecContext(ctx, "UPDATE token SET refresh = $1 WHERE refresh = $2", newHashToken, oldHashToken); err != nil {
		log.Fatal(err)
	}
}

func (Db *DbAccess) GetEmail(ctx context.Context, guid string) string {
	row, err := Db.Db.QueryContext(ctx, "SELECT email FROM users WHERE guid = $1", guid)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	var email string
	row.Scan(&email)

	return email
}
