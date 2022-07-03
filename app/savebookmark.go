package app

import (
	"fmt"
	"regexp"
	"net/http"

	m "github.com/sunshine69/webnote-go/models"
)
// Create a note with title `Bookmarks` and with the content is the file bookmark-note.html in the same folder.
// remember to replace the edit link with the note ID - it is like a normal note so you can see it int he url when you edit the note.

func SaveBookMark(w http.ResponseWriter, r *http.Request) {
	bmarkdNote := m.GetNote("Bookmarks")
	myurl := m.GetRequestValue(r, "url", "")
	mytitle := m.GetRequestValue(r, "title", "")
	is_ajax := m.GetRequestValue(r, "is_ajax", "0")
	if myurl != "" {
		//The marker text is the pattern
		ptn := regexp.MustCompile(`(\<li\>[=]+ Form [=]+\<\/li\>)`)
		newText := fmt.Sprintf("<li><a href=\"%s\" title=\"%s\">%s</a></li><a href=\"/delbookmark?url=%s\">remove</a>\n<br/>$1", myurl, mytitle,myurl, myurl)
		newCt := ptn.ReplaceAllString(bmarkdNote.Content, newText)
		bmarkdNote.Content = newCt
		bmarkdNote.Save()
	}
	if is_ajax != "1" {
		http.Redirect(w, r, fmt.Sprintf("/view/?id=%d&t=2", bmarkdNote.ID), http.StatusFound)
	}
}

func DeleteBookMark(w http.ResponseWriter, r *http.Request) {
	bmarkdNote := m.GetNote("Bookmarks")
	myurl := m.GetRequestValue(r, "url", "")
	fmt.Println(myurl)
	is_ajax := m.GetRequestValue(r, "is_ajax", "0")
	if myurl != "" {
		lineToRemovePtn := regexp.MustCompile( fmt.Sprintf(`.*%s.*`, regexp.QuoteMeta(myurl)) )
		newText := lineToRemovePtn.ReplaceAllString(bmarkdNote.Content, "")
		bmarkdNote.Content = newText
		bmarkdNote.Save()
	}
	if is_ajax != "1" {
		http.Redirect(w, r, fmt.Sprintf("/view/?id=%d&t=2", bmarkdNote.ID), http.StatusFound)
	}
}