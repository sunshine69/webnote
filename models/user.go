package models

import (
	"strconv"
	"log"
	"github.com/pquerna/otp/totp"
)

type User struct {
	ID int64
	FirstName string
	LastName string
	Email string
	Address string
	PasswordHash string
	SaltLength int8
	HomePhone string
	WorkPhone string
	MobilePhone string
	ExtraInfo string
	LastAttempt int64
	AttemptCount string
	LastLogin int64
	PrefID int64
	TotpPassword string
}

//UserNew -
func UserNew(in map[string]interface{}) (*User) {
	n := User{}

	n.FirstName = GetMapByKey(in, "FirstName", "").(string)
	n.LastName = GetMapByKey(in, "LastName", "").(string)
	n.Email = GetMapByKey(in, "Email", "").(string)
	n.Address = GetMapByKey(in, "Address", "").(string)
	n.HomePhone = GetMapByKey(in, "HomePhone", "").(string)
	n.WorkPhone = GetMapByKey(in, "WorkPhone", "").(string)
	n.MobilePhone = GetMapByKey(in, "MobilePhone", "").(string)
	n.ExtraInfo = GetMapByKey(in, "ExtraInfo", "").(string)
	n.Address = GetMapByKey(in, "Address", "").(string)
	n.TotpPassword = GetMapByKey(in, "TotpPassword", "").(string)
	defaultSaltLength, _ := strconv.Atoi(GetConfig("salt_length", "12"))
	n.SaltLength = GetMapByKey(in, "SaltLength", int8(defaultSaltLength)).(int8)

	return &n
}

//Save a User. If new User then create on. If existing User then create a revisions before update.
func (n *User) Save() {
	var UserID int64
	currentUser := GetUser(n.Email)
	DB := GetDB("")
	defer DB.Close()
	var sql string

	tx, _ := DB.Begin()
	if currentUser == nil {//New User
		sql = `INSERT INTO user(
			f_name,
			l_name,
			email,
			address,
			passwd,
			h_phone,
			w_phone,
			m_phone,
			extra_info,
			last_attempt,
			attempt_count ,
			last_login,
			pref_id,
			totp_passwd,
			salt_length) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, int8($15));`
		res, e := tx.Exec(sql, n.FirstName, n.LastName, n.Email, n.Address, n.PasswordHash, n.HomePhone, n.WorkPhone, n.MobilePhone, n.ExtraInfo, n.LastAttempt, n.AttemptCount, n.LastLogin, n.PrefID, n.TotpPassword, n.SaltLength)
		if e != nil {
			tx.Rollback()
			log.Fatalf("ERROR can not insert user - %v\n", e)
		}
		UserID, _ = res.LastInsertId()
	} else {
		//Update the User
		sql = `UPDATE user SET
			f_name = $1,
			l_name = $2,
			address = $3,
			passwd = $4,
			h_phone = $5,
			w_phone = $6,
			m_phone = $7,
			extra_info = $8,
			last_attempt = $9,
			attempt_count  = $10,
			last_login = $11,
			pref_id = $12,
			totp_passwd = $13,
			salt_length = int8($14)
			WHERE email = $15`
		_, e := tx.Exec(sql, n.FirstName, n.LastName, n.Address, n.PasswordHash, n.HomePhone, n.WorkPhone, n.MobilePhone, n.ExtraInfo, n.LastAttempt, n.AttemptCount, n.LastLogin, n.PrefID, n.TotpPassword, n.SaltLength, n.Email)
		if e != nil {
			tx.Rollback()
			log.Fatalf("ERROR can not update user %v\n", e)
		}
	}
	tx.Commit()
	n.ID = UserID
}

func GetUser(email string) (*User) {
	DB := GetDB("")
	defer DB.Close()
	u := User{ Email: email }
	if e := DB.QueryRow(`SELECT
		id() as user_id,
		f_name,
		l_name,
		address,
		passwd,
		h_phone,
		w_phone,
		m_phone,
		extra_info,
		last_attempt,
		attempt_count,
		last_login,
		pref_id,
		totp_passwd,
		salt_length,
		FROM user WHERE email = $1`, email).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Address, &u.PasswordHash, &u.HomePhone, &u.WorkPhone, &u.MobilePhone, &u.ExtraInfo, &u.LastAttempt, &u.AttemptCount, &u.LastLogin, &u.PrefID, &u.TotpPassword, &u.SaltLength); e != nil {
		log.Printf("INFO - Can not find user email '%s' - %v\n", email, e)
		return nil
	}
	return &u
}

func (n *User) String() string {return n.FirstName + " " + n.LastName}

//VerifyLogin -
func VerifyLogin(username, password, otp string) (*User) {
	user := GetUser(username)
	if user != nil {
		if user.SaltLength == 0 {
			saltLengthStr := GetConfigSave("salt_length", "12")
			saltLength, _ := strconv.Atoi(saltLengthStr)
			user.SaltLength = int8(saltLength)
		}
		if ! VerifyHash(password, user.PasswordHash, int(user.SaltLength)) {
			return nil
		}
		if otp != "" {
			if ! totp.Validate(otp, user.TotpPassword) { return nil }
		}
	} else {
		return nil
	}
	return user
}

func (u *User) SetUserPassword(p string) {
	salt := MakeSalt(u.SaltLength)
	u.PasswordHash = ComputeHash(p, *salt)
	u.Save()
}