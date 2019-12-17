package models

import (
	"fmt"
	"os"
	"testing"
)

func TestUser(t *testing.T) {
	os.Setenv("DBPATH", "testwebnote.db")
	SetupDefaultConfig()
	SetupAppDatabase()
	u := UserNew(map[string]interface{} {
		"FirstName": "Steve",
		"LastName": "Kieu",
		"Email": "msh.computing@gmail.com",
	})
	fmt.Println(u)
	u.Save()
	fmt.Println(u.Email)
	u1 := GetUser("msh.computing@gmail.com")
	fmt.Println(u1)
}