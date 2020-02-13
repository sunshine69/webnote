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
//update - Only be called from the GetUserXXX which complete the user object with external objects, references etc.
func (u *User) update() {
	log.Printf("DEBUG GroupNames: %s\n", u.GroupNames)
	if u.GroupNames != "" {
		for _, group := range( strings.Split(u.GroupNames, ",") ) {
			if group == "" { continue }
			group = strings.TrimSpace(group)
			log.Printf("INFO add user %v to group %s\n", u, group)
			u.SetGroup(group)
		}
	}
	DB := GetDB(""); defer DB.Close()
	rows, e := DB.Query(`SELECT
		g.id,
		g.name,
		g.description
	FROM user_group AS ug, ngroup AS g
	WHERE ug.group_id = g.id
	AND ug.user_id = $1`, u.ID)
	if e != nil {
		log.Fatalf("ERROR %v\n", e)
	}
	defer rows.Close()
	var gNames []string
	for rows.Next() {
		gr := Group{}
		if e := rows.Scan(&gr.ID, &gr.Name, &gr.Description); e != nil {
			log.Fatalf("ERROR user update. Can not query group - %v\n", e)
		}
		u.Groups = append(u.Groups, &gr)
		gNames = append(gNames, gr.Name)
	}
	u.GroupNames = strings.Join(gNames, `,`)
}

//SetGroup -
func (u *User) SetGroup(gnames ...string) {
	DB := GetDB(""); defer DB.Close()
	userID := u.ID
	for _, gname := range(gnames) {
		g := GetGroup(gname)
		if g != nil{
			if e := DB.QueryRow(`SELECT group_id FROM user_group WHERE user_id = $1 AND group_id = $2`, userID, g.ID).Scan(&g.ID); e != nil {
				log.Printf("ERROR %v\n", e)
				log.Printf("INFO SetGroup can not get the group. Going to insert new one - %v\n", e)
				tx, _ := DB.Begin()
				res, e := tx.Exec(`INSERT INTO user_group(user_id, group_id) VALUES($1, $2)`, userID, g.ID)
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

//UserNew - It will Call Save
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
	// n.TotpPassword = GetMapByKey(in, "TotpPassword", "").(string)
	n.SaltLength = GetMapByKey(in, "SaltLength", int8(12)).(int8)
	n.GroupNames = GetMapByKey(in, "GroupNames", "default").(string)

	n.Save()
	Password := GetMapByKey(in, "Password", "").(string)
	if Password != "" { n.SetUserPassword(Password) }
	return &n
}

//Save a User. If new User then create on. If existing User then update.
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
			h_phone,
			w_phone,
			m_phone,
			extra_info,
			last_attempt,
			attempt_count,
			last_login,
			pref_id,
			salt_length,
			passwd,
			totp_passwd) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);`
		res, e := tx.Exec(sql, n.FirstName, n.LastName, n.Email, n.Address, n.HomePhone, n.WorkPhone, n.MobilePhone, n.ExtraInfo, n.LastAttempt, n.AttemptCount, n.LastLogin, n.PrefID, n.SaltLength, n.PasswordHash, n.TotpPassword)
		//There seems to be a race condition/bug in sqlite3 driver when dealing with type. If we set passwd not null default "", somehow at a stage golang sqlite see the next fields 'salt_length' after is of type text rather than integer and when scanning causing error.
		//We have to insert it in the passwd and all other empty string to avoid null.
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
			h_phone = $4,
			w_phone = $5,
			m_phone = $6,
			extra_info = $7,
			last_attempt = $8,
			attempt_count  = $9,
			last_login = $10,
			pref_id = $11,
			salt_length = $12
			WHERE email = $13`
		_, e := tx.Exec(sql, n.FirstName, n.LastName, n.Address, n.HomePhone, n.WorkPhone, n.MobilePhone, n.ExtraInfo, n.LastAttempt, n.AttemptCount, n.LastLogin, n.PrefID, n.SaltLength, n.Email)
		if e != nil {
			tx.Rollback()
			log.Fatalf("ERROR can not update user %v\n", e)
		}
		n.ID = currentUser.ID
	}
	tx.Commit()
	//Refresh user object from udpated db
	n.update()
}

//GetUserByID - always return an up-to-date user object in full, it will call .update() to update data not directly from database
func GetUserByID(id int64) (*User) {
	DB := GetDB("")
	defer DB.Close()
	u := User{ ID: id }
	if e := DB.QueryRow(`SELECT
		id,
		f_name,
		l_name,
		email,
		address,
		passwd,
		salt_length,
		totp_passwd,
		h_phone,
		w_phone,
		m_phone,
		extra_info,
		last_attempt,
		attempt_count,
		last_login,
		pref_id
		FROM user WHERE id = $1`, id).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Address, &u.PasswordHash, &u.SaltLength, &u.TotpPassword, &u.HomePhone, &u.WorkPhone, &u.MobilePhone, &u.ExtraInfo, &u.LastAttempt, &u.AttemptCount, &u.LastLogin, &u.PrefID); e != nil {
		log.Printf("INFO - Can not find user ID '%d' - %v\n", id, e)
		return nil
	}
	u.update()
	return &u
}

//GetUser - by email always return an up-to-date user object in full, it will call .update() to update data not directly from database
func GetUser(email string) (*User) {
	DB := GetDB("")
	defer DB.Close()
	u := User{ Email: email }
	if e := DB.QueryRow(`SELECT
		id,
		f_name,
		l_name,
		email,
		address,
		passwd,
		salt_length,
		totp_passwd,
		h_phone,
		w_phone,
		m_phone,
		extra_info,
		last_attempt,
		attempt_count,
		last_login,
		pref_id
		FROM user WHERE email = $1`, email).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Address, &u.PasswordHash, &u.SaltLength, &u.TotpPassword, &u.HomePhone, &u.WorkPhone, &u.MobilePhone, &u.ExtraInfo, &u.LastAttempt, &u.AttemptCount, &u.LastLogin, &u.PrefID); e != nil {
		log.Printf("INFO - Can not find user email '%s' - %v\n", email, e)
		return nil
	}
	u.update()
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
	PasswordHash := ComputeHash(p, *salt)
	DB := GetDB(""); defer DB.Close()
	tx, _ := DB.Begin()
	sql := `UPDATE user SET
			passwd = $1,
			salt_length = $2
			WHERE email = $3`
	_, e := tx.Exec(sql, PasswordHash, u.SaltLength, u.Email)
	if e != nil {
		tx.Rollback()
		log.Fatalf("ERROR SetUserPassword can not update user %v\n", e)
	}
	tx.Commit()
	u.PasswordHash = PasswordHash
}

func (u *User) SaveUserOTP() {
	DB := GetDB(""); defer DB.Close()
	tx, _ := DB.Begin()
	sql := `UPDATE user SET
			totp_passwd = $1
			WHERE email = $2`
		_, e := tx.Exec(sql, u.TotpPassword, u.Email)
	if e != nil {
		tx.Rollback()
		log.Fatalf("ERROR SaveUserOTP can not update user %v\n", e)
	}
	tx.Commit()
}