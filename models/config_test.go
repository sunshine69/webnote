package models

import (
	"os"
	"log"
	"testing"
)

func TestConfig(t *testing.T) {
	os.Setenv("DBPATH", "testwebnote.db")
	SetupDefaultConfig()
	log.Println(GetConfig("list_flags"))
	log.Println(GetConfig("date_layout"))
	SetConfig("testkey", "Test value")
	log.Printf("Assert testkey %v\n", (GetConfig("testkey") == "Test value"))
	DeleteConfig("testkey")
	log.Printf("Assert testkey == 'default test key' %v\n", GetConfig("testkey", "default test key") == "default test key" )
	SetupAppDatabase()
}