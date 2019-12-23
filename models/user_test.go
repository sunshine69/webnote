package models

import (
	"log"
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
	fmt.Printf("User ID: %d - Email: %s\n", u.ID, u.Email)
	u1 := GetUser("msh.computing@gmail.com")
	u2 := GetUserByID(u1.ID)
	log.Printf("Get user by id: %v\n", u2)
	u1.SaltLength = 16
	u1.SetUserPassword("1qa2ws")
	if ! VerifyHash("1qa2ws", u1.PasswordHash, int(u1.SaltLength)) {
		t.Fatalf("ERROR VerifyHash\n")
	}
	log.Println(VerifyLogin("msh.computing@gmail.com", "1qa2ws", ""))
}