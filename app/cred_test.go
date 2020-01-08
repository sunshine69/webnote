package app

import (
	"log"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
	"os"
	"testing"
)

func TestSetupCredSchema(t *testing.T) {
	os.Setenv("DBPATH", "/home/stevek/src/webnote-go/testwebnote.db")
	SetupSchema()
}

func TestCredModel(t *testing.T) {
	os.Setenv("DBPATH", "/home/stevek/src/webnote-go/testwebnote.db")
	//Test to create one cred url pair
	u := UrlNew("https://note.inxuanthuy.com/")
	fmt.Printf("u: %v\n", *u)
	c := CredentialNew(map[string]interface{}{
		"user_id": int64(1),
		"cred_username": "mylogin",
		"cred_password": "somepassword1",
	} )
	fmt.Printf("c: %v\n", *c)
	uc := UrlCredNew(map[string]interface{}{
		"cred_id": c.Id,
		"url_id": u.Id,
		"cred_note": "Test credential only first",
		"qrlink": "none",
	})
	fmt.Printf("uc: %v\n", *uc)
	//Test Search cred
	ucs := SearchCredentials("inxuanthuy", int64(1))
	fmt.Printf("ucs: %v\n", *ucs)
	uc.Delete()
	u.Delete()
	c.Delete()
}

func TestMigrateCred(t *testing.T) {
	os.Setenv("DBPATH", "/home/stevek/src/webnote-go/testwebnote.db")
	DB1, _ := sql.Open("sqlite3", "file:///home/stevek/webnote.sqlite3")
	q := `SELECT
	uc.cred_id,
	uc.url_id,
	uc.note,
	uc.datelog,
	uc.qrlink,
	c.cred_username,
	c.cred_password,
	u.url
	FROM webnote_urlcredential as uc, webnote_credential as c, webnote_url as u
	WHERE uc.cred_id = c.id AND uc.url_id = u.id
`
	// fmt.Println(q)
	rows, e := DB1.Query(q)
	if e != nil {
		log.Fatalf("ERROR query %v\n", e)
	}
	defer rows.Close()
	for rows.Next() {
		var cred_id, url_id int64
		var note, qrlink, username, password, url, datelog sql.NullString
		if e := rows.Scan(&cred_id, &url_id, &note, &datelog, &qrlink, &username, &password, &url); e != nil {
			log.Fatalf("error scaning %v\n", e)
		}
		// log.Printf("DEBUG cred_id '%d' url_id '%d' datelog '%s', note: '%s', qrlink: '%s', uname: '%s', pass: '%s', url: '%s'\n", cred_id, url_id, datelog, note, qrlink, username, password, url )
		if ! url.Valid || !username.Valid || !password.Valid { continue }
		u := UrlNew(url.String)
		c := CredentialNew(map[string]interface{}{
			"user_id": int64(1),
			"cred_username": username.String,
			"cred_password": password.String,
		})
		uc := UrlCredNew(map[string]interface{}{
			"cred_id": c.Id,
			"url_id": u.Id,
			"cred_note": note.String,
			"qrlink": qrlink.String,
		})
		fmt.Printf("Migrated %v\n", uc)
	}
}