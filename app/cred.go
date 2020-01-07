package app

import (
	"time"
	"fmt"
	"log"
	"net/http"
	m "github.com/sunshine69/webnote-go/models"
)

func DoCredApp(w http.ResponseWriter, r *http.Request) {
	action := m.GetRequestValue(r, "action", "")
	switch action {
	case "cred_add":
		DoCredAdd(&w, r)
	case "cred_delete":
		DoCredDelete(&w, r)
	case "cred_search":
		DoCredSearch(&w, r)
	case "update_qrlink":
		DoCredUpdateQrlink(&w, r)
	case "generate_qr":
		DoCredGenerateQr(&w, r)
	}
}

func DoCredGenerateQr(w *http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(*w, "TODO")
}

func DoCredUpdateQrlink(w *http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(*w, "TODO")
}

func DoCredAdd(w *http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(*w, "TODO")
}

func DoCredDelete(w *http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(*w, "TODO")
}

func DoCredSearch(w *http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(*w, "TODO")
}

func SetupSchema() {
	q := `-- credential app
	CREATE TABLE IF NOT EXISTS credential (
		id integer NOT NULL PRIMARY KEY,
		user_id integer,
		cred_username text ,
		cred_password text
	);
	CREATE UNIQUE INDEX IF NOT EXISTS credentialidx ON credential(user_id, cred_username, cred_password);

	CREATE TABLE IF NOT EXISTS url (
		id integer NOT NULL PRIMARY KEY,
		url text
	);

	CREATE TABLE IF NOT EXISTS url_cred (
		id integer NOT NULL PRIMARY KEY,
		cred_id integer,
		url_id integer,
		note text,
		datelog integer,
		qrlink text
	);
	CREATE UNIQUE INDEX IF NOT EXISTS url_credidx ON url_cred(cred_id, url_id);
-- End credential app`
	DB := m.GetDB(""); defer DB.Close()
	tx, _ := DB.Begin()
	_, err := tx.Exec(q)
	if err != nil {
		tx.Rollback()
		log.Fatalf("ERROR setup schema for cred app %v\n", err)
	}
	tx.Commit()
}

type Credential struct {
	Id int64
	User_id int64
	Cred_username string
	Cred_password string
	Url string
}

type Url struct {
	Id int64
	Url string
}

type UrlCred struct {
	Id int64
	Cred_id int64
	Credential *Credential
	Url_id int64
	Url *Url
	Note string
	Datelog int64
	Qrlink string
}

func (uc *UrlCred) Save() {
	DB := m.GetDB(""); defer DB.Close()
	tx, _ := DB.Begin()
	q := `INSERT OR REPLACE INTO url_cred(
		cred_id,
		url_id,
		note,
		datelog,
		qrlink)
		VALUES($1, $2, $3, $4, $5)
	`
	res, err := tx.Exec(q, uc.Cred_id, uc.Url_id, uc.Note, uc.Datelog, uc.Qrlink)
	if err != nil {
		log.Printf("WARN - error insert url_cred %v\n", err)
		tx.Rollback()
	} else {
		tx.Commit()
		uc.Id, _ = res.LastInsertId()
		uc.Update()
	}
}

func UrlCredNew(in map[string]interface{}) *UrlCred {
	uc := UrlCred{}
	uc.Cred_id = m.GetMapByKey(in, "cred_id", 0).(int64)
	uc.Url_id = m.GetMapByKey(in, "url_id", 0).(int64)
	uc.Note = m.GetMapByKey(in, "note", "").(string)
	uc.Datelog = time.Now().UnixNano()
	uc.Qrlink = m.GetMapByKey(in, "qr_link", "").(string)
	uc.Save()
	uc.Update()
	return &uc
}

func (uc *UrlCred) Update() {
	uc.Credential = GetCredentialByID(uc.Cred_id)
	uc.Url = GetUrlByID(uc.Url_id)
}

func GetUrlByID(id int64) *Url {
	DB := m.GetDB(""); defer DB.Close()
	q := `SELECT
	id,
	url FROM url WHERE id = $1`
	u := Url{}
	if e := DB.QueryRow(q, id).Scan(&u.Id, &u.Url); e != nil {
		log.Printf("WARN GetUrlByID %v\n", e)
		return nil
	}
	return &u
}

func SearchCredentials(kw string, UserID int64) *[]UrlCred {
	DB := m.GetDB(""); defer DB.Close()
	q := `SELECT
	u.id,
	u.cred_id,
	u.url_id,
	u.note,
	u.datelog,
	u.qrlink
	FROM url_cred as u, credential as c, url
	WHERE c.id == u.cred_id
	AND url.id = u.url_id
	AND c.user_id = $1
	AND ((url.url LIKE "%$2%")
		OR (u.note LIKE "%$2%")
	)
	`
	rows, err := DB.Query(q, UserID, kw)
	if err != nil {
		log.Printf("WARN Error searching url-cred - %v\n", err)
		return nil
	}
	o := []UrlCred{}
	for rows.Next(){
		UrlCred := UrlCred{}
		if err := rows.Scan(
			&UrlCred.Id, &UrlCred.Cred_id, &UrlCred.Url_id, &UrlCred.Note, &UrlCred.Datelog, &UrlCred.Qrlink,
		); err != nil {
			log.Printf("WARN Error searching url-cred - %v\n", err)
			return &o
		}
		o = append(o, UrlCred)
	}
	return &o
}

func GetCredentialByID(id int64) *Credential {
	DB := m.GetDB(""); defer DB.Close()
	q := `SELECT
	id,
	user_id,
	cred_username,
	cred_password
	FROM credential WHERE id = $1`
	cred := Credential{}
	if err := DB.QueryRow(q, id).Scan(&cred.Id, &cred.User_id, &cred.Cred_username, &cred.Cred_password); err != nil {
		log.Printf("WARN Can not get cred with this ID %d, %v\n", id, err)
		return nil
	}
	return &cred
}

func (c *Credential) Save() {

}