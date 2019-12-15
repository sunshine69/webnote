package models

import (
	"log"
	"os"
	"testing"
)

func TestNote(t *testing.T) {
	os.Setenv("DBPATH", "testwebnote.db")
	SetupDefaultConfig()
	SetupAppDatabase()
	aNote := NoteNew(map[string]interface{} {
		"title": "New note 1",
		"flags": ":TODO",
		"content": "Content note 1",
	})
	aNote.Save()

	mySavedNote := GetNote("New note 1")
	log.Printf("Saved note: %v\n", mySavedNote)
}