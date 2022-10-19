package models

import (
	"database/sql"
	"os"
	"time"

	"github.com/jbrodriguez/mlog"
)

func Migrate() {
	os.Setenv("DBPATH", "/home/stevek/src/webnote-go/testwebnote.db")
	sqlite3File := "/home/stevek/webnote.sqlite3"
	DB, e := sql.Open("sqlite3", sqlite3File)

	mlog.FatalIfError(e)

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
	mlog.FatalIfError(e)

	for rows.Next() {
		aNote := Note{}
		aNote.AuthorID = int64(1)
		aNote.GroupID = int64(1)
		var nContent, nURL, nFlags sql.NullString
		var dateLog, timeStamp sql.NullInt64
		if err := rows.Scan(
			&aNote.Title, &dateLog, &nContent, &nURL,
			&nFlags, &timeStamp, &aNote.Permission, &aNote.RawEditor,
		); err != nil {
			mlog.Fatal("Scan result %v\n", err)
		}
		if nContent.Valid {
			aNote.Content = nContent.String
		}
		if nURL.Valid {
			aNote.URL = nURL.String
		}
		if nFlags.Valid {
			aNote.Flags = nFlags.String
		}
		if dateLog.Valid {
			aNote.Datelog = dateLog.Int64 * int64(time.Second/time.Nanosecond)
		}
		if timeStamp.Valid {
			aNote.Timestamp = timeStamp.Int64 * int64(time.Second/time.Nanosecond)
		}
		aNote.Save()
	}
}
