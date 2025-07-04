package models

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jbrodriguez/mlog"
	u "github.com/sunshine69/golang-tools/utils"
)

type Attachment struct {
	Object
	ID           int64
	Name         string
	Description  string
	Author       *User
	Group        *Group
	AttachedFile string
	FileSize     int64
	Mimetype     string
	Created      int64
	Updated      int64
}

func (a *Attachment) String() string {
	txt := a.Name + " - " + a.Mimetype + "- Created by " + a.Author.FirstName + " " + a.Author.LastName + " on " + u.NsToTime(a.Created).Format(DateLayout)
	return txt
}

func SearchAttachment(kw string, u *User) []*Attachment {
	DB := GetDB("")
	defer DB.Close()
	var o []*Attachment

	q := fmt.Sprintf(`SELECT
		id,
		name,
		description,
		author_id,
		group_id,
		permission,
		attached_file,
		file_size,
		mimetype,
		created,
		updated
		FROM attachment WHERE name LIKE '%%%s%%' OR attached_file LIKE '%%%s%%'
		ORDER BY updated desc LIMIT 200;`, kw, kw)

	// fmt.Printf("DEBUG SearchAttachment kw %s - query %s\n", kw, q)
	rows, e := DB.Query(q)
	if e != nil {
		mlog.Error(fmt.Errorf("search attachement - %s", e.Error()))
		return o
	}
	defer rows.Close()

	for rows.Next() {
		a := &Attachment{}
		if e := rows.Scan(&a.ID, &a.Name, &a.Description, &a.AuthorID, &a.GroupID, &a.Permission, &a.AttachedFile, &a.FileSize, &a.Mimetype, &a.Created, &a.Updated); e != nil {
			mlog.Error(fmt.Errorf("search attachement, scanning %s", e.Error()))
			continue
		}
		if pok := CheckPerm(a.Object, u.ID, "r"); pok {
			a.Update()
			o = append(o, a)
		}
	}
	return o
}

func (a *Attachment) Update() {
	a.Author = GetUserByID(a.AuthorID)
	a.Group = GetGroupByID(a.GroupID)
}

func (a *Attachment) Save() {
	DB := GetDB("")
	defer DB.Close()
	curAttachment := GetAttachement(a.Name)
	if curAttachment == nil {
		if a.Created == 0 {
			a.Created = time.Now().UnixNano()
		}
		if a.Updated == 0 {
			a.Updated = time.Now().UnixNano()
		}
		if a.AttachedFile == "" {
			a.AttachedFile = UpLoadPath + a.Name
		}
		tx, _ := DB.Begin()
		res, e := tx.Exec(`INSERT INTO attachment(
			name,
			description,
			author_id,
			group_id,
			permission,
			attached_file,
			file_size,
			mimetype,
			created,
			updated)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`, a.Name, a.Description, a.AuthorID, a.GroupID, a.Permission, a.AttachedFile, a.FileSize, a.Mimetype, a.Created, a.Updated)
		if e != nil {
			tx.Rollback()
			mlog.FatalIfError(fmt.Errorf("can not insert attachment - %s", e.Error()))
		}
		tx.Commit()
		a.ID, _ = res.LastInsertId()
	} else {
		tx, _ := DB.Begin()
		_, e := tx.Exec(`UPDATE attachment SET
			name = $1,
			description = $2,
			author_id = $3,
			group_id = $4,
			permission = $5,
			attached_file = $6,
			file_size = $7,
			mimetype = $8,
			created = $9,
			updated = $10
			WHERE name = $11`, a.Name, a.Description, a.AuthorID, a.GroupID, a.Permission, a.AttachedFile, a.FileSize, a.Mimetype, a.Created, a.Updated, a.Name)
		if e != nil {
			tx.Rollback()
			mlog.FatalIfError(fmt.Errorf("can not update attachment - %s", e.Error()))
		}
		tx.Commit()
	}
	a.Update()
}

func GetAttachement(aName string) *Attachment {
	DB := GetDB("")
	defer DB.Close()
	a := &Attachment{}
	if e := DB.QueryRow(`SELECT
		id,
		name ,
		description ,
		author_id  ,
		group_id  ,
		permission ,
		attached_file,
		file_size,
		mimetype ,
		created ,
		updated
		FROM attachment
		WHERE name = $1`, aName).Scan(&a.ID, &a.Name, &a.Description, &a.AuthorID, &a.GroupID, &a.Permission, &a.AttachedFile, &a.FileSize, &a.Mimetype, &a.Created, &a.Updated); e != nil {
		mlog.Warning("WARN No attachement %s found - %v\n", aName, e)
		return nil
	}
	a.Update()
	return a
}

func GetAttachementByID(id int64) *Attachment {
	DB := GetDB("")
	defer DB.Close()
	a := &Attachment{}
	if e := DB.QueryRow(`SELECT
		id,
		name,
		description,
		author_id,
		group_id,
		permission,
		attached_file,
		file_size,
		mimetype,
		created,
		updated
		FROM attachment
		WHERE id = $1`, id).Scan(&a.ID, &a.Name, &a.Description, &a.AuthorID, &a.GroupID, &a.Permission, &a.AttachedFile, &a.FileSize, &a.Mimetype, &a.Created, &a.Updated); e != nil {
		mlog.Warning("WARN No attachement ID %d found - %s\n", id, e.Error())
		return nil
	}
	a.Update()
	return a
}

func (a *Attachment) DeleteAttachment(u *User) error {
	DB := GetDB("")
	defer DB.Close()
	if !CheckPerm(a.Object, u.ID, "d") {
		return errors.New("permission denied")
	}
	//Check if it has links to note
	rows, e := DB.Query(`SELECT
		n.id,
		n.title
		FROM note_attachment AS na, note AS n
		WHERE na.note_id = n.id
		AND na.attachment_id = $1
	`, a.ID)
	if e != nil {
		return e
	}
	noteList := []*Note{}
	for rows.Next() {
		n := Note{}
		if e := rows.Scan(&n.ID, &n.Title); e != nil {
			mlog.Error(fmt.Errorf("scan in delete attachment %s", e.Error()))
			continue
		}
		noteList = append(noteList, &n)
	}
	if len(noteList) > 0 {
		msg := fmt.Sprintf("ERROR. This attachment is in use by the following notes as links - %v\n", noteList)
		return errors.New(msg)
	}
	tx, _ := DB.Begin()
	q := `DELETE FROM attachment WHERE id = $1;`

	_, e = tx.Exec(q, a.ID)

	if e != nil {
		etxt := fmt.Sprintf("ERROR can not remove attachment - %v\n", e)
		tx.Rollback()
		return errors.New(etxt)
	}
	tx.Commit()
	var _id int64
	if e := DB.QueryRow(`SELECT id from attachment WHERE attached_file = $1`, a.AttachedFile).Scan(&_id); e != nil { //No more attachment reference to this AttachedFile. Remove the file
		if e := os.Remove(a.AttachedFile); e != nil {
			etxt := fmt.Sprintf("ERROR removing the file - %v\n", e)
			return errors.New(etxt)
		}
	}
	return nil
}

// ScanAttachment - Scan files in the uploads folder or some locations and create the attachment object if not yet existed.
func ScanAttachment(dir string, user *User) []*Attachment {
	dir = u.Ternary(dir == "", "uploads", dir)
	o := []*Attachment{}

	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			a := GetAttachement(info.Name())
			if a == nil {
				fmt.Printf("INFO Create new attachment for %s Size: %d - Name %s\n", path, info.Size(), info.Name())
				a = &Attachment{
					Name:         info.Name(),
					AttachedFile: path,
					FileSize:     info.Size(),
				}
				a.AuthorID = user.ID
				a.GroupID = user.Groups[0].ID
				a.Permission = int8(1)
			} else {
				fmt.Printf("INFO attachment exists. Updating file path and size only ...\n")
				a.AttachedFile = path
				a.FileSize = info.Size()
			}
			a.Save()
			o = append(o, a)
			return nil
		})

	if err != nil {
		mlog.Error(err)
	}
	return o
}
