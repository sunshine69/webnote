package models

import (
	"os"
	"log"
	"database/sql"
)

func Migrate() {
	os.Setenv("DBPATH", "/home/stevek/src/webnote-go/testwebnote.db")
	sqlite3File := "/home/stevek/webnote.sqlite3"
	DB, e := sql.Open("sqlite3", sqlite3File)
	if e != nil {
		log.Fatalf("ERROR opening sqlite")
	}
	defer DB.Close()
	q := `SELECT
		title,
		datelog,
		content,
		url,
		reminder_ticks,
		flags,
		timestamp,
		time_spent,
		permission,
		raw_editor
	FROM webnote_note
	`
	q = `SELECT
		title,
		content,
		url,
		flags,
		timestamp,
		permission,
		raw_editor
	FROM webnote_note
	`
	rows, e := DB.Query(q)
	if e != nil {
		log.Fatalf("ERROR run Query %v\n", e)
	}
	for rows.Next() {
		aNote := Note{}
		aNote.AuthorID = int64(1)
		aNote.GroupID = int8(1)
		var nContent, nURL, nFlags sql.NullString
		// var nReminder, nTimeSpent sql.NullInt64
		if err := rows.Scan(
			&aNote.Title, &nContent, &nURL,
			&nFlags, &aNote.Timestamp, &aNote.Permission, &aNote.RawEditor,
		); err != nil {
			log.Fatalf("ERROR Scan result %v\n", err)
		}
		if nContent.Valid {	aNote.Content = nContent.String	}
		if nURL.Valid { aNote.URL = nURL.String }
		if nFlags.Valid { aNote.Flags = nFlags.String }
		// if nReminder.Valid { aNote.ReminderTicks = nReminder.Int64 }
		// if nTimeSpent.Valid { aNote.TimeSpent = nTimeSpent.Int64 }
		aNote.Save()
	}
}