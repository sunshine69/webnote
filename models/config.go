package models

import (
	"net/http"
	"github.com/gorilla/sessions"
	"modernc.org/ql"
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
	// fmt.Printf("Use dbpath %v\n", dbPath)
	ql.RegisterDriver2()
	// ql.Options {
	// 	RemoveEmptyWAL: true,
	// }

	DB, err := sql.Open("ql2", dbPath)
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
		BASE_URL: "https://note.xvt.technology",
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
		key STRING,
		val STRING,
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
	BEGIN TRANSACTION;

	CREATE TABLE IF NOT EXISTS note (
		title STRING,
		datelog int64,
		flags STRING,
		content STRING,
		url STRING,
		reminder_ticks int64 default 0,
		timestamp int64 default 0,
		time_spent int64 default 0,
		author_id int64 default 0,
		group_id int8 default 0,
		permission int8,
		raw_editor int8
	);
	CREATE UNIQUE INDEX IF NOT EXISTS note_title_idx ON note(title);

	CREATE TABLE IF NOT EXISTS note_revision (
		note_id int64,
		timestamp int64,
		flags STRING,
		url STRING,
		content STRING,
		author_id int64,
		group_id int8,
		permission int8
	);
	CREATE UNIQUE INDEX IF NOT EXISTS note_revision_idx ON note_revision(note_id, timestamp);

	CREATE TABLE IF NOT EXISTS note_comment (
		user_id int64,
		note_id int64,
		datelog int64,
		comment STRING
	);

	CREATE TABLE IF NOT EXISTS note_attachment (
		user_id int64 ,
		note_id int64,
		attachment_id int64,
		timestamp int64
	);
	CREATE UNIQUE INDEX IF NOT EXISTS note_attachmentidx ON note_attachment(note_id, attachment_id);

	CREATE TABLE IF NOT EXISTS attachment(
		name STRING,
		description STRING,
		author_id int64 ,
		group_id int8 ,
		permission int8,
		attached_file STRING,
		mimetype STRING,
		created int64,
		updated int64
	);
	CREATE UNIQUE INDEX IF NOT EXISTS attachmentidx ON attachment(name);

	CREATE TABLE IF NOT EXISTS user (
		f_name STRING,
		l_name STRING,
		email STRING,
		address STRING,
		passwd STRING,
		salt_length int8,
		h_phone STRING,
		w_phone STRING,
		m_phone STRING,
		extra_info STRING,
		last_attempt int64,
		attempt_count int8,
		last_login int64,
		pref_id int8 default 0,
		totp_passwd STRING
		);
	CREATE UNIQUE INDEX IF NOT EXISTS useremailidx ON user(email);

	CREATE TABLE IF NOT EXISTS ngroup (
		group_id int8,
		name STRING,
		description STRING,
	);
	CREATE UNIQUE INDEX IF NOT EXISTS groupidx ON ngroup(name);
	CREATE UNIQUE INDEX IF NOT EXISTS groupididx ON ngroup(group_id);
	DELETE FROM ngroup;
	INSERT INTO ngroup(group_id, name, description) VALUES(1, "default", "default group");
	INSERT INTO ngroup(group_id, name, description) VALUES(2, "family", "family group");
	INSERT INTO ngroup(group_id, name, description) VALUES(3, "friend", "friend group");

	CREATE TABLE IF NOT EXISTS user_group (
		user_id int64,
		group_id int8,
	);
	CREATE UNIQUE INDEX IF NOT EXISTS user_groupidx ON user_group(user_id, group_id);

-- End main application. Below is the extra app that the webnote per each sub app has
-- Andrew account ledger
	CREATE TABLE IF NOT EXISTS andrewaccount (
		datelog int64,
		description STRING,
		amount float64
	);
-- End Andrew account ledger

-- credential app
	CREATE TABLE IF NOT EXISTS credential (
		user_id int64,
		cred_username STRING ,
		cred_password STRING ,
	);
	CREATE UNIQUE INDEX IF NOT EXISTS credentialidx ON credential(user_id, cred_username, cred_password);

	CREATE TABLE IF NOT EXISTS url (
		url STRING
	);

	CREATE TABLE IF NOT EXISTS url_cred (
		cred_id int64  ,
		url_id int64  ,
		note STRING,
		datelog int64,
		qrlink STRING,
	);
	CREATE UNIQUE INDEX IF NOT EXISTS url_credidx ON url_cred(cred_id, url_id);
-- End credential app

	COMMIT;
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

// func CheckPerm(Obj interface{}, UserID int64, Action string) (bool) {
// 		//obj must have fields :  permission, author_id, group_id
// 		if obj.Permission.(string) == "5" {
// 			if Action == "r" {
// 				return true
// 			} else {
// 				if UserID {
// 					return true
// 				} else {return false}
// 			}
// 		}
// 		if ! UserID {return false}

		// user = User.objects.get(id=user_id)
		// if (obj.author_id == user.id): return True
		// if (action == u'd'): return False
		// if (obj.permission == 4): return True
		// if (obj.group_id not in [ug['id'] for ug in user.groups.values()]):
		// 	if  ((action == u'r') and (obj.permission == 3)): return True
		// 	else: return False
		// if (obj.permission >= 2): return True
		// if (obj.permission == 0): return False
		// if (action == u'r'): return True
		// else: return False
	// }