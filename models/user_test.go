package models

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestUser(t *testing.T) {
	os.Remove("testwebnote.db")
	os.Setenv("DBPATH", "testwebnote.db")
	SetupDefaultConfig()
	SetupAppDatabase()
	u := User{
		FirstName: "Steve",
		LastName:  "Kieu",
		Email:     "msh.computing@gmail.com",
	}
	fmt.Println(u)
	u.Save()
	fmt.Printf("User ID: %d - Email: %s\n", u.ID, u.Email)
	u1 := GetUser("msh.computing@gmail.com")
	u2 := GetUserByID(u1.ID)
	u2.SaltLength = 16
	u2.SetUserPassword("1qa2ws")
	log.Printf("Get user by id: %v - s ln %d - pHash %s\n", u2, u2.SaltLength, u2.PasswordHash)
	if !VerifyHash("1qa2ws", u2.PasswordHash, int(u2.SaltLength)) {
		t.Fatalf("ERROR VerifyHash\n")
	}
	log.Println(VerifyLogin("msh.computing@gmail.com", "1qa2ws", "", "127.0.0.1"))
}

func TestChangeUserPassword(t *testing.T) {
	os.Setenv("DBPATH", "/home/stevek/src/webnote-go/testwebnote.db")
	u := GetUser("msh.computing@gmail.com")
	u.SetUserPassword("XXXXXXX")
}
