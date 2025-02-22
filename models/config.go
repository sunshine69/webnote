package models

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/jbrodriguez/mlog"
	_ "github.com/mattn/go-sqlite3"
	"github.com/microcosm-cc/bluemonday"
	u "github.com/sunshine69/golang-tools/utils"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var markdown goldmark.Markdown

func init() {
	mlog.Start(mlog.LevelInfo, "webnote.log")
	markdown = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
}

const MaxUploadSizeInMemory = 4 * 1024 * 1024 // 4 MB
const MaxUploadSize = 4 * 1024 * 1024 * 1024

var UpLoadPath = "uploads/"

// DateLayout - global
var DateLayout string

// WebNotePassword
var WebNotePassword string

// WebNoteUser
var WebNoteUser string

var SqlSetup, DBPATH string

// SessionStore -
var SessionStore *sessions.CookieStore

// AppSettings -
type AppSettings struct {
	BASE_URL          string
	ADMIN_EMAIL       string
	UPLOAD_ITEM_COUNT int
}

// Settings -
var Settings *AppSettings

// PermissionsList -
var PermissionList *map[int8]string

// GetSessionVal -
func GetSessionVal(r *http.Request, k string, defaultVal interface{}) interface{} {
	ses, e := SessionStore.Get(r, "auth-session")
	if e != nil {
		mlog.Error(fmt.Errorf("can not get session - %s", e.Error()))
		return defaultVal
	}
	o := ses.Values[k]
	if o == nil && defaultVal != nil {
		o = defaultVal
	}
	return o
}

// SaveSessionVal -
func SaveSessionVal(r *http.Request, w *http.ResponseWriter, k string, defaultVal interface{}) {
	ses, e := SessionStore.Get(r, "auth-session")
	if e != nil {
		mlog.Error(fmt.Errorf("can not get session - %s", e.Error()))
	}
	ses.Values[k] = defaultVal
	ses.Save(r, *w)
}

// GetDB -
func GetDB(dbPath string) *sql.DB {
	if dbPath == "" {
		if DBPATH == "" {
			DBPATH = os.Getenv("DBPATH")
		}
		dbPath = DBPATH
	}
	DB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic("failed to connect database")
	}
	//_, err = DB.Exec(`PRAGMA page_size = 4096;
	//PRAGMA journal_mode=wal;
	//PRAGMA cache_size=5000;
	//PRAGMA busy_timeout = 15000`)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//DB.SetMaxOpenConns(1)

	return DB
}

// InitConfig - SetupDB. This is the initial point of config setup. Note init() does not work if it relies
// on DbConn as at the time the DBPATH is not yet available
func InitConfig() {
	DateLayout = GetConfig("date_layout")
	WebNoteUser = GetConfig("webnote_user")
	Settings = &AppSettings{
		BASE_URL:          GetConfig("base_url"),
		ADMIN_EMAIL:       GetConfig("admin_email"),
		UPLOAD_ITEM_COUNT: 6,
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

// CreateAdminUser -
func CreateAdminUser() {
	u := UserNew(map[string]interface{}{
		"FirstName":  "Admin",
		"LastName":   "Admin",
		"Email":      GetConfig("admin_email"),
		"Password":   "1qa2ws",
		"SaltLength": int8(16),
		"GroupNames": "default,family,friend",
	})
	mlog.Info("Create new user %v - id %d\n", u, u.ID)
}

// CreatePublicReadUser - Used when u need an user obejct and the object have the public read access
func CreatePublicReadUser() {
	u := UserNew(map[string]interface{}{
		"FirstName":  "Reader",
		"LastName":   "Public",
		"Email":      "public_read_user@gonote.com",
		"Password":   "1qa2ws",
		"SaltLength": int8(16),
		"GroupNames": "default",
	})
	mlog.Info("Create new user %v - id %d\n", u, u.ID)
}

// SetupDefaultConfig - Setup/reset default configuration set
func SetupDefaultConfig() {
	DB := GetDB("")
	defer DB.Close()
	tx, e := DB.Begin()
	if e != nil {
		mlog.FatalIfError(e)
	}
	sql := `DROP TABLE IF EXISTS appconfig;
	CREATE TABLE appconfig(
		key text,
		val text
	);
	CREATE UNIQUE INDEX IF NOT EXISTS appconfigkeyidx ON appconfig(key);
	`
	if _, e := tx.Exec(sql); e != nil {
		mlog.FatalIfError(e)
	}
	configSet := map[string]string{
		"config_created": "",
		"list_flags":     "TODO<|>IMPORTANT<|>URGENT",
		"date_layout":    "02-01-2006 15:04:05 MST",
		//note_revision to keep
		"revision_to_keep": "1000",
		"admin_email":      "admin@admin.com",
	}
	for key, val := range configSet {
		fmt.Printf("Inserting %s - %s\n", key, val)
		_, e := tx.Exec(`INSERT INTO appconfig(key, val) VALUES($1, $2)`, key, val)
		if e != nil {
			mlog.FatalIfError(e)
		}
	}
	if e := tx.Commit(); e != nil {
		mlog.FatalIfError(e)
	}
}

// GetConfig - by key and return value. Give second arg as default value.
func GetConfig(key ...string) string {
	DB := GetDB("")
	defer DB.Close()
	var val string
	if err := DB.QueryRow(`SELECT val FROM appconfig WHERE key = $1;`, key[0]).Scan(&val); err != nil {
		mlog.Info("key not found %v\n", err)
		argLen := len(key)
		if argLen > 1 {
			return key[1]
		} else {
			return ""
		}
	}
	return val
}

// GetConfigSave -
func GetConfigSave(key ...string) string {
	v := GetConfig(key...)
	if len(key) == 2 && v == key[1] {
		SetConfig(key[0], key[1])
	}
	return v
}

// SetConfig - Set a config key with value
func SetConfig(key, val string) {
	curVal := GetConfig(key, "NOTFOUND")
	DB := GetDB("")
	defer DB.Close()
	tx, _ := DB.Begin()
	if curVal != "NOTFOUND" { //Key exists, need update?
		if curVal != val {
			if _, e := tx.Exec(`UPDATE appconfig SET val = $1 WHERE key = $2`, val, key); e != nil {
				tx.Rollback()
				mlog.FatalIfError(e)
			}
		}
	} else { //Not exist, just do insert
		if _, e := tx.Exec(`INSERT INTO appconfig(key, val) VALUES($1, $2)`, key, val); e != nil {
			tx.Rollback()
			mlog.FatalIfError(e)
		}
	}
	tx.Commit()
}

// DeleteConfig - delete the config key
func DeleteConfig(key string) {
	DB := GetDB("")
	tx, _ := DB.Begin()
	if _, e := tx.Exec(`DELETE FROM appconfig WHERE key = $1`, key); e != nil {
		tx.Rollback()
		mlog.FatalIfError(e)
	}
	tx.Commit()
}

// SetupAppDatabase -
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
		"group_id" integer NOT NULL REFERENCES "ngroup" ("ID"),
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
		permission integer,
		"raw_editor" integer default 0,
		diff_text text
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
		file_size integer,
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
		name text,
		description text
	);
	CREATE UNIQUE INDEX IF NOT EXISTS groupidx ON ngroup(name);
	DELETE FROM ngroup;
	INSERT INTO ngroup(name, description) VALUES("default", "default group");
	INSERT INTO ngroup(name, description) VALUES("family", "family group");
	INSERT INTO ngroup(name, description) VALUES("friend", "friend group");

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
	var DB *sql.DB
	if os.Getenv("DB_STORAGE_ON_CIFS") == "true" {
		mlog.Info("DB_STORAGE_ON_CIFS is true trying to work around with CIFS")
		DB = GetDB("file:/tmp/tmp-gonote.sqlite3?cache=shared&mode=rwc&_journal_mode=WAL")
	} else {
		mlog.Info("DB_STORAGE_ON_CIFS is not true setup normal")
		DB = GetDB("")
	}

	tx, _ := DB.Begin()
	if _, e := tx.Exec(SqlSetup); e != nil {
		mlog.FatalIfError(fmt.Errorf("can not setup app db - %s", e.Error()))
		tx.Rollback()
	}
	tx.Commit()
	DB.Close()
	if os.Getenv("DB_STORAGE_ON_CIFS") == "true" {
		source, err := os.Open("/tmp/tmp-gonote.sqlite3")
		if err != nil {
			mlog.FatalIfError(err)
		}
		defer source.Close()

		destination, err := os.Create(os.Getenv("DBPATH"))
		if err != nil {
			mlog.FatalIfError(err)
		}
		defer destination.Close()
		nBytes, err := io.Copy(destination, source)
		if err != nil {
			mlog.FatalIfError(err)
		}
		mlog.Info("Copy %d bytes", nBytes)
		os.Remove("/tmp/tmp-gonote.sqlite3")
	}
}

// CheckPerm - Check permission to do an operation on a object
// obj must have fields :  Permission, AuthorID/Author,  GroupID/Group (similar to a note)
// Action can be a string of 'r' (read), 'w' (write), 'rw' (read-write), 'd' (delete)
func CheckPerm(obj Object, UserID int64, Action string) bool {
	if obj.Permission == 5 { //World read, everyone logged in can do anything
		if Action == "r" {
			return true
		} else if UserID > 0 {
			return true
		}
		return false
	}

	if UserID == 0 {
		return false
	} //From here we require a logged in

	user := GetUserByID(UserID)
	if obj.AuthorID == user.ID {
		return true
	} //Object created by this userID can do all

	//From now user is not the owner of the object
	if Action == "d" {
		return false
	} //Only owner can delete object

	if obj.Permission == 4 {
		return true
	} //Logged in user can do anything except deletion

	groupIDMap := make(map[int64]string)
	for _, g := range user.Groups {
		groupIDMap[g.ID] = g.Name
	}
	if _, ok := groupIDMap[obj.GroupID]; !ok {
		//user has no group which matches with this object group
		if obj.Permission == 3 { //Group w, all read
			if Action == "r" {
				return true
			}
		}
		return false
	}
	//From now on user has a group that this object is in
	if obj.Permission >= 2 {
		return true
	} // group rw granted

	if obj.Permission == 0 {
		return false
	} //Only owner can do
	//Only left Permission == 1
	if Action == "r" {
		return true
	}
	return false
}

// TemplateFuncMap - custom template func map
var TemplateFuncMap *template.FuncMap
var AllTemplates *template.Template

func LoadAllTemplates() {
	//Template custom functions
	_TemplateFuncMap := u.GoTemplateFuncMap
	_TemplateFuncMap["md2html"] = func(md string) template.HTML {
		var buf bytes.Buffer
		if err := markdown.Convert([]byte(md), &buf); err != nil {
			panic(err)
		}
		cleanupBytes := bluemonday.UGCPolicy().SanitizeBytes(buf.Bytes())
		return template.HTML(cleanupBytes)
	}

	TemplateFuncMap = &_TemplateFuncMap
	t, err := template.New("templ").Funcs(*TemplateFuncMap).ParseGlob("assets/templates/*.html")
	if err != nil {
		mlog.FatalIfError(fmt.Errorf("can not parse templates %s", err.Error()))
	}
	AllTemplates = t
}
