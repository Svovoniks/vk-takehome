package event_db

import (
	"backend/config"
	"backend/ping"
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type DB struct {
	Db    *sql.DB
	Table string
}

func GetDB(cfg *config.Config) (*DB, error) {
	conntectStr := fmt.Sprintf(
		"user='%s' password='%s' host='%s' port='%s' sslmode=disable dbname=%s",
		cfg.DbUser,
		cfg.DbPassword,
		cfg.DbHost,
		cfg.DbPort,
		cfg.DbName,
	)

	db, err := sql.Open("postgres", conntectStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)

	db.SetMaxIdleConns(10)

	return &DB{
		Db:    db,
		Table: "\"main.ping_event\"",
	}, nil
}

func (d *DB) PutBulk(pingData ping.PingEventList) error {
	valueStrings := make([]string, 0, len(pingData))
	valueArgs := make([]interface{}, 0, len(pingData)*3)

	buff := bytes.Buffer{}
	c := 1
	for _, post := range pingData {
		buff.WriteString("($")
		valueArgs = append(valueArgs, post.Ip)
		buff.WriteString(strconv.Itoa(c))
		c++

		buff.WriteString(",$")
		valueArgs = append(valueArgs, post.Ping_ms)
		buff.WriteString(strconv.Itoa(c))
		c++

		buff.WriteString(",$")
		valueArgs = append(valueArgs, post.Pinged_at)
		buff.WriteString(strconv.Itoa(c))
		c++

		buff.WriteRune(')')

		valueStrings = append(valueStrings, buff.String())
		buff.Reset()
	}
	stmt := fmt.Sprintf("INSERT INTO %s(ip, ping_ms, pinged_at) VALUES %s ON CONFLICT (ip) DO UPDATE SET ping_ms=excluded.ping_ms, pinged_at=excluded.pinged_at",
		d.Table, strings.Join(valueStrings, ", "))
	_, err := d.Db.Exec(stmt, valueArgs...)

	return err
}

func (d *DB) GetAll() ([]ping.PingEvent, error) {
	query := fmt.Sprintf("SELECT * FROM %s", d.Table)
	rows, err := d.Db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var events []ping.PingEvent

	for rows.Next() {
		event := ping.PingEvent{}

		if err := rows.Scan(&event.Ip, &event.Ping_ms, &event.Pinged_at); err != nil {
			log.Printf("%v\n %v\n", "Couldn't scan row", err)
			continue
		}

		events = append(events, event)
	}

	return events, nil
}
