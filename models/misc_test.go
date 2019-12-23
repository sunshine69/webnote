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
		Group_id: int8(4),
		Name: "newgroup",
	}
	g.Save()
	log.Println(g)
}
