package app

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jbrodriguez/mlog"
	u "github.com/sunshine69/golang-tools/utils"
	m "github.com/sunshine69/webnote-go/models"
)

func init() {
	mlog.Start(mlog.LevelInfo, "webnote.log")
}

// Credential Application. To setup in the browser access https://<your_dns>:<your_port>/cred?action=setup and follow the instruction
// This is a simple onepage app to store/search credentials (password management)
// This is per webnote user and each user has no view of other user credential
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
	case "setup":
		SetupSchema(&w, r)
	}
}

func DoCredGenerateQr(w *http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(*w, "TODO")
}

func DoCredUpdateQrlink(w *http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(*w, "TODO")
}

func DoCredAdd(w *http.ResponseWriter, r *http.Request) {
	cred_url := m.GetRequestValue(r, "cred_url", "")
	cred_username := m.GetRequestValue(r, "cred_username", "")
	cred_password := m.GetRequestValue(r, "cred_password", "")
	cred_note := m.GetRequestValue(r, "cred_note", "")
	qrlink := m.GetRequestValue(r, "qrlink", "")

	u := UrlNew(cred_url)

	useremail := m.GetSessionVal(r, "useremail", "").(string)
	user := m.GetUser(useremail)
	c := CredentialNew(map[string]interface{}{
		"user_id":       user.ID,
		"cred_username": cred_username,
		"cred_password": cred_password,
	})
	uc := UrlCredNew(map[string]interface{}{
		"cred_id":   c.Id,
		"url_id":    u.Id,
		"cred_note": cred_note,
		"qrlink":    qrlink,
	})
	fmt.Fprintf(*w, "OK Cred URL added ID %d", uc.Id)
}

func DoCredDelete(w *http.ResponseWriter, r *http.Request) {
	msg := "OK deleted"
	useremail := m.GetSessionVal(r, "useremail", "").(string)
	user := m.GetUser(useremail)
	idStr := m.GetRequestValue(r, "id", "-1")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if id != -1 {
		if uc := GetUrlCredByID(id); uc != nil {
			mlog.Warning("%v", uc)
			c := uc.Credential
			mlog.Warning("%v", c)
			if c.User_id == user.ID {
				uc.Delete()
			} else {
				msg = "Permission denied"
			}
		}
	}
	fmt.Fprint(*w, msg)
}

func DoCredSearch(w *http.ResponseWriter, r *http.Request) {
	useremail := m.GetSessionVal(r, "useremail", "").(string)
	user := m.GetUser(useremail)
	kw := m.GetRequestValue(r, "kw", "")
	ucs := SearchCredentials(kw, user.ID)
	commonMapData := map[string]interface{}{
		"user":                user,
		"date_layout":         m.DateLayout,
		"cred_search_results": ucs,
	}
	var tpl bytes.Buffer
	if err := m.AllTemplates.ExecuteTemplate(&tpl, "cred_search_results.html", commonMapData); err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprint(*w, tpl.String())
}

func SetupSchema(w *http.ResponseWriter, r *http.Request) {
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
	CREATE UNIQUE INDEX IF NOT EXISTS urlidx ON url(url);

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
	DB := m.GetDB("")
	defer DB.Close()
	tx, _ := DB.Begin()
	_, err := tx.Exec(q)
	if err != nil {
		tx.Rollback()
		http.Error(*w, err.Error(), http.StatusInternalServerError)
		mlog.Error(fmt.Errorf("setup schema for cred app %v", err))
	}
	tx.Commit()
	uitext, _ := os.ReadFile("./app/cred.html")
	responseText := fmt.Sprintf(`Database schema setup completed successfully. Below is the UI text. Copy this and create a note with that content with title you wish. When save and then use view2 you will see the GUI of the app.
	>>>>> UI TEXT <<<<<
	%s`, uitext)
	fmt.Fprint(*w, responseText)
}

type Credential struct {
	Id            int64
	User_id       int64
	Cred_username string
	Cred_password string
}

type Url struct {
	Id  int64
	Url string
}

type UrlCred struct {
	Id         int64
	Cred_id    int64
	Credential *Credential
	Url_id     int64
	Url        *Url
	Note       string
	Datelog    int64
	Qrlink     string
}

func (uc *UrlCred) Save() {
	DB := m.GetDB("")
	defer DB.Close()
	tx, _ := DB.Begin()
	var q string
	if uc.Id == -1 {
		q = `INSERT INTO url_cred(
			cred_id,
			url_id,
			note,
			datelog,
			qrlink)
			VALUES($1, $2, $3, $4, $5)
		`
	} else {
		q = fmt.Sprintf(`UPDATE url_cred SET
			cred_id = $1,
			url_id  = $2,
			note    = $3,
			datelog = $4,
			qrlink  = $5  WHERE id = %d`, uc.Id)
	}
	res, err := tx.Exec(q, uc.Cred_id, uc.Url_id, uc.Note, uc.Datelog, uc.Qrlink)
	if err != nil {
		tx.Rollback()
		mlog.Warning("WARN - Save insert url_cred %v\n", err)
	} else {
		tx.Commit()
		uc.Id, _ = res.LastInsertId()
		uc.Update()
	}
}

func UrlNew(url string) *Url {
	u := GetUrl(url)
	DB := m.GetDB("")
	defer DB.Close()
	tx, _ := DB.Begin()
	q := `INSERT INTO url(url) VALUES($1);`
	res, e := tx.Exec(q, url)
	if e != nil {
		tx.Rollback()
		mlog.Warning("can not insert into url - %v\n", e)
	} else {
		tx.Commit()
		u.Id, _ = res.LastInsertId()
	}
	return u
}

func GetUrlCredByID(id int64) *UrlCred {
	DB := m.GetDB("")
	defer DB.Close()
	q := `SELECT
	id,
	cred_id,
	url_id,
	note,
	datelog,
	qrlink
	FROM url_cred WHERE id = $1
	`
	uc := UrlCred{}
	if e := DB.QueryRow(q, id).Scan(&uc.Id, &uc.Cred_id, &uc.Url_id, &uc.Note, &uc.Datelog, &uc.Qrlink); e != nil {
		mlog.Error(fmt.Errorf("can not get url-cred %v", e))
		return nil
	}
	uc.Update()
	return &uc
}

func GetUrlCred(cred_id, url_id int64) *UrlCred {
	DB := m.GetDB("")
	defer DB.Close()
	q := `SELECT
	id,
	cred_id,
	url_id,
	note,
	datelog,
	qrlink
	FROM url_cred WHERE cred_id = $1 AND url_id = $2
	`
	uc := UrlCred{}
	if e := DB.QueryRow(q, cred_id, url_id).Scan(&uc.Id, &uc.Cred_id, &uc.Url_id, &uc.Note, &uc.Datelog, &uc.Qrlink); e != nil {
		mlog.Warning("ERROR can not get url-cred %v\n", e)
		uc.Id, uc.Cred_id, uc.Url_id = int64(-1), cred_id, url_id
	}
	return &uc
}

func UrlCredNew(in map[string]interface{}) *UrlCred {
	Cred_id := u.MapLookup(in, "cred_id", int64(-1)).(int64)
	Url_id := u.MapLookup(in, "url_id", int64(-1)).(int64)

	uc := GetUrlCred(Cred_id, Url_id)
	uc.Note = u.MapLookup(in, "cred_note", "").(string)
	uc.Datelog = time.Now().UnixNano()
	uc.Qrlink = u.MapLookup(in, "qrlink", "").(string)
	uc.Save()
	uc.Update()
	return uc
}

func (uc *UrlCred) Update() {
	uc.Credential = GetCredentialByID(uc.Cred_id)
	uc.Url = GetUrlByID(uc.Url_id)
}

func (uc *UrlCred) Delete() {
	DB := m.GetDB("")
	defer DB.Close()
	q := `DELETE FROM url_cred WHERE id = $1`
	tx, _ := DB.Begin()
	_, e := tx.Exec(q, uc.Id)
	if e != nil {
		tx.Rollback()
		mlog.Warning("ERROR removing url_cred %d\n", uc.Id)
		return
	}
	tx.Commit()
	u := GetUrlByID(uc.Url_id)
	c := GetCredentialByID(uc.Cred_id)
	//These cleanup need to be sure there is no other links by themself
	u.Delete()
	c.Delete()
}

func (u *Url) Delete() {
	//Check if any ref in url_cred
	DB := m.GetDB("")
	defer DB.Close()
	var dummy int64
	if e := DB.QueryRow(`SELECT url_id FROM url_cred WHERE url_id = $1`, u.Id).Scan(&dummy); e != nil {
		tx, _ := DB.Begin()
		if _, e := tx.Exec(`DELETE FROM url WHERE id = $1`, u.Id); e != nil {
			tx.Rollback()
			mlog.Warning("ERROR can not remove url %v\n", e)
		} else {
			tx.Commit()
		}
	} else {
		mlog.Info("url not delete because in use\n")
	}
}

func (c *Credential) Delete() {
	DB := m.GetDB("")
	defer DB.Close()
	if e := DB.QueryRow(`SELECT cred_id FROM url_cred WHERE cred_id = $1`, c.Id).Scan(&c.Id); e != nil {
		tx, _ := DB.Begin()
		q := `DELETE FROM credential WHERE id = $1`
		if _, e := tx.Exec(q, c.Id); e != nil {
			mlog.Warning("ERROR can not remove cred %v\n", e)
		} else {
			tx.Commit()
		}
	} else {
		mlog.Info(" cred not delete because in use\n")
	}

}

func GetUrl(url string) *Url {
	DB := m.GetDB("")
	defer DB.Close()
	q := `SELECT
	id,
	url FROM url WHERE url = $1`
	u := Url{}
	if e := DB.QueryRow(q, url).Scan(&u.Id, &u.Url); e != nil {
		mlog.Warning("GetUrlByID %v\n", e)
		u.Url = url
	}
	return &u
}

func GetUrlByID(id int64) *Url {
	DB := m.GetDB("")
	defer DB.Close()
	q := `SELECT
	id,
	url FROM url WHERE id = $1`
	u := Url{}
	if e := DB.QueryRow(q, id).Scan(&u.Id, &u.Url); e != nil {
		mlog.Warning("GetUrlByID %v\n", e)
		return nil
	}
	return &u
}

func SearchCredentials(kw string, UserID int64) *[]UrlCred {
	DB := m.GetDB("")
	defer DB.Close()
	q := fmt.Sprintf(`SELECT
	u.id,
	u.cred_id,
	u.url_id,
	u.note,
	u.datelog,
	u.qrlink
	FROM url_cred as u, credential as c, url
	WHERE c.id == u.cred_id
	AND url.id = u.url_id
	AND c.user_id = %d
	AND ((url.url LIKE "%%%s%%")
		OR (u.note LIKE "%%%s%%")
	)
	`, UserID, kw, kw)
	// mlog.Warning("%v", q)
	rows, err := DB.Query(q, UserID, kw)
	if err != nil {
		mlog.Warning("Error searching url-cred - %v\n", err)
		return nil
	}
	o := []UrlCred{}
	for rows.Next() {
		UrlCred := UrlCred{}
		if err := rows.Scan(
			&UrlCred.Id, &UrlCred.Cred_id, &UrlCred.Url_id, &UrlCred.Note, &UrlCred.Datelog, &UrlCred.Qrlink,
		); err != nil {
			mlog.Warning("Error searching url-cred - %v\n", err)
			return &o
		}
		UrlCred.Update()
		o = append(o, UrlCred)
	}
	return &o
}

func GetCredential(userID int64, credUserName, credPassword string) *Credential {
	DB := m.GetDB("")
	defer DB.Close()
	q := `SELECT
	id,
	user_id,
	cred_username,
	cred_password
	FROM credential WHERE user_id = $1 AND cred_username = $2 AND cred_password = $3`
	cred := Credential{}
	if err := DB.QueryRow(q, userID, credUserName, credPassword).Scan(&cred.Id, &cred.User_id, &cred.Cred_username, &cred.Cred_password); err != nil {
		mlog.Warning("Can not get cred from db.  %v\n", err)
		cred.Id = -1
		cred.User_id = userID
		cred.Cred_username = credUserName
		cred.Cred_password = credPassword
	}
	return &cred
}

func GetCredentialByID(id int64) *Credential {
	DB := m.GetDB("")
	defer DB.Close()
	q := `SELECT
	id,
	user_id,
	cred_username,
	cred_password
	FROM credential WHERE id = $1`
	cred := Credential{}
	if err := DB.QueryRow(q, id).Scan(&cred.Id, &cred.User_id, &cred.Cred_username, &cred.Cred_password); err != nil {
		mlog.Warning("Can not get cred with this ID %d, %v\n", id, err)
		return nil
	}
	return &cred
}

func CredentialNew(in map[string]interface{}) *Credential {
	userID := u.MapLookup(in, "user_id", int64(0)).(int64)
	username := u.MapLookup(in, "cred_username", "").(string)
	password := u.MapLookup(in, "cred_password", "").(string)
	c := GetCredential(userID, username, password)
	c.Save()
	return c
}

// Save -
// Always add new row or dont do anything. We need some sql to remove dangling cred later on
func (c *Credential) Save() {
	DB := m.GetDB("")
	defer DB.Close()
	tx, _ := DB.Begin()
	q := `INSERT INTO credential (
		user_id,
		cred_username,
		cred_password ) VALUES ($1, $2, $3)
		`
	res, e := tx.Exec(q, c.User_id, c.Cred_username, c.Cred_password)
	if e != nil {
		tx.Rollback()
		mlog.Warning("save Cred %v\n", e)
	} else {
		tx.Commit()
		c.Id, _ = res.LastInsertId()
	}
}
