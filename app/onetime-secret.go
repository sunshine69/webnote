package app

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	u "github.com/sunshine69/golang-tools/utils"
	m "github.com/sunshine69/webnote-go/models"
)

func GenerateOnetimeSecURL(w http.ResponseWriter, r *http.Request) {
	submit_type := m.GetRequestValue(r, "submit", "")
	base_url := m.GetRequestValue(r, "base_url", "")
	var secret string
	if submit_type == "submit_genpass" {
		length_str := m.GetRequestValue(r, "password_len", "12")
		password_len, err := strconv.Atoi(length_str)
		if u.CheckErrNonFatal(err, "GenerateOnetimeSecURL") != nil {
			fmt.Fprintf(w, "ERROR length should be a integer")
			return
		}
		secret = u.GenRandomString(password_len)
	} else {
		secret = m.GetRequestValue(r, "sec_content", "")
		fmt.Printf("DEBUG sec is %s\n", secret)
	}
	var anote *m.Note = nil
	var note_title string
	for {
		gen_number, _ := rand.Int(rand.Reader, big.NewInt(922337203685477580))
		note_title = fmt.Sprintf("%d", gen_number)
		// check if we already have a note with this title - if we do then loop to generate a new title; otherwise exit this loop
		anote = m.GetNote(note_title)
		if anote == nil {
			break
		}
	}
	secnote := m.NoteNew(map[string]interface{}{
		"content":    secret,
		"title":      note_title,
		"permission": int8(0),
	})
	secnote.Save()

	secURL := fmt.Sprintf("%s/nocsrf/onetimesec/%s", base_url, note_title)
	fmt.Fprintf(w, "<html><body><b>Secret link: </b><a href=\"%s\">%s</a><br/>Secret Value: <i>%s</i><br/><b>Create new one: <a href=%s/assets/media/html/onetime-secret.html>new link</a></body></html>", secURL, secURL, secret, base_url)
}

func GetOnetimeSecret(w http.ResponseWriter, r *http.Request) {
	note_id_str := r.PathValue("secret_id")

	if note_id_str == "-1" {
		fmt.Fprintf(w, "ERROR got -1 in id")
		return
	}
	note_sec := m.GetNote(note_id_str)
	if note_sec == nil {
		fmt.Fprintf(w, "ERROR")
		return
	}
	sec := note_sec.Content
	note_sec.Delete()
	fmt.Fprint(w, sec)
}
