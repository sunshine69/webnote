package models

import (
	"log"
)

type User struct {
	ID int64
	FirstName string
	LastName string
	Email string
	Address string
	PasswordHash string
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

	ct, ok := in["FirstName"].(string)
	if !ok {
		ct = ""
	}
	n.FirstName = ct

	ct, ok = in["LastName"].(string)
	if !ok {
		ct = ""
	}
	n.LastName = ct

	ct, ok = in["Email"].(string)
	if !ok {
		ct = ""
	}
	n.Email = ct

	ct, ok = in["Address"].(string)
	if !ok {
		ct = ""
	}
	n.Address = ct

	ct, ok = in["HomePhone"].(string)
	if !ok {
		ct = ""
	}
	n.HomePhone = ct

	ct, ok = in["WorkPhone"].(string)
	if !ok {
		ct = ""
	}
	n.WorkPhone = ct

	ct, ok = in["MobilePhone"].(string)
	if !ok {
		ct = ""
	}
	n.MobilePhone = ct

	ct, ok = in["ExtraInfo"].(string)
	if !ok {
		ct = ""
	}
	n.ExtraInfo = ct

	ct, ok = in["Address"].(string)
	if !ok {
		ct = ""
	}
	n.Address = ct

	ct, ok = in["TotpPassword"].(string)
	if !ok {
		ct = ""
	}
	n.TotpPassword = ct

	return &n
}

//Save a User. If new User then create on. If existing User then create a revisions before update.
func (n *User) Save() {
	var UserID int64
	currentUser := GetUser(n.Email)
	DB := GetDB("")
	defer DB.Close()
	var sql string

	log.Println(currentUser)

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
			totp_passwd) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);`
		res, e := tx.Exec(sql, n.FirstName, n.LastName, n.Email, n.Address, n.PasswordHash, n.HomePhone, n.WorkPhone, n.MobilePhone, n.ExtraInfo, n.LastAttempt, n.AttemptCount, n.LastLogin, n.PrefID, n.TotpPassword)
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
			totp_passwd = $13
			WHERE email = $14`
		_, e := tx.Exec(sql, n.FirstName, n.LastName, n.Address, n.PasswordHash, n.HomePhone, n.WorkPhone, n.MobilePhone, n.ExtraInfo, n.LastAttempt, n.AttemptCount, n.LastLogin, n.PrefID, n.TotpPassword, n.Email)
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
		totp_passwd
		FROM user WHERE email = $1`, email).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Address, &u.PasswordHash, &u.HomePhone, &u.WorkPhone, &u.MobilePhone, &u.ExtraInfo, &u.LastAttempt, &u.AttemptCount, &u.LastLogin, &u.PrefID, &u.TotpPassword); e != nil {
		log.Printf("INFO - Can not find user email '%s' - %v\n", email, e)
		return nil
	}
	return &u
}

func (n *User) String() string {return n.FirstName + " " + n.LastName}