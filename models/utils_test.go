package models

import (
	"os"
	"fmt"
	"log"
	"testing"
)

func TestVerifyPass(t *testing.T) {
	pass := MakePassword(32)
	salt := MakeSalt(16)
	hash := ComputeHash(pass, *salt)
	if ! VerifyHash(pass, hash, 16) {
		t.Fatalf("ERROR VerifyHash\n")
	}
}
func TestMakePass(t *testing.T) {
	t.Log("Start")
	p := MakePassword(32)
	log.Printf("PASSWORD: '%s'\n",p)
}

func TestCheckIP(t *testing.T) {
	log.Println(CheckUserIPInWhiteList("52.62.137.194", "192.168.0.0/24, 127.0.0.1/8, 52.62.137.194/32"))
}

func TestZipEncrypt(t *testing.T) {
	os.Chdir("../")
	tkey := ZipEncript("uploads/jenkins-DEPLOY-spc-r2r-43.log", "uploads/jenkins-DEPLOY-spc-r2r-43.zip" ,"1234")
	fmt.Printf("Key: '%s'\n", tkey)
}