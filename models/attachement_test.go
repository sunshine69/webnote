package models

import (
	"log"
	"os"
	"testing"
)

func init() {
	os.Setenv("DBPATH", "test-webnote.db")
	SetupDefaultConfig()
	SetupAppDatabase()
}

func TestAttachement(t *testing.T) {

	a := Attachment{
		Name: "test attachment 1",
		AttachedFile: "/tmp/t",
	}
	a.GroupID = 1
	a.AuthorID = 1
	a.Save()
	a1 := GetAttachement("test attachment 1")
	log.Println(a1)
	a1.AttachedFile = "/tmp/t1"
	a1.Save()
	a2 := GetAttachement("test attachment 1")
	log.Print(a2.AttachedFile == "/tmp/t1" )
}

func TestScanAttachement(t *testing.T) {
	ScanAttachment("/home/stevek/tmp")
}