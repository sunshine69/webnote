package app

import (
	"fmt"
	"regexp"
	"net/http"

	m "github.com/sunshine69/webnote-go/models"
)

func SaveBookMark(w http.ResponseWriter, r *http.Request) {
	bmarkdNote := m.GetNote("Bookmarks")
	myurl := m.GetRequestValue(r, "url", "")
	is_ajax := m.GetRequestValue(r, "is_ajax", "0")
	if myurl != "" {
		ptn := regexp.MustCompile(`(\<li\>[=]+ Form [=]+\<\/li\>)`)
		newText := fmt.Sprintf("<li><a href=\"%s\">%s</a></li><a href=\"/delbookmark?url=%s\">remove</a>\n<br/>$1", myurl, myurl, myurl)
		newCt := ptn.ReplaceAllString(bmarkdNote.Content, newText)
		bmarkdNote.Content = newCt
		bmarkdNote.Save()
	}
	if is_ajax != "1" {
		http.Redirect(w, r, fmt.Sprintf("/view/?id=%d&t=2", bmarkdNote.ID), http.StatusFound)
	}
	return
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
	return
}