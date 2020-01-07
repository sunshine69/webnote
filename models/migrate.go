package models

import (
	"time"
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
		var dateLog, timeStamp sql.NullInt64
		if err := rows.Scan(
			&aNote.Title, &dateLog, &nContent, &nURL,
			&nFlags, &timeStamp, &aNote.Permission, &aNote.RawEditor,
		); err != nil {
			log.Fatalf("ERROR Scan result %v\n", err)
		}
		if nContent.Valid {	aNote.Content = nContent.String	}
		if nURL.Valid { aNote.URL = nURL.String }
		if nFlags.Valid { aNote.Flags = nFlags.String }
		if dateLog.Valid { aNote.Datelog = dateLog.Int64 * int64(time.Second / time.Nanosecond) }
		if timeStamp.Valid { aNote.Timestamp = timeStamp.Int64 * int64(time.Second / time.Nanosecond) }
		aNote.Save()
	}
}