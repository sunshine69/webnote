package models

import (
	"log"
	"os"
	"testing"
)

func TestAttachement(t *testing.T) {
	os.Setenv("DBPATH", "testwebnote.db")
	SetupDefaultConfig()
	SetupAppDatabase()
	a := Attachment{
		Name: "test attachment 1",
		AttachedFile: "/tmp/t",
	}
	a.GroupID = 1
	a.AuthorID = 52
	a.Save()
	a1 := GetAttachement("test attachment 1")
	log.Println(a1)
	a1.AttachedFile = "/tmp/t1"
	a1.Save()
	a2 := GetAttachement("test attachment 1")
	log.Print(a2.AttachedFile == "/tmp/t1" )
}