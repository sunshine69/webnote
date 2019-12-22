package models

import (
	"os"
	"log"
	"testing"
)

func TestConfig(t *testing.T) {
	os.Remove("testwebnote.db")
	os.Setenv("DBPATH", "testwebnote.db")
	SetupDefaultConfig()
	log.Println(GetConfig("list_flags"))
	log.Println(GetConfig("date_layout"))
	SetConfig("testkey", "Test value")
	SetConfig("testkey", "Test value again")
	log.Printf("Assert testkey %v\n", (GetConfig("testkey") == "Test value again"))
	DeleteConfig("testkey")
	log.Printf("Assert testkey == 'default test key' %v\n", GetConfig("testkey", "default test key") == "default test key")
	v := GetConfigSave("testkey1", "test value 1")
	log.Println(v)
	log.Println(GetConfig("testkey1"))
	SetupAppDatabase()
}