-- This is the sqlite3 version of sqlite3. Leave it here for reference.
CREATE TABLE IF NOT EXISTS "webnote_category" (
    "id" integer NOT NULL PRIMARY KEY,
    "name" varchar(254) NOT NULL UNIQUE,
    "desc" text
);
CREATE TABLE IF NOT EXISTS "webnote_comment" (
    "id" integer NOT NULL PRIMARY KEY,
    "user_id" integer NOT NULL,
    "note_id" integer NOT NULL,
    "datelog" integer UNIQUE,
    "content" varchar(1536)
);
CREATE TABLE IF NOT EXISTS "webnote_deletednotes" (
    "id" integer NOT NULL PRIMARY KEY,
    "note_id" integer,
    "title" varchar(254) UNIQUE,
    "datelog" integer UNIQUE
);
CREATE TABLE IF NOT EXISTS "webnote_image" (
    "id" integer NOT NULL PRIMARY KEY,
    "name" varchar(512),
    "path" varchar(512),
    UNIQUE ("path", "name")
);
CREATE TABLE IF NOT EXISTS "webnote_group" (
    "id" integer NOT NULL PRIMARY KEY,
    "name" varchar(50) NOT NULL UNIQUE,
    "description" varchar(256)
);
CREATE TABLE IF NOT EXISTS "webnote_note" (
    "id" integer NOT NULL PRIMARY KEY,
    "title" varchar(254) NOT NULL,
    "datelog" integer,
    "content" text,
    "url" varchar(356),
    "reminder_ticks" integer,
    "flags" varchar(25),
    "timestamp" integer,
    "econtent" varchar(1536),
    "alert_count" integer,
    "time_spent" integer,
    "author_id" integer NOT NULL,
    "group_id" integer NOT NULL REFERENCES "webnote_group" ("id"),
    "permission" integer NOT NULL, raw_editor integer default 0,
    UNIQUE ("title", "datelog")
);
CREATE TABLE IF NOT EXISTS "webnote_noteattachment" (
    "id" integer NOT NULL PRIMARY KEY,
    "note_id" integer NOT NULL REFERENCES "webnote_note" ("id"),
    "attachment_id" integer NOT NULL REFERENCES "webnote_attachment" ("id"),
    "user_id" integer NOT NULL,
    "timestamp" integer NOT NULL,
    UNIQUE ("note_id", "attachment_id")
);
CREATE TABLE IF NOT EXISTS "webnote_notecat" (
    "id" integer NOT NULL PRIMARY KEY,
    "note_id" integer NOT NULL REFERENCES "webnote_note" ("id"),
    "cat_id" integer NOT NULL REFERENCES "webnote_category" ("id"),
    UNIQUE ("note_id", "cat_id")
);
CREATE TABLE IF NOT EXISTS "webnote_noteimage" (
    "id" integer NOT NULL PRIMARY KEY,
    "note_id" integer NOT NULL REFERENCES "webnote_note" ("id"),
    "image_id" integer NOT NULL REFERENCES "webnote_image" ("id"),
    UNIQUE ("note_id", "image_id")
);
CREATE TABLE IF NOT EXISTS "webnote_notelink" (
    "id" integer NOT NULL PRIMARY KEY,
    "note_id" integer NOT NULL REFERENCES "webnote_note" ("id"),
    "link_note_id" integer NOT NULL REFERENCES "webnote_note" ("id"),
    "link_info" varchar(1536) NOT NULL,
    UNIQUE ("note_id", "link_note_id")
);
CREATE TABLE IF NOT EXISTS "webnote_preference" (
    "id" integer NOT NULL PRIMARY KEY,
    "tinymce_init" text
);
CREATE TABLE IF NOT EXISTS "webnote_user" (
    "id" integer NOT NULL PRIMARY KEY,
    "f_name" varchar(512),
    "l_name" varchar(512),
    "email" varchar(96) NOT NULL UNIQUE,
    "address" varchar(1024),
    "passwd" text,
    "h_phone" varchar(128),
    "w_phone" varchar(128),
    "m_phone" varchar(128),
    "extra_info" varchar(1536),
    "last_attempt" integer NOT NULL,
    "attempt_count" integer NOT NULL,
    "last_login" integer,
    "pref_id" integer NOT NULL REFERENCES "webnote_preference" ("id")
, totp_passwd text);
CREATE TABLE IF NOT EXISTS "webnote_usercat" (
    "id" integer NOT NULL PRIMARY KEY,
    "user_id" integer NOT NULL REFERENCES "webnote_user" ("id"),
    "cat_id" integer NOT NULL REFERENCES "webnote_category" ("id"),
    UNIQUE ("user_id", "cat_id")
);
CREATE TABLE IF NOT EXISTS "webnote_usergroup" (
    "id" integer NOT NULL PRIMARY KEY,
    "user_id" integer NOT NULL REFERENCES "webnote_user" ("id"),
    "group_id" integer NOT NULL REFERENCES "webnote_group" ("id"),
    UNIQUE ("user_id", "group_id")
);
CREATE TABLE IF NOT EXISTS "webnote_userpref" (
    "id" integer NOT NULL PRIMARY KEY,
    "user_id" integer NOT NULL REFERENCES "webnote_user" ("id"),
    "pref_id" integer NOT NULL REFERENCES "webnote_preference" ("id"),
    "name" varchar(384) NOT NULL,
    "desc" text NOT NULL,
    UNIQUE ("user_id", "pref_id", "name")
);
CREATE TABLE webnote_attachment("id" integer NOT NULL PRIMARY KEY,"name" varchar(768),"desc" varchar(1536),"author_id" integer NOT NULL,"group_id" integer NOT NULL,"permission" integer NOT NULL, "attached_file" varchar(100),"mimetype" varchar(64),"created" datetime,"updated" datetime);
CREATE TABLE IF NOT EXISTS "webnote_andrewaccount" (
    "id" integer NOT NULL PRIMARY KEY,
    "datelog" datetime NOT NULL,
    "description" varchar(512),
    "amount" decimal NOT NULL
);
CREATE TABLE IF NOT EXISTS "webnote_credential" (
    "id" integer NOT NULL PRIMARY KEY,
    "user_id" integer NOT NULL REFERENCES "webnote_user" ("id"),
    "cred_username" varchar(1024) NOT NULL,
    "cred_password" varchar(1024) NOT NULL,
    UNIQUE ("user_id", "cred_username", "cred_password")
);
CREATE TABLE IF NOT EXISTS "webnote_url" (
    "id" integer NOT NULL PRIMARY KEY,
    "url" varchar(4096) NOT NULL UNIQUE
);
CREATE TABLE IF NOT EXISTS "webnote_urlcredential" (
    "id" integer NOT NULL PRIMARY KEY,
    "cred_id" integer NOT NULL REFERENCES "webnote_credential" ("id"),
    "url_id" integer NOT NULL REFERENCES "webnote_url" ("id"),
    "note" varchar(1024),
    "datelog" datetime, qrlink varchar(1014),
    UNIQUE ("cred_id", "url_id")
);
