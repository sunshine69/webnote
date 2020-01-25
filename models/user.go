package models

import (
	"strings"
	"strconv"
	"log"
	"github.com/pquerna/otp/totp"
)

type User struct {
	ID int64
	GroupNames string //coma sep group names
	Groups []*Group
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
	AttemptCount int8
	LastLogin int64
	PrefID int8
	TotpPassword string
}
//Update -
func (u *User) Update() {
	for _, group := range( strings.Split(u.GroupNames, ",") ) {
		if group == "" { continue }
		u.SetGroup(group)
	}
	DB := GetDB(""); defer DB.Close()
	rows, e := DB.Query(`SELECT
		g.group_id,
		g.name,
		g.description
	FROM user_group AS ug, ngroup AS g
	WHERE ug.group_id = g.group_id
	AND ug.user_id = $1`, u.ID)
	if e != nil {
		log.Fatalf("ERROR %v\n", e)
	}
	defer rows.Close()

	for rows.Next() {
		gr := Group{}
		if e := rows.Scan(&gr.Group_id, &gr.Name, &gr.Description); e != nil {
			log.Fatalf("ERROR user Update. Can not query group - %v\n", e)
		}
		u.Groups = append(u.Groups, &gr)
	}
}

//SetGroup -
func (u *User) SetGroup(gnames ...string) {
	DB := GetDB(""); defer DB.Close()
	userID := u.ID
	for _, gname := range(gnames) {
		g := GetGroup(gname)
		if g != nil{
			if e := DB.QueryRow(`SELECT group_id FROM user_group WHERE user_id = $1 AND group_id = $2`, userID, g.Group_id).Scan(&g.Group_id); e != nil {
				log.Printf("INFO SetGroup can not get the group. Going to insert new one - %v\n", e)

				tx, _ := DB.Begin()
				res, e := tx.Exec(`INSERT INTO user_group(user_id, group_id) VALUES($1, $2)`, userID, g.Group_id)
				if e != nil {
					tx.Rollback()
					log.Fatalf("ERROR SetGroup can not set group to user - %v\n", e)
				}
				tx.Commit()

				id, _ := res.LastInsertId()
				log.Printf("INFO Insert one row to user_group - ID %d - , user ID %d\n", id, userID)
			}
		}
	}
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
	n.GroupNames = GetMapByKey(in, "GroupNames", "default").(string)

	return &n
}

//Save a User. If new User then create on. If existing User then create a revisions before update.
func (n *User) Save() {
	currentUser := GetUser(n.Email)
	DB := GetDB("")
	defer DB.Close()
	var sql string

	tx, _ := DB.Begin()
	if currentUser == nil {
		log.Printf("INFO New User %s\n", n.Email)
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
			salt_length) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);`
		res, e := tx.Exec(sql, n.FirstName, n.LastName, n.Email, n.Address, n.PasswordHash, n.HomePhone, n.WorkPhone, n.MobilePhone, n.ExtraInfo, n.LastAttempt, n.AttemptCount, n.LastLogin, n.PrefID, n.TotpPassword, n.SaltLength)
		if e != nil {
			tx.Rollback()
			log.Fatalf("ERROR can not insert user - %v\n", e)
		}
		n.ID, _ = res.LastInsertId()
	} else {
		log.Printf("INFO Update the User %s\n", n.Email)
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
			salt_length = $14
			WHERE email = $15`
		_, e := tx.Exec(sql, n.FirstName, n.LastName, n.Address, n.PasswordHash, n.HomePhone, n.WorkPhone, n.MobilePhone, n.ExtraInfo, n.LastAttempt, n.AttemptCount, n.LastLogin, n.PrefID, n.TotpPassword, n.SaltLength, n.Email)
		if e != nil {
			tx.Rollback()
			log.Fatalf("ERROR can not update user %v\n", e)
		}
	}
	tx.Commit()
	n.Update()
}

//GetUserByID -
func GetUserByID(id int64) (*User) {
	DB := GetDB("")
	defer DB.Close()
	u := User{ ID: id }
	if e := DB.QueryRow(`SELECT
		id as user_id,
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
		attempt_count,
		last_login,
		pref_id,
		totp_passwd,
		salt_length
		FROM user WHERE id = $1`, id).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Address, &u.PasswordHash, &u.HomePhone, &u.WorkPhone, &u.MobilePhone, &u.ExtraInfo, &u.LastAttempt, &u.AttemptCount, &u.LastLogin, &u.PrefID, &u.TotpPassword, &u.SaltLength); e != nil {
		log.Printf("INFO - Can not find user ID '%d' - %v\n", id, e)
		return nil
	}
	u.Update()
	return &u
}

func GetUser(email string) (*User) {
	DB := GetDB("")
	defer DB.Close()
	u := User{ Email: email }
	if e := DB.QueryRow(`SELECT
		id as user_id,
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
		attempt_count,
		last_login,
		pref_id,
		totp_passwd,
		salt_length
		FROM user WHERE email = $1`, email).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Address, &u.PasswordHash, &u.HomePhone, &u.WorkPhone, &u.MobilePhone, &u.ExtraInfo, &u.LastAttempt, &u.AttemptCount, &u.LastLogin, &u.PrefID, &u.TotpPassword, &u.SaltLength); e != nil {
		log.Printf("INFO - Can not find user email '%s' - %v\n", email, e)
		return nil
	}
	u.Update()
	return &u
}

func (n *User) String() string {
	if n.FirstName != "" || n.LastName != "" {
		return n.FirstName + " " + n.LastName
	} else {
		return n.Email
	}
}

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