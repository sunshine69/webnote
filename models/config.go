package models

import (
	"strings"
	"github.com/yuin/goldmark"
	"bytes"
	"github.com/microcosm-cc/bluemonday"
	"html/template"
	"net/http"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"database/sql"
	"os"
	"fmt"
)

//DateLayout - global
var DateLayout string

//WebNotePassword
var WebNotePassword string
//WebNoteUser
var WebNoteUser string

var SqlSetup, DBPATH string

//SessionStore -
var SessionStore *sessions.CookieStore

//AppSettings -
type AppSettings struct {
	BASE_URL string
	ADMIN_EMAIL string
}
//Settings -
var Settings *AppSettings

//PermissionsList -
var PermissionList *map[int8]string

//GetSessionVal -
func GetSessionVal(r *http.Request, k string, defaultVal interface{}) interface{} {
	ses, e := SessionStore.Get(r, "auth-session")
	if e != nil {
		log.Printf("ERROR can not get session - %v\n", e)
		return nil
	}
	o := ses.Values[k]
	if o == nil && defaultVal != nil {
		o = defaultVal
	}
	return o
}

//SaveSessionVal -
func SaveSessionVal(r *http.Request, w *http.ResponseWriter, k string, defaultVal interface{}) {
	ses, e := SessionStore.Get(r, "auth-session")
	if e != nil {
		log.Printf("ERROR can not get session - %v\n", e)
	}
	ses.Values[k] = defaultVal
	ses.Save(r, *w)
}

//GetDB -
func GetDB(dbPath string) (*sql.DB) {
	if dbPath == "" {
		if DBPATH == "" {
			DBPATH = os.Getenv("DBPATH")
		}
		dbPath = DBPATH
	}
	DB, err := sql.Open("sqlite3", "file://" + dbPath + "?_timeout=15")
	if err != nil {
	  panic("failed to connect database")
	}
	return DB
}

//InitConfig - SetupDB. This is the initial point of config setup. Note init() does not work if it relies
//on DbConn as at the time the DBPATH is not yet available
func InitConfig() {
	DateLayout = GetConfig("date_layout")
	WebNoteUser = GetConfig("webnote_user")
	Settings = &AppSettings{
		BASE_URL: "https://note.xvt.technology:8080",
		ADMIN_EMAIL: "msh.computing@gmail.com",
	}
	PermissionList = &map[int8]string{
		0: "only owner",
		1: "group read",
		2: "group rw",
		3: "group w, all read",
		4: "all rw",
		5: "world read, all rw",
	}
}

//CreateAdminUser -
func CreateAdminUser() {
	u := UserNew(map[string]interface{} {
		"FirstName": "Admin",
		"LastName": "Admin",
		"Email": "msh.computing@gmail.com",
	})
	u.Save()
	u.SaltLength = 16
	u.SetUserPassword("1qa2ws")
	log.Printf("DEBUG user object before calling SetGroup %v\n", *u)
	u.SetGroup("default", "family", "friend")
}

//SetupDefaultConfig - Setup/reset default configuration set
func SetupDefaultConfig() {
	DB := GetDB("")
	defer DB.Close()
	tx, e := DB.Begin()
	if e != nil {
		log.Fatalf("ERROR %v\n", e)
	}
	sql := `DROP TABLE IF EXISTS appconfig;
	CREATE TABLE appconfig(
		key text,
		val text
	);
	CREATE UNIQUE INDEX IF NOT EXISTS appconfigkeyidx ON appconfig(key);
	`
	if _, e := tx.Exec(sql); e != nil {
		log.Fatalf("ERROR %v\n", e)
	}
	configSet := map[string]string{
		"config_created": "",
		"list_flags" : "TODO<|>IMPORTANT<|>URGENT",
		"date_layout": "02-01-2006 15:04:05 MST",
		//note_revision to keep
		"revision_to_keep": "1000",
	}
	for key, val := range(configSet) {
		fmt.Printf("Inserting %s - %s\n", key, val)
		_, e := tx.Exec(`INSERT INTO appconfig(key, val) VALUES($1, $2)`, key, val)
		if e != nil {
			log.Fatalf("ERROR %v\n", e)
		}
	}
	if e := tx.Commit(); e != nil {
		log.Fatalf("ERROR %v\n", e)
	}
}

//GetConfig - by key and return value. Give second arg as default value.
func GetConfig(key ...string) (string) {
	DB := GetDB("")
	defer DB.Close()
	var val string
	if err := DB.QueryRow(`SELECT val FROM appconfig WHERE key = $1;`, key[0]).Scan(&val); err != nil {
		// log.Printf("INFO key not found %v\n", err)
		argLen := len(key)
		if argLen > 1 {
			return key[1]
		} else {
			return ""
		}
	}
	return val
}

//GetConfigSave -
func GetConfigSave(key ...string) (string) {
	v := GetConfig(key...)
	if len(key) == 2 && v == key[1] {
		SetConfig(key[0], key[1])
	}
	return v
}

//SetConfig - Set a config key with value
func SetConfig(key, val string) {
	curVal := GetConfig(key, "NOTFOUND")
	DB := GetDB("")
	defer DB.Close()
	tx, _ := DB.Begin()
	if curVal != "NOTFOUND" {//Key exists, need update?
		if curVal != val {
			if _, e := tx.Exec(`UPDATE appconfig SET val = $1 WHERE key = $2`, val, key); e != nil {
				tx.Rollback()
				log.Fatalf("ERROR %v\n", e)
			}
		}
	} else {//Not exist, just do insert
		if _, e := tx.Exec(`INSERT INTO appconfig(key, val) VALUES($1, $2)`, key, val); e != nil {
			tx.Rollback()
			log.Fatalf("ERROR %v\n", e)
		}
	}
	tx.Commit()
}

//DeleteConfig - delete the config key
func DeleteConfig(key string) {
	DB := GetDB("")
	tx, _ := DB.Begin()
	if _, e := tx.Exec(`DELETE FROM appconfig WHERE key = $1`, key); e !=nil {
		tx.Rollback()
		log.Fatalf("ERROR %v\n", e)
	}
	tx.Commit()
}

//SetupAppDatabase -
func SetupAppDatabase() {
	SqlSetup = `
	CREATE TABLE IF NOT EXISTS "note" (
		"id" integer NOT NULL PRIMARY KEY,
		"title" varchar(254) NOT NULL,
		"datelog" integer,
		"content" text,
		"url" varchar(356),
		"reminder_ticks" integer,
		"flags" varchar(512),
		"timestamp" integer,
		"time_spent" integer,
		"author_id" integer NOT NULL,
		"group_id" integer NOT NULL REFERENCES "ngroup" ("id"),
		"permission" integer NOT NULL,
		"raw_editor" integer default 0
	);
	CREATE UNIQUE INDEX IF NOT EXISTS note_title_idx ON note(title);

	CREATE TABLE IF NOT EXISTS note_revision (
		id integer NOT NULL PRIMARY KEY,
		note_id integer,
		timestamp integer,
		flags text,
		url text,
		content text,
		author_id integer,
		group_id integer,
		permission integer
	);
	CREATE UNIQUE INDEX IF NOT EXISTS note_revision_idx ON note_revision(note_id, timestamp);

	CREATE TABLE IF NOT EXISTS note_comment (
		id integer NOT NULL PRIMARY KEY,
		user_id integer,
		note_id integer,
		datelog integer,
		comment text
	);

	CREATE TABLE IF NOT EXISTS note_attachment (
		id integer NOT NULL PRIMARY KEY,
		user_id integer ,
		note_id integer,
		attachment_id integer,
		timestamp integer
	);
	CREATE UNIQUE INDEX IF NOT EXISTS note_attachmentidx ON note_attachment(note_id, attachment_id);

	CREATE TABLE IF NOT EXISTS attachment(
		id integer NOT NULL PRIMARY KEY,
		name text,
		description text,
		author_id integer ,
		group_id integer ,
		permission integer,
		attached_file text,
		mimetype text,
		created integer,
		updated integer
	);
	CREATE UNIQUE INDEX IF NOT EXISTS attachmentidx ON attachment(name);

	CREATE TABLE IF NOT EXISTS user (
		id integer NOT NULL PRIMARY KEY,
		f_name text,
		l_name text,
		email text,
		address text,
		passwd text,
		salt_length integer,
		h_phone text,
		w_phone text,
		m_phone text,
		extra_info text,
		last_attempt integer,
		attempt_count integer,
		last_login integer,
		pref_id integer default 0,
		totp_passwd text
		);
	CREATE UNIQUE INDEX IF NOT EXISTS useremailidx ON user(email);

	CREATE TABLE IF NOT EXISTS ngroup (
		id integer NOT NULL PRIMARY KEY,
		group_id integer,
		name text,
		description text
	);
	CREATE UNIQUE INDEX IF NOT EXISTS groupidx ON ngroup(name);
	CREATE UNIQUE INDEX IF NOT EXISTS groupididx ON ngroup(group_id);
	DELETE FROM ngroup;
	INSERT INTO ngroup(group_id, name, description) VALUES(1, "default", "default group");
	INSERT INTO ngroup(group_id, name, description) VALUES(2, "family", "family group");
	INSERT INTO ngroup(group_id, name, description) VALUES(3, "friend", "friend group");

	CREATE TABLE IF NOT EXISTS user_group (
		id integer NOT NULL PRIMARY KEY,
		user_id integer,
		group_id integer
	);
	CREATE UNIQUE INDEX IF NOT EXISTS user_groupidx ON user_group(user_id, group_id);

-- End main application. Below is the extra app that the webnote per each sub app has
-- Andrew account ledger
	CREATE TABLE IF NOT EXISTS andrewaccount (
		datelog integer,
		description text,
		amount float64
	);
-- End Andrew account ledger
	`
	DB := GetDB("")
	defer DB.Close()
	tx, _ := DB.Begin()
	if _, e := tx.Exec(SqlSetup); e != nil {
		log.Fatalf("ERROR can not setup app db - %v\n", e)
		tx.Rollback()
	}
	tx.Commit()
}

//CheckPerm - Check permission to do an operation on a object
//obj must have fields :  Permission, AuthorID/Author,  GroupID/Group (similar to a note)
//Action can be a string of 'r' (read), 'w' (write), 'rw' (read-write), 'd' (delete)
func CheckPerm(obj Object, UserID int64, Action string) (bool) {
		if obj.Permission == 5 {//World read, everyone logged in can do anything
			if Action == "r" {
				return true
			} else if UserID > 0 {
				return true
			}
			return false
		}

		if UserID == 0 { return false }//From here we require a logged in

		user := GetUserByID(UserID)
		if (obj.AuthorID == user.ID) { return true } //Object created by this userID can do all

		//From now user is not the owner of the object
		if (Action == "d") {return false} //Only owner can delete object

		if (obj.Permission == 4) {return true} //Logged in user can do anything except deletion

		var groupIDMap map[int8]string
		for _, g := range(user.Groups) {
			groupIDMap[g.Group_id] = g.Name
		}
		if _, ok := groupIDMap[obj.GroupID]; !ok {
			//user has no group which matches with this object group
			if (obj.Permission == 3) {//Group w, all read
				if (Action == "r") {
					return true
				}
				return false
			}
		}
		//From now on user has a group that this object is in
		if (obj.Permission >= 2) {return true}// group rw granted

		if (obj.Permission == 0) {return false}//Only owner can do
		//Only left Permission == 1
		if (Action == "r") {
			return true
		}
		return false
	}

//TemplateFuncMap - custom template func map
var TemplateFuncMap *template.FuncMap
var AllTemplates *template.Template

func LoadAllTemplates() {
	//Template custom functions
	_TemplateFuncMap := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"inc": func(i int) int {
			return i + 1
		},
		"add": func(x, y int) int {
			return x + y
		},
		"time_fmt": func(timelayout string, timeticks int64) string {
			return NsToTime(timeticks).Format(timelayout)
		},
		"raw_html": func(html string) template.HTML {
			cleanupBytes := bluemonday.UGCPolicy().SanitizeBytes([]byte(html))
			return template.HTML(cleanupBytes)
		},
		"unsafe_raw_html": func(html string) template.HTML {
			return template.HTML(html)
		},
		"if_ie": func() template.HTML {
			return template.HTML("<!--[if IE]>")
		},
		"end_if_ie": func() template.HTML {
			return template.HTML("<![endif]-->")
		},
		"truncatechars": func(length int, in string) template.HTML {
			return template.HTML(ChunkString(in, length)[0])
		},
		"cycle": func(idx int, vals ...string) template.HTML {
			_idx := idx % len(vals)
			return template.HTML(vals[_idx])
		},
		"md2html": func(md string) template.HTML {
			var buf bytes.Buffer
			if err := goldmark.Convert([]byte(md), &buf); err != nil {
				panic(err)
			}
			cleanupBytes := bluemonday.UGCPolicy().SanitizeBytes(buf.Bytes())
			return template.HTML( cleanupBytes )
		},
		"replace": func(old, new, data string) template.HTML {
			o := strings.ReplaceAll(data, old, new)
			return template.HTML(o)
		},
		"contains": func(subStr, data string) bool {
			return strings.Contains(data, subStr)
		},
	}
	TemplateFuncMap = &_TemplateFuncMap
	t, err := template.New("templ").Funcs(*TemplateFuncMap).ParseGlob("assets/templates/*.html")
	if err != nil {
		log.Fatalf("ERROR can not parse templates %v\n", err)
	}
	AllTemplates = t
}