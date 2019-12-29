package models

import (
	"log"
)

type Attachment struct {
	ID int64
	Name string
	Description string
	AuthorID int64
	Author *User
	GroupID int8
	Group *Group
	Permission int8
	AttachedFile string
	Mimetype string
	Created int64
	Updated int64
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
		tx, _ := DB.Begin()
		res, e := tx.Exec(`INSERT INTO attachment(
			name,
			description,
			author_id,
			group_id,
			permission,
			attached_file,
			mimetype,
			created,
			updated)
			VALUES($1, $2, $3, int8($4), int8($5), $6, $7, $8, $9)`, a.Name, a.Description, a.AuthorID, a.GroupID, a.Permission, a.AttachedFile, a.Mimetype, a.Created, a.Updated)
		if e != nil {
			tx.Rollback()
			log.Fatalf("ERROR can not insert attachment - %v\n", e)
		}
		tx.Commit()
		a.ID, _ = res.LastInsertId()
	} else {
		tx, _ := DB.Begin()
		_, e := tx.Exec(`UPDATE attachment SET
			name = $1,
			description = $2,
			author_id = $3,
			group_id = int8($4),
			permission = int8($5),
			attached_file = $6,
			mimetype = $7,
			created = $8,
			updated = $9,
			WHERE name = $10`, a.Name, a.Description, a.AuthorID, a.GroupID, a.Permission, a.AttachedFile, a.Mimetype, a.Created, a.Updated, a.Name)
		if e != nil {
			tx.Rollback()
			log.Fatalf("ERROR can not update attachment - %v\n", e)
		}
		tx.Commit()
	}
	a.Update()
}

func GetAttachement(aName string) *Attachment {
	DB := GetDB(""); defer DB.Close()
	a := Attachment{}
	if e := DB.QueryRow(`SELECT
		id() as id,
		name ,
		description ,
		author_id  ,
		group_id  ,
		permission ,
		attached_file ,
		mimetype ,
		created ,
		updated
		FROM attachment
		WHERE name = $1`, aName).Scan(&a.ID, &a.Name, &a.Description, &a.AuthorID, &a.GroupID, &a.Permission, &a.AttachedFile, &a.Mimetype, &a.Created, &a.Updated); e != nil {
			log.Printf("WARN No attachement %s found - %v\n", aName, e)
			return nil
	}
	a.Update()
	return &a
}

func GetAttachementByID(id int64) *Attachment {
	DB := GetDB(""); defer DB.Close()
	a := Attachment{}
	if e := DB.QueryRow(`SELECT
		id() as id,
		name,
		description,
		author_id,
		group_id,
		permission,
		attached_file,
		mimetype,
		created,
		updated
		FROM attachment
		WHERE id() = $1`, id).Scan(&a.ID, &a.Name, &a.Description, &a.AuthorID, &a.GroupID, &a.Permission, &a.AttachedFile, &a.Mimetype, &a.Created, &a.Updated); e != nil {
			log.Printf("WARN No attachement ID %d found - %v\n", id, e)
			return nil
	}
	a.Update()
	return &a
}