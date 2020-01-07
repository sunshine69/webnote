package app

import (
	"os"
	"testing"
)

func TestSetupCredSchema(t *testing.T) {
	os.Setenv("DBPATH", "/home/stevek/src/webnote-go/testwebnote.db")
	SetupSchema()
}