package models

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jbrodriguez/mlog"
	"github.com/sergi/go-diff/diffmatchpatch"
	u "github.com/sunshine69/golang-tools/utils"
)

// Object - Generic Object which has some special fields to allow us to check permissions etc...
// All other types (except user and group itself) should embed this type. Example: Note, Attachment
type Object struct {
	Permission int8
	AuthorID   int64
	GroupID    int64
}

// Note -
type Note struct {
	Object
	ID            int64
	Title         string
	Datelog       int64
	Content       string
	URL           string
	ReminderTicks int64
	Flags         string
	Timestamp     int64
	TimeSpent     int64
	Author        *User
	Group         *Group
	RawEditor     int8
	Attachments   []*Attachment
}

// Update - Populate dynamic fields such as Author, Group, etc. Not allowed saving data into DB
func (n *Note) Update() {
	if n.AuthorID != 0 {
		n.Author = GetUserByID(n.AuthorID)
	}
	if n.GroupID != 0 {
		n.Group = GetGroupByID(n.GroupID)
	}
	n.GetNoteAttachments()
}

func (n *Note) UnlinkAttachment(aID int64, u *User) error {
	if pok := CheckPerm(n.Object, u.ID, "w"); !pok {
		return errors.New("permission denied")
	}
	a := GetAttachementByID(aID)
	if pok := CheckPerm(a.Object, u.ID, "r"); !pok {
		return errors.New("permission denied")
	}
	DB := GetDB("")
	defer DB.Close()
	tx, _ := DB.Begin()
	q := `DELETE FROM note_attachment WHERE
		user_id = $1 AND
		note_id = $2 AND
		attachment_id = $3
		`
	if _, e := tx.Exec(q, u.ID, n.ID, a.ID); e != nil {
		tx.Rollback()
		return e
	}
	tx.Commit()
	return nil
}

func (n *Note) LinkAttachment(aID int64, u *User) error {
	if pok := CheckPerm(n.Object, u.ID, "w"); !pok {
		return errors.New("permission denied")
	}
	a := GetAttachementByID(aID)
	if pok := CheckPerm(a.Object, u.ID, "r"); !pok {
		return errors.New("permission denied")
	}
	DB := GetDB("")
	defer DB.Close()
	tx, _ := DB.Begin()
	q := `INSERT INTO note_attachment(
		user_id,
		note_id,
		attachment_id,
		timestamp)
		VALUES($1, $2, $3, $4)`
	if _, e := tx.Exec(q, u.ID, n.ID, a.ID, time.Now().UnixNano()); e != nil {
		tx.Rollback()
		return e
	}
	tx.Commit()
	return nil
}

func (n *Note) GetNoteAttachments() {
	DB := GetDB("")
	defer DB.Close()
	rows, e := DB.Query(`SELECT
		a.id,
		a.name,
		a.description,
		a.author_id,
		a.group_id,
		a.permission,
		a.attached_file,
		a.file_size,
		a.mimetype,
		a.created,
		a.updated
		FROM note_attachment AS na, attachment AS a
		WHERE na.attachment_id = a.id
		AND na.note_id = $1
	`, n.ID)
	if e != nil {
		mlog.FatalIfError(fmt.Errorf("GetNoteAttachments - %s", e.Error()))
	}

	defer rows.Close()

	var o []*Attachment

	for rows.Next() {
		a := Attachment{}
		if e := rows.Scan(&a.ID, &a.Name, &a.Description, &a.AuthorID, &a.GroupID, &a.Permission, &a.AttachedFile, &a.FileSize, &a.Mimetype, &a.Created, &a.Updated); e != nil {
			mlog.FatalIfError(fmt.Errorf("GetNoteAttachments can not fetch attachments - %s", e.Error()))
		}
		a.Update()
		o = append(o, &a)
	}
	n.Attachments = o
}

// NoteNew
func NoteNew(in map[string]interface{}) *Note {
	n := Note{}

	ct := u.GetMapByKey(in, "content", "").(string)
	titleText := u.GetMapByKey(in, "title", "").(string)

	if titleText == "" {
		if ct != "" {
			_l := len(ct)
			if _l >= 64 {
				_l = 64
			}
			titleText = strings.ReplaceAll(ct[0:_l], "\n", " ")
		}
	}
	n.Content = ct
	n.Title = titleText

	if dateData, ok := in["datelog"]; ok {
		switch v := dateData.(type) {
		case string:
			dateLog, e := time.Parse(DateLayout, v)
			if e != nil {
				mlog.Error(fmt.Errorf("can not parse date"))
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
	if TimeStamp, ok := in["timestamp"]; ok {
		n.Timestamp = TimeStamp.(int64)
	} else {
		n.Timestamp = time.Now().UnixNano()
	}
	n.Flags = u.GetMapByKey(in, "flags", "").(string)
	n.URL = u.GetMapByKey(in, "url", "").(string)

	n.AuthorID = u.GetMapByKey(in, "author_id", int64(0)).(int64)

	n.GroupID = u.GetMapByKey(in, "group_id", int64(1)).(int64)

	n.Permission = u.GetMapByKey(in, "permission", int8(1)).(int8)
	n.RawEditor = u.GetMapByKey(in, "raw_editor", int8(0)).(int8)

	n.Update()
	return &n
}

type NoteDiff struct {
	Flags   string
	Content string
	URL     string
}

func (nd *NoteDiff) String() string {
	return fmt.Sprintf("f: %s<br/>c: %s<br/>u: %s", nd.Flags, nd.Content, nd.URL)
}

// Diff - Compare two same title notes and find out the diff. If they are the same then return nil
// Only compare Flags, Content and URL
func (n *Note) Diff(n1 *Note) *NoteDiff {
	nd := NoteDiff{}
	if n.Flags == n1.Flags && n.Content == n1.Content && n.URL == n1.URL {
		return nil
	}

	dmp := diffmatchpatch.New()
	if n.Flags != n1.Flags {
		diffs := dmp.DiffMain(n.Flags, n1.Flags, false)
		nd.Flags = dmp.DiffPrettyHtml(diffs)
	}
	if n.Content != n1.Content {
		diffs := dmp.DiffMain(n.Content, n1.Content, false)
		nd.Content = dmp.DiffPrettyHtml(diffs)
	}
	if n.URL != n1.URL {
		diffs := dmp.DiffMain(n.URL, n1.URL, false)
		nd.URL = dmp.DiffPrettyHtml(diffs)
	}
	return &nd
}

// Save a note. If new note then create one. If existing note then create a revisions before update.
func (n *Note) Save() {
	currentNote := GetNote(n.Title) //This needs to be outside the BEGIN block othewise we get deadlock as Begin TX lock the whole db even for read (different from sqlite3)
	DB := GetDB("")
	defer DB.Close()
	var sql string
	if currentNote == nil { //New note
		if n.Title == "" {
			if n.Content != "" {
				_l := len(n.Content)
				if _l >= 64 {
					_l = 64
				}
				n.Title = strings.ReplaceAll(n.Content[0:_l], "\n", " ")
			}
		}
		mlog.Info("new note title %s\n", n.Title)
		tx, _ := DB.Begin()
		sql = `INSERT INTO note(
			title,
			flags,
			content,
			url,
			datelog,
			reminder_ticks,
			timestamp,
			time_spent,
			author_id,
			group_id,
			permission,
			raw_editor) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`
		res, e := tx.Exec(sql, n.Title, n.Flags, n.Content, n.URL, n.Datelog, n.ReminderTicks, n.Timestamp, n.TimeSpent, n.AuthorID, n.GroupID, n.Permission, n.RawEditor)
		if e != nil {
			tx.Rollback()
			mlog.FatalIfError(fmt.Errorf("can not insert note - %s", e.Error()))
		}
		n.ID, _ = res.LastInsertId()
		tx.Commit()
	} else { //Insert into revision and update current
		if n.Flags == currentNote.Flags &&
			n.Content == currentNote.Content &&
			n.URL == currentNote.URL &&
			n.Permission == currentNote.Permission &&
			n.RawEditor == currentNote.RawEditor &&
			n.GroupID == currentNote.GroupID {
			return
		}
		sql = `INSERT INTO note_revision(
			note_id,
			timestamp,
			flags,
			content,
			url,
			author_id,
			group_id,
			permission,
			raw_editor) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`
		tx, _ := DB.Begin()
		_, e := tx.Exec(sql, currentNote.ID, time.Now().UnixNano(), currentNote.Flags, currentNote.Content, currentNote.URL, currentNote.AuthorID, currentNote.GroupID, currentNote.Permission, currentNote.RawEditor)
		if e != nil {
			tx.Rollback()
			mlog.FatalIfError(fmt.Errorf("can not insert note %v", e))
		}
		tx.Commit()
		//Cleanup too old revision
		var timestampMark int
		revisionToKeep, _ := strconv.Atoi(GetConfig("revision_to_keep"))

		if e := DB.QueryRow(`SELECT timestamp FROM note_revision WHERE note_id = $1 ORDER BY timestamp ASC LIMIT 1 OFFSET $2`, currentNote.ID, revisionToKeep).Scan(&timestampMark); e != nil {
			tx, _ = DB.Begin()
			res, e := tx.Exec(`DELETE FROM note_revision WHERE timestamp < $1`, timestampMark)
			if e != nil {
				tx.Rollback()
				mlog.FatalIfError(fmt.Errorf("can not delete note_revision - %v", e))
			}
			af, _ := res.RowsAffected()
			tx.Commit()
			mlog.Info("Cleanup %d rows in note_revision\n", af)
		}
		//Update the note
		tx, _ = DB.Begin()
		sql = `UPDATE note SET
			flags = $1,
			content = $2,
			url = $3,
			datelog = $4,
			reminder_ticks = $5,
			timestamp = $6,
			time_spent = $7,
			author_id = $8,
			group_id = $9,
			permission = $10,
			raw_editor = $11 WHERE title = $12`
		_, e = tx.Exec(sql, n.Flags, n.Content, n.URL, n.Datelog, n.ReminderTicks, n.Timestamp, n.TimeSpent, n.AuthorID, n.GroupID, n.Permission, n.RawEditor, n.Title)
		if e != nil {
			tx.Rollback()
			mlog.FatalIfError(fmt.Errorf("can not update note %v", e))
		}
		tx.Commit()
	}
}

func GetNoteByID(id int64) *Note {
	DB := GetDB("")
	defer DB.Close()
	n := Note{ID: id}
	if e := DB.QueryRow(`SELECT
		title,
		flags,
		content,
		url,
		datelog,
		reminder_ticks,
		timestamp,
		time_spent,
		author_id,
		group_id,
		permission,
		raw_editor
		FROM note WHERE id = $1`, id).Scan(&n.Title, &n.Flags, &n.Content, &n.URL, &n.Datelog, &n.ReminderTicks, &n.Timestamp, &n.TimeSpent, &n.AuthorID, &n.GroupID, &n.Permission, &n.RawEditor); e != nil {
		// log.Printf("INFO - Can not find note ID %d - %v\n", id, e)
		return nil
	}
	n.Update()
	return &n
}

func GetNote(title string) *Note {
	DB := GetDB("")
	defer DB.Close()
	n := Note{Title: title}
	if e := DB.QueryRow(`SELECT
		id,
		flags,
		content,
		url,
		datelog,
		reminder_ticks,
		timestamp,
		time_spent,
		author_id,
		group_id,
		permission,
		raw_editor
		FROM note WHERE title = $1`, title).Scan(&n.ID, &n.Flags, &n.Content, &n.URL, &n.Datelog, &n.ReminderTicks, &n.Timestamp, &n.TimeSpent, &n.AuthorID, &n.GroupID, &n.Permission, &n.RawEditor); e != nil {
		mlog.Info("Can not find note title %s - %v\n", title, e)
		return nil
	}
	n.Update()
	return &n
}

func GetNoteRevisionByID(id int64) *Note {
	n := Note{
		ID: id,
	}
	DB := GetDB("")
	defer DB.Close()

	if e := DB.QueryRow(`SELECT id, timestamp, flags, url, content, author_id, group_id, permission, raw_editor FROM note_revision WHERE id = $1`, id).Scan(&n.ID, &n.Datelog, &n.Flags, &n.URL, &n.Content, &n.AuthorID, &n.GroupID, &n.Permission, &n.RawEditor); e != nil {
		mlog.Error(fmt.Errorf("can not get note revision - %v", e))
	} else {
		n.Update()
	}
	return &n
}

// GetNoteRevision - Get all revision of a note. Pass in identity which can be note_id (int64) or title (string). The first result in the slice is the current version of the note. Next is all revision order by timestamp
func GetNoteRevisions(noteIdentity interface{}) []Note {
	o := []Note{}
	var cNote *Note
	noteID, ok := noteIdentity.(int64)
	if ok {
		cNote = GetNoteByID(noteID)
	} else {
		title, ok := noteIdentity.(string)
		if !ok {
			mlog.Warning("WARN GetNoteRevisions does not have correct type. It needs to be an noteID or note title - \n")
			return o
		}
		cNote = GetNote(title)
		cNote.Update()
	}
	// o = append(o, *cNote)

	noteID = cNote.ID
	DB := GetDB("")
	defer DB.Close()

	res, e := DB.Query(`SELECT id, timestamp, flags, url, content, author_id, group_id,	permission FROM note_revision WHERE note_id = $1 ORDER BY timestamp DESC LIMIT 200`, noteID)
	if e != nil {
		mlog.FatalIfError(fmt.Errorf("can not get note revision - %v", e))
	}
	for res.Next() {
		n := Note{}
		res.Scan(&n.ID, &n.Timestamp, &n.Flags, &n.URL, &n.Content, &n.AuthorID, &n.GroupID, &n.Permission)
		n.Title = cNote.Title
		n.Datelog = cNote.Datelog
		n.Update()
		o = append(o, n)
	}
	return o
}

func (n *Note) String() string { return n.Title }

func NoteDeleteByID(id int64) bool {
	var o bool
	DB := GetDB("")
	defer DB.Close()
	tx, _ := DB.Begin()
	_, e := tx.Exec(`DELETE FROM note WHERE id = $1`, id)
	if e != nil {
		tx.Rollback()
		mlog.Warning("WARN Can not delete note ID %d\n", id)
		o = false
	} else {
		tx.Commit()
		o = true
	}
	return o
}

// Delete - Delete note
func (n *Note) Delete() {
	DB := GetDB("")
	defer DB.Close()
	tx, _ := DB.Begin()
	_, e := tx.Exec(`DELETE FROM note WHERE title = $1`, n.Title)
	if e != nil {
		tx.Rollback()
		mlog.Warning("WARN Can not delete note %v - %v\n", n, e)
	} else {
		tx.Commit()
	}
}

// Search by keyword. Type a keyword it will search that kw. To search for `needlA` and `needB` type `needleA & needleB`. If search `A` but exclude B
// then `A & !B` or `A & -B`
// You can search by note flags only, by prefix then using `f:` or `F:`, `FLAGS:`
func SearchNote(keyword string, u *User) []Note {
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
		for i, t := range tokens {
			if i == _l-1 {
				q = fmt.Sprintf(`%s (flags LIKE "%%%s%%") ORDER BY datelog DESC LIMIT 200`, q, t)
			} else {
				q = fmt.Sprintf(`%s (flags LIKE "%%%s%%") OR `, q, t)
			}
		}
		q = fmt.Sprintf("SELECT id as note_id, title, flags, content, url, datelog , reminder_ticks, timestamp, time_spent, author_id, group_id ,permission, raw_editor from note WHERE %s", q)
	} else {
		_l := len(tokens)

		for i, t := range tokens {
			negate := ""
			combind := "OR"
			if strings.HasPrefix(t, "!") || strings.HasPrefix(t, "-") {
				negate = " NOT "
				t = strings.Replace(t, "!", "", 1)
				t = strings.Replace(t, "-", "", 1)
				combind = "AND"
			}
			if i == _l-1 {
				q = fmt.Sprintf(`%s (%s(title LIKE "%%%s%%") %s %s(flags LIKE "%%%s%%") %s %s(content LIKE "%%%s%%") %s %s(url LIKE "%%%s%%")) ORDER BY timestamp DESC`, q, negate, t, combind, negate, t, combind, negate, t, combind, negate, t)
			} else {
				q = fmt.Sprintf(`%s (%s(title LIKE "%%%s%%") %s %s(flags LIKE "%%%s%%") %s %s(content LIKE "%%%s%%") %s %s(url LIKE "%%%s%%")) AND `, q, negate, t, combind, negate, t, combind, negate, t, combind, negate, t)
			}
		}
		q = fmt.Sprintf("SELECT id as note_id, title, flags, content, url, datelog , reminder_ticks, timestamp, time_spent, author_id, group_id ,permission, raw_editor from note WHERE %s", q)
	}
	if !strings.Contains(q, "LIMIT") {
		q = fmt.Sprintf("%s LIMIT 100;", q)
	}
	if os.Getenv("DEBUG") != "" {
		fmt.Println(q)
	}
	DB := GetDB("")
	defer DB.Close()
	res, e := DB.Query(q)
	mlog.FatalIfError(e)

	o := []Note{}

	for res.Next() {
		n := Note{}
		res.Scan(&n.ID, &n.Title, &n.Flags, &n.Content, &n.URL, &n.Datelog, &n.ReminderTicks, &n.Timestamp, &n.TimeSpent, &n.AuthorID, &n.GroupID, &n.Permission, &n.RawEditor)
		if pok := CheckPerm(n.Object, u.ID, "r"); pok {
			n.Update()
			o = append(o, n)
		}
	}
	return o
}

func Query(sqlwhere string, user *User, without_content bool) []Note {
	DB := GetDB("")
	defer DB.Close()
	var sql string
	if without_content {
		sql = "SELECT id as note_id, title, flags, url, datelog, reminder_ticks, timestamp, time_spent, author_id, group_id ,permission, raw_editor from note WHERE " + sqlwhere
	} else {
		sql = "SELECT id as note_id, title, flags, content, url, datelog, reminder_ticks, timestamp, time_spent, author_id,group_id ,permission, raw_editor from note WHERE " + sqlwhere
	}
	res, err := DB.Query(sql)
	if u.CheckErrNonFatal(err, "Query") != nil {
		return []Note{}
	}
	o := []Note{}
	for res.Next() {
		n := Note{}
		if without_content {
			res.Scan(&n.ID, &n.Title, &n.Flags, &n.URL, &n.Datelog, &n.ReminderTicks, &n.Timestamp, &n.TimeSpent, &n.AuthorID, &n.GroupID, &n.Permission, &n.RawEditor)
		} else {
			res.Scan(&n.ID, &n.Title, &n.Flags, &n.Content, &n.URL, &n.Datelog, &n.ReminderTicks, &n.Timestamp, &n.TimeSpent, &n.AuthorID, &n.GroupID, &n.Permission, &n.RawEditor)
		}
		if pok := CheckPerm(n.Object, user.ID, "r"); pok {
			// We do not need to update to get extra link infor lie author etc - this query func is explicitly return only note data. For full note, use GetNoteByID or searchNote instead
			// This func is only used when interfacng with external system, like gnote.
			// n.Update()
			o = append(o, n)
		}
	}
	return o
}
