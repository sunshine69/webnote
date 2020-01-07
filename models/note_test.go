package models

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestNote(t *testing.T) {
	os.Setenv("DBPATH", "testwebnote.db")
	// SetupDefaultConfig()
	// SetupAppDatabase()
	aNote := NoteNew(map[string]interface{} {
		"title": "New note 1",
		"flags": ":TODO",
		"content": "Content note 1",
		"group_id": int8(3),
		"author_id": int64(1),
	})
	log.Printf("Note: %v\n", aNote)
	aNote.Save()
	mySavedNote := GetNote("New note 1")
	log.Printf("Title be the same: %v\n", mySavedNote.Title == aNote.Title)
	mySavedNote.Content = `New content with new version`
	mySavedNote.Save()
	mySavedNote = GetNote("New note 1")
	log.Printf("Saved note: %v\n", mySavedNote)
	ov := GetNoteRevision(mySavedNote.ID)
	fmt.Println(ov)
	nbyID := GetNoteByID(mySavedNote.ID)
	fmt.Printf("Note Author: %v\n", nbyID.Author )
	fmt.Printf("Note Group: %v\n", nbyID.Group )
}

func TestSearchNote(t *testing.T) {
	os.Setenv("DBPATH", "testwebnote.db")
	r := SearchNote("New")
	fmt.Println(r)
	r = SearchNote("kodi + -log & !date")
	fmt.Println(r)
}