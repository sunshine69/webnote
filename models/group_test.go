package models

import (
	"log"
	"os"
	"testing"
)
func TestMisc(t *testing.T) {
	os.Setenv("DBPATH", "testwebnote.db")
	// SetupDefaultConfig()
	// SetupAppDatabase()
	g := Group{
		Name: "newgroup",
	}
	g.Save()
	log.Println(g)
}
