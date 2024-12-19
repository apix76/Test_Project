package db

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
)

type DbAccess struct {
	db *sql.DB
}

func New(dsn string) (DbAccess, error) {
	db := DbAccess{}

	var err error
	db.db, err = sql.Open("pgx", dsn)
	return db, err
}

func (Db *DbAccess) Close() {
	err := Db.db.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func (Db *DbAccess) Check(id string) (string, error) {
	row := Db.db.QueryRow("SELECT refresh FROM token WHERE id = $1", id)

	var token string
	if err := row.Scan(&token); err != nil {
		if err != sql.ErrNoRows {
			return "", err
		}
	}

	return token, nil
}

func (Db *DbAccess) Add(guid, refresh string, id string) error {
	if _, err := Db.db.Exec("INSERT INTO token (guid, refresh, id) VALUES ($1,$2,$3)", guid, refresh, id); err != nil {
		return err
	}
	return nil
}

func (Db *DbAccess) Refresh(oldHashToken, newHashToken string) error {
	_, err := Db.db.Exec("UPDATE token SET refresh = $1 WHERE refresh = $2", newHashToken, oldHashToken)
	return err
}

func (Db *DbAccess) GetEmail(guid string) string {
	row := Db.db.QueryRow("SELECT email FROM users WHERE guid = $1", guid)

	var email string
	if err := row.Scan(&email); err != nil {
		if err != sql.ErrNoRows {
			return ""
		}
	}

	return email
}

func (Db *DbAccess) Delete(id string) error {
	_, err := Db.db.Exec("DELETE FROM token WHERE id = $1", id)
	return err
}

//func (Db *DbAccess) GetSession(id, ip string, data time.Time) {
//	if id != "" {
//		id := fmt.Sprintf("id = %v", id)
//	}
//	if ip != "" {
//		ip := fmt.Sprintf("ip = %v", id)
//	}
//	if ok := data.IsZero(); !ok {
//
//	}
//	req := fmt.Sprintf("%v, %v, %v", id, ip, data)
//	var rows *sql.Rows
//	if req != "" {
//		rows, err := Db.db.Query("SELECT guid, id, data FROM token WHERE $1", req)
//	} else {
//		rows, err := Db.db.Query("SELECT guid, id, data FROM token ")
//	}
//
//}
