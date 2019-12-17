package models

import (
	"regexp"
	"fmt"
	"strings"
	"log"
	"time"
	"strconv"
)

//Note
type Note struct {
	ID int64
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

//Save a note. If new note then create on. If existing note then create a revisions before update.
func (n *Note) Save() {
	var noteID int64
	currentNote := GetNote(n.Title)//This needs to be outside the BEGIN block othewise we get deadlock as Begin TX lock the whole db even for read (different from sqlite3)
	DB := GetDB("")
	defer DB.Close()
	var sql string

	if currentNote == nil {//New note
		tx, _ := DB.Begin()
		sql = `INSERT INTO note(title, flags, content, url, datelog , reminder_ticks, timestamp, time_spent, author_id, group_id ,permission, raw_editor) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`
		res, e := tx.Exec(sql, n.Title, n.Flags, n.Content, n.URL, n.Datelog, n.ReminderTicks, n.Timestamp, n.TimeSpent, n.AuthorID, n.GroupID, n.Permission, n.RawEditor)
		if e != nil {
			tx.Rollback()
			log.Fatalf("ERROR can not insert note - %v\n", e)
		}
		noteID, _ = res.LastInsertId()
		tx.Commit()
	} else {//Insert into revision and update current
		sql = `INSERT INTO note_revision(note_id, timestamp, flags, content, url, author_id, group_id, permission) VALUES($1, $2, $3, $4, $5, $6, $7, $8)`
		tx, _ := DB.Begin()
		res, e := tx.Exec(sql, currentNote.ID, time.Now().UnixNano(), currentNote.Flags, currentNote.Content, currentNote.URL, currentNote.AuthorID, currentNote.GroupID, currentNote.Permission)
		if e != nil {
			tx.Rollback()
			log.Fatalf("ERROR can not insert note %v\n", e)
		}
		tx.Commit()
		//Cleanup too old revision
		var timestampMark int
		revisionToKeep, _ := strconv.Atoi(GetConfig("revision_to_keep"))

		if e := DB.QueryRow(`SELECT timestamp FROM note_revision WHERE note_id = $1 ORDER BY timestamp ASC LIMIT 1 OFFSET $2`, currentNote.ID, revisionToKeep).Scan(&timestampMark); e != nil {
			tx, _ = DB.Begin()
			res, e = tx.Exec(`DELETE FROM note_revision WHERE timestamp < $1`, timestampMark)
			if e != nil {
				tx.Rollback()
				log.Fatalf("ERROR can not delete note_revision - %v\n", e)
			}
			af, _ := res.RowsAffected()
			tx.Commit()
			log.Printf("INFO Cleanup %d rows in note_revision\n", af)
		}
		//Update the note
		tx, _ = DB.Begin()
		sql = `UPDATE note SET flags = $1, content = $2, url = $3, datelog = $4, reminder_ticks = $5, timestamp = $6, time_spent = $7, author_id = $8, group_id = $9, permission = $10, raw_editor = $11 WHERE title = $12`
		res, e = tx.Exec(sql, n.Flags, n.Content, n.URL, n.Datelog, n.ReminderTicks, n.Timestamp, n.TimeSpent, n.AuthorID, n.GroupID, n.Permission, n.RawEditor, n.Title)
		if e != nil {
			tx.Rollback()
			log.Fatalf("ERROR can not insert note %v\n", e)
		}
		tx.Commit()
	}
	n.ID = noteID
}

func GetNote(title string) (*Note) {
	DB := GetDB("")
	defer DB.Close()
	n := Note{ Title: title }
	// var flags, content, url string
	// var datelog , reminder_ticks, timestamp, time_spent int64
	// var author_id, group_id ,permission, raw_editor int8
	if e := DB.QueryRow(`SELECT id() as note_id, flags, content, url, datelog , reminder_ticks, timestamp, time_spent, author_id, group_id ,permission, raw_editor FROM note WHERE title = $1`, title).Scan(&n.ID, &n.Flags, &n.Content, &n.URL, &n.Datelog, &n.ReminderTicks, &n.Timestamp, &n.TimeSpent, &n.AuthorID, &n.GroupID, &n.Permission, &n.RawEditor); e != nil {
		log.Printf("INFO - Can not find note title %s - %v\n", title, e)
		return nil
	}
	return &n
}

func GetNoteRevision(noteIdentity interface{}) []Note {
	o := []Note{}
	DB := GetDB("")
	defer DB.Close()
	cNote := Note{Title: ""}
	noteID, ok := noteIdentity.(int64)
	if ok {
		DB.QueryRow(`SELECT title, flags, content, url, datelog , reminder_ticks, timestamp, time_spent, author_id, group_id ,permission, raw_editor FROM note WHERE id() = $1`, noteID).Scan(&cNote.Title, &cNote.Flags, &cNote.Content, &cNote.URL, &cNote.Datelog, &cNote.ReminderTicks, &cNote.Timestamp, &cNote.TimeSpent, &cNote.AuthorID, &cNote.GroupID, &cNote.Permission, &cNote.RawEditor)
		cNote.ID = noteID
	} else {
		cNote.Title, ok = noteIdentity.(string)
		if !ok {
			log.Printf("WARN GetNoteRevision does not have correct type. It needs to be an noteID or note title - \n")
			return o
		}
		DB.QueryRow(`SELECT id() as note_id, flags, content, url, datelog , reminder_ticks, timestamp, time_spent, author_id, group_id ,permission, raw_editor FROM note WHERE title = $1`, cNote.Title).Scan(&cNote.ID, &cNote.Flags, &cNote.Content, &cNote.URL, &cNote.Datelog, &cNote.ReminderTicks, &cNote.Timestamp, &cNote.TimeSpent, &cNote.AuthorID, &cNote.GroupID, &cNote.Permission, &cNote.RawEditor)
	}
	o = append(o, cNote)

	res, e := DB.Query(`SELECT timestamp, flags, url, content, author_id, group_id,	permission FROM note_revision WHERE note_id = $1`, noteID)
	if e != nil {
		log.Fatalf("ERROR can not get note revision - %v\n", e)
	}
	for res.Next() {
		n := Note{}
		res.Scan(&n.Timestamp, &n.Flags, &n.URL, &n.Content, &n.AuthorID, &n.GroupID, &n.Permission)
		n.Title = cNote.Title
		n.Datelog = cNote.Datelog
		o = append(o, n)
	}
	return o
}

func (n *Note) String() string {return n.Title}

//Delete - Delete note
func (n *Note) Delete() {
	DB := GetDB("")
	defer DB.Close()
	tx, _ := DB.Begin()
	_, e := tx.Exec(`DELETE FROM note WHERE title = $1`, n.Title)
	if e != nil {
		tx.Rollback()
		log.Printf("WARN Can not delete note %v - %v\n", n, e)
	}
	tx.Commit()
}

func SearchNote(keyword string) []Note {
	keyword = strings.TrimSpace(keyword)
	splitPtn := regexp.MustCompile(`[\s]+[\&\+][\s]+`)
	var q string
	searchFlags := false
	tokens := splitPtn.Split(keyword, -1)

	if strings.HasPrefix(keyword, "F:") || strings.HasPrefix(keyword, "f:") {
		tokens = strings.Split(keyword[2:], ":")
		searchFlags = true
	} else if strings.HasPrefix(keyword, "FLAGS:") || strings.HasPrefix(keyword, "flags:") {
		tokens = strings.Split(keyword[6:], ":")
		searchFlags = true
	}
	if searchFlags {
		_l := len(tokens)
		for i, t := range(tokens) {
			if i == _l - 1 {
				q = fmt.Sprintf(`%s (flags LIKE "%s") ORDER BY datelog DESC LIMIT 200;`, q, t)
			} else {
				q = fmt.Sprintf(`%s (flags LIKE "%s") OR `, q, t)
			}
		}
		q = fmt.Sprintf("SELECT id() as note_id, flags, content, url, datelog , reminder_ticks, timestamp, time_spent, author_id, group_id ,permission, raw_editor from note WHERE %s", q)
	} else {
		_l := len(tokens)

		for i, t := range(tokens) {
			negate := ""
			combind := "OR"
			if strings.HasPrefix(t, "!") || strings.HasPrefix(t, "-") {
				negate = " ! "
				t = strings.Replace(t, "!", "", 1)
				t = strings.Replace(t, "-", "", 1)
				combind = "AND"
			}
			if i == _l - 1 {
				q = fmt.Sprintf(`%s (%s(title LIKE "%s") %s %s(flags LIKE "%s") %s %s(content LIKE "%s") %s %s(url LIKE "%s")) ORDER BY timestamp DESC;`, q, negate, t, combind, negate, t, combind, negate, t, combind, negate, t)
			} else {
				q = fmt.Sprintf(`%s (%s(title LIKE "%s") %s %s(flags LIKE "%s") %s %s(content LIKE "%s") %s %s(url LIKE "%s")) AND `, q, negate, t, combind, negate, t, combind, negate, t, combind, negate, t)
			}
		}
		q = fmt.Sprintf("SELECT id() as note_id, flags, content, url, datelog , reminder_ticks, timestamp, time_spent, author_id, group_id ,permission, raw_editor from note WHERE %s", q)
	}
	// fmt.Println(q)
	DB := GetDB("")
	defer DB.Close()
	res, e := DB.Query(q)
	if e != nil {
		log.Fatalf("ERROR query - %v\n", e)
	}
	o := []Note{}

	for res.Next() {
		n := Note{}
		res.Scan(&n.ID, &n.Flags, &n.Content, &n.URL, &n.Datelog, &n.ReminderTicks, &n.Timestamp, &n.TimeSpent, &n.AuthorID, &n.GroupID, &n.Permission, &n.RawEditor)
		o = append(o, n)
	}
	return o
}