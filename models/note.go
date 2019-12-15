package models

import (
	"log"
	"time"

)

//Note
type Note struct {
	Title string
	Datelog int64
	Content string
	URL string
	ReminderTicks int64
	Flags string
	Timestamp int64
	TimeSpent int64
	AuthorID int64
	GroupID int64
	Permission string
	RawEditor string
}

//NoteNew
func NoteNew(in map[string]interface{}) (*Note) {
	n := Note{}

	ct, ok := in["content"].(string)
	if !ok {
		// fmt.Printf("INFO. content is empty\n")
		ct = ""
	}
	titleText, ok := in["title"].(string)
	if !ok {
		// fmt.Printf("INFO No title provided, parse from content\n")
		if ct != ""{
			_l := len(ct)
			if _l >= 64 {_l = 64}
			titleText = ct[0:_l]
			n.Content = ct
		} else {
			// fmt.Printf("INFO No content and title provided. Not creating note\n")
			return &n
		}
	}
	n.Content = ct
	n.Title = titleText

	if dateData, ok := in["datelog"]; ok {
		switch v := dateData.(type) {
		case string:
			dateLog, e := time.Parse(DateLayout, v)
			if e != nil {
				log.Printf("ERROR can not parse date\n")
				n.Datelog = time.Now().UnixNano()
			} else {
				n.Datelog = dateLog.UnixNano()
			}
		case int64:
			n.Datelog = v
		}
	} else {
		n.Datelog = time.Now().UnixNano()
	}

	n.Timestamp = time.Now().UnixNano()

	if flags, ok := in["flags"]; ok {
		n.Flags = flags.(string)
	} else{
		n.Flags = ""
	}

	if url, ok := in["url"]; ok {
		n.URL = url.(string)
	} else{
		n.URL = ""
	}

	if flags, ok := in["flags"]; ok {
		n.Flags = flags.(string)
	} else{
		n.URL = ""
	}

	if authorid, ok := in["author_id"]; ok {
		n.AuthorID = authorid.(int64)
	} else{
		n.AuthorID = 0
	}

	if groupid, ok := in["group_id"]; ok {
		n.GroupID = groupid.(int64)
	} else{
		n.GroupID = 0
	}

	if perm, ok := in["permission"]; ok {
		n.Permission = perm.(string)
	} else{
		n.Permission = "3"
	}
	if raweditor, ok := in["raw_editor"]; ok {
		n.RawEditor = raweditor.(string)
	} else{
		n.RawEditor = "0"
	}
	return &n
}

func (n *Note) Save() (int64) {
	var noteID int64
	DB := GetDB("")
	defer DB.Close()
	var sql string
	currentNote := GetNote(n.Title)//This needs to be outside the BEGIN block othewise we get deadlock as Begin TX lock the whole db even for read (different from sqlite3)
	tx, _ := DB.Begin()
	if currentNote == nil {//New note
		sql = `INSERT INTO note(title, flags, content, url, datelog , reminder_ticks, timestamp, time_spent, author_id, group_id ,permission, raw_editor) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`
		res, e := tx.Exec(sql, n.Title, n.Flags, n.Content, n.URL, n.Datelog, n.ReminderTicks, n.Timestamp, n.TimeSpent, n.AuthorID, n.GroupID, n.Permission, n.RawEditor)
		if e != nil {
			tx.Rollback()
			log.Fatalf("ERROR can not insert note - %v\n", e)
		}
		noteID, _ = res.LastInsertId()
	} else {//Update existing
		sql = `UPDATE note SET flags = $1, content = $2, url = $3, datelog = $4, reminder_ticks = $5, timestamp = $6, time_spent = $7, author_id = $8, group_id = $9, permission = $10, raw_editor = $11 WHERE title = $12`
		res, e := tx.Exec(sql, n.Flags, n.Content, n.URL, n.Datelog, n.ReminderTicks, n.Timestamp, n.TimeSpent, n.AuthorID, n.GroupID, n.Permission, n.RawEditor, n.Title)
		if e != nil {
			tx.Rollback()
			log.Fatalf("ERROR can not insert note %v\n", e)
		}
		if c, _ := res.RowsAffected(); c != 1 {
			log.Fatalf("ERROR I expect only 1 note updated but got %d\n", c)
		}
		if e := DB.QueryRow(`SELECT id() FROM note WHERE title = $1`, n.Title).Scan(&noteID); e != nil {
			log.Fatalf("ERROR can not get back note row ID %v\n", e)
		}
	}
	tx.Commit()
	return noteID
}

func GetNote(title string) (*Note) {
	DB := GetDB("")
	defer DB.Close()
	n := Note{}
	// var flags, content, url string
	// var datelog , reminder_ticks, timestamp, time_spent int64
	// var author_id, group_id ,permission, raw_editor int8
	if e := DB.QueryRow(`SELECT flags, content, url, datelog , reminder_ticks, timestamp, time_spent, author_id, group_id ,permission, raw_editor FROM note WHERE title = $1`, title).Scan(&n.Flags, &n.Content, &n.URL, &n.Datelog, &n.ReminderTicks, &n.Timestamp, &n.TimeSpent, &n.AuthorID, &n.GroupID, &n.Permission, &n.RawEditor); e != nil {
		log.Printf("INFO - Can not find note title %s - %v\n", title, e)
		return nil
	}
	n.Title = title
	return &n
}