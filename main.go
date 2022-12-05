package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "time/tzdata"

	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jbrodriguez/mlog"
	jsoniter "github.com/json-iterator/go"
	u "github.com/sunshine69/golang-tools/utils"
	"github.com/sunshine69/webnote-go/app"
	"github.com/sunshine69/webnote-go/models"
)

var version, ServerPort, SSLKey, SSLCert string
var EnableCompression *string

func init() {
	SSLKey = models.GetConfig("ssl_key", "")
	SSLCert = models.GetConfig("ssl_cert", "")
	mlog.Start(mlog.LevelInfo, "webnote.log")
}

func GetCurrentUser(w *http.ResponseWriter, r *http.Request) *models.User {
	isAuth := models.GetSessionVal(r, "authenticated", nil)
	if isAuth == nil || !isAuth.(bool) {
		return nil
	}
	useremail := models.GetSessionVal(r, "useremail", "").(string)
	if useremail == "" {
		return nil
	}
	return models.GetUser(useremail)
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	noteID, _ := strconv.ParseInt(u.GetRequestValue(r, "id", "0"), 10, 64)
	raw_editor, _ := strconv.Atoi(u.GetRequestValue(r, "raw_editor", "1"))

	var aNote *models.Note
	u := GetCurrentUser(&w, r)
	if noteID == 0 {
		aNote = models.NoteNew(map[string]interface{}{
			"ID":         noteID,
			"group_id":   u.Groups[0].ID,
			"raw_editor": int8(raw_editor),
		})
	} else {
		aNote = models.GetNoteByID(noteID)
	}
	CommonRenderTemplate("frontpage.html", &w, r, &map[string]interface{}{
		"title":            "Webnote - note " + aNote.Title,
		"page":             "frontpage",
		"msg":              "",
		"note":             aNote,
		"first_time_login": models.GetSessionVal(r, "first_time_login", "no").(string),
	})
}

func GetFirstnChar(text string, n int) (o string) {
	text = strings.TrimSpace(text)
	ptn0 := regexp.MustCompile(`\<[^\<]+\>`)
	text = ptn0.ReplaceAllString(text, "")
	ptn := regexp.MustCompile(`([^\n]+)\n`)
	text = strings.TrimSpace(text)
	o1 := ptn.FindString(text)
	o1 = strings.TrimSpace(o1)
	l := len(o1)
	if l > n {
		o = o1[0:n]
	} else {
		o = o1
	}
	o = strings.TrimSpace(o)
	return o
}
func DoSaveNote(w http.ResponseWriter, r *http.Request) {
	msg := "OK note saved"
	user := GetCurrentUser(&w, r)

	r.ParseForm()
	noteID, _ := strconv.ParseInt(u.GetRequestValue(r, "id", "0"), 10, 64)
	ngroup := models.GetGroup(u.GetFormValue(r, "ngroup", "default"))
	_permission, _ := strconv.Atoi(u.GetFormValue(r, "permission", "0"))
	permission := int8(_permission)

	_raw_editor, _ := strconv.Atoi(u.GetFormValue(r, "raw_editor", "0"))
	raw_editor := int8(_raw_editor)

	var aNote *models.Note
	content := r.FormValue("content")
	title := r.FormValue("title")
	if title == "" {
		title = GetFirstnChar(content, 128)
	}
	TimeStamp, err := strconv.ParseInt(u.GetFormValue(r, "timestamp", "0"), 10, 64)
	if u.CheckErrNonFatal(err, "ParseInt") != nil || TimeStamp == 0 {
		TimeStamp = time.Now().UnixNano()
	}

	if noteID == 0 { //New note created by current user
		aNote = models.NoteNew(map[string]interface{}{
			"title":      title,
			"datelog":    r.FormValue("datelog"),
			"flags":      r.FormValue("flags"),
			"content":    content,
			"url":        r.FormValue("url"),
			"raw_editor": raw_editor, //If checked return string 1, otherwise empty string
			"permission": permission,
			"author_id":  user.ID,
			"group_id":   ngroup.ID,
			"timestamp":  TimeStamp,
		},
		)
		aNote.Save()
	} else { //Existing note loaded. Need to check permmission
		aNote = models.GetNoteByID(noteID)
		if models.CheckPerm(aNote.Object, user.ID, "w") {
			aNote.Title = r.FormValue("title")
			aNote.Flags = r.FormValue("flags")
			aNote.Content = r.FormValue("content")
			aNote.URL = r.FormValue("url")
			aNote.RawEditor = raw_editor //If checked return string 1, otherwise empty string
			aNote.Permission = permission
			aNote.GroupID = ngroup.ID
			aNote.Timestamp = TimeStamp
			aNote.Save()
		} else {
			msg = "Permission denied."
		}
	}
	isAjax := u.GetRequestValue(r, "is_ajax", "0")
	if isAjax == "1" {
		fmt.Fprint(w, msg)
	} else {
		if msg != "OK note saved" {
			fmt.Fprint(w, msg)
		} else {
			http.Redirect(w, r, fmt.Sprintf("/?id=%d", aNote.ID), http.StatusFound)
		}
	}
}

func DoSearchNote(w http.ResponseWriter, r *http.Request) {
	keyword := u.GetRequestValue(r, "keyword", "")
	user := GetCurrentUser(&w, r)
	var attachments []*models.Attachment
	if user != nil {
		attachments = models.SearchAttachment(keyword, user)
	} else {
		attachments = []*models.Attachment{}
	}
	notes := models.SearchNote(keyword, GetCurrentUser(&w, r))
	CommonRenderTemplate("search_result.html", &w, r, &map[string]interface{}{
		"title":       "Webnote - Search result",
		"page":        "search_result",
		"msg":         "",
		"notes":       notes,
		"attachments": attachments,
	})
}

// DoViewNote -
func DoViewNote(w http.ResponseWriter, r *http.Request) {
	viewType := u.GetRequestValue(r, "t", "1")
	tName := "noteview" + viewType + ".html"

	var aNote *models.Note
	noteID, _ := strconv.ParseInt(u.GetRequestValue(r, "id", "0"), 10, 64)
	if noteID == 0 {
		noteTitle := u.GetRequestValue(r, "title", "")
		aNote = models.GetNote(noteTitle)
	} else {
		aNote = models.GetNoteByID(noteID)
	}
	if aNote == nil {
		return
	}
	if aNote.Permission < 5 {
		isAuth := models.GetSessionVal(r, "authenticated", nil)
		if isAuth == nil || !isAuth.(bool) {
			mlog.Error(fmt.Errorf("no session"))
			from_uri := r.URL.RequestURI()
			from_uri = strings.TrimPrefix(from_uri, "/")
			http.Redirect(w, r, "/login?from_uri="+from_uri, http.StatusTemporaryRedirect)
			return
		}
		u := GetCurrentUser(&w, r)
		if !models.CheckPerm(aNote.Object, u.ID, "r") {
			mlog.Error(fmt.Errorf("no permission"))
			fmt.Fprintf(w, "Permission denied")
			return
		}
	}

	data := map[string]interface{}{
		"title":     "Webnote - " + aNote.Title,
		"page":      "noteview",
		"msg":       "",
		"note":      aNote,
		"revisions": models.GetNoteRevisions(aNote.ID),
	}
	if len(aNote.Attachments) > 0 {
		data["attachments"] = aNote.Attachments
	}
	isAjax := u.GetRequestValue(r, "is_ajax", "0")
	if isAjax == "0" {
		CommonRenderTemplate(tName, &w, r, &data)
	} else {
		fmt.Fprint(w, u.JsonDump(aNote.Sanitize(), "  "))
	}
}

func DoDeleteNote(w http.ResponseWriter, r *http.Request) {
	noteIDStr := u.GetRequestValue(r, "id", "0")
	msg := "OK"
	if noteIDStr != "0" {
		user := GetCurrentUser(&w, r)
		noteID, _ := strconv.ParseInt(noteIDStr, 10, 64)
		aNote := models.GetNoteByID(noteID)
		if models.CheckPerm(aNote.Object, user.ID, "d") {
			aNote.Delete()
		} else {
			msg = "Permission denied."
		}
	} else {
		msg = "Invalid noteID"
	}
	isAjax := u.GetRequestValue(r, "is_ajax", "0")
	page := u.GetRequestValue(r, "page", "0")
	keyword := u.GetRequestValue(r, "keyword", "")
	var redirectURL string
	switch page {
	case "frontpage":
		redirectURL = "/"
	default:
		redirectURL = "/search?keyword=" + keyword
	}
	if isAjax == "1" {
		fmt.Fprint(w, msg)
	} else {
		if msg != "OK" {
			fmt.Fprint(w, msg)
		} else {
			http.Redirect(w, r, fmt.Sprint(redirectURL), http.StatusFound)
		}
	}
}

func ClearSession(w *http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "auth-session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(*w, cookie)
}

func DoLogout(w http.ResponseWriter, r *http.Request) {
	ClearSession(&w)
	http.Redirect(w, r, "/", http.StatusFound)
}

var CSRF_TOKEN string

//HandleRequests -

func DoViewRevNote(w http.ResponseWriter, r *http.Request) {
	viewType := u.GetRequestValue(r, "t", "1")
	tName := "noteview" + viewType + ".html"

	noteID, _ := strconv.ParseInt(u.GetRequestValue(r, "id", "0"), 10, 64)
	aNote := models.GetNoteRevisionByID(noteID)
	user := GetCurrentUser(&w, r)
	if !models.CheckPerm(aNote.Object, user.ID, "r") {
		fmt.Fprintf(w, "ERROR Permission denied")
		return
	}
	CommonRenderTemplate(tName, &w, r, &map[string]interface{}{
		"title": "Webnote - " + aNote.Title,
		"page":  "noteview",
		"msg":   "",
		"note":  aNote,
	})
}

func DoViewDiffNote(w http.ResponseWriter, r *http.Request) {
	noteID, _ := strconv.ParseInt(u.GetRequestValue(r, "id", "0"), 10, 64)
	revNoteID, _ := strconv.ParseInt(u.GetRequestValue(r, "rev_id", "0"), 10, 64)
	n := models.GetNoteByID(noteID)
	user := GetCurrentUser(&w, r)
	if !models.CheckPerm(n.Object, user.ID, "r") {
		fmt.Fprintf(w, "ERROR Permission denied")
		return
	}
	n1 := models.GetNoteRevisionByID(revNoteID)
	nd := n1.Diff(n)
	t, _ := template.New("diff").Parse(`<html>
	<body>{{ .htmlText }}</body>
	</html>
	`)
	if e := t.ExecuteTemplate(w, "diff", map[string]interface{}{
		"htmlText": template.HTML(nd.String()),
	}); e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
}

func DoUpload(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		CommonRenderTemplate("upload.html", &w, r, &map[string]interface{}{
			"title": "Webnote - Upload attachment",
			"page":  "upload",
			"msg":   "",
		})
	case "POST":
		user := GetCurrentUser(&w, r)
		if user == nil {
			http.Error(w, "ERROR", http.StatusForbidden)
			return
		}
		var aList []*models.Attachment

		if err := r.ParseMultipartForm(models.MaxUploadSizeInMemory); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for count := 0; count < models.Settings.UPLOAD_ITEM_COUNT; count++ {
			cStr := strconv.Itoa(count)
			file, handler, err := r.FormFile("myFile" + cStr)
			if u.CheckErrNonFatal(err, "Error Retrieving the File") != nil {
				continue
			}
			defer file.Close()
			fmt.Printf("Uploaded File: %+v\n", handler.Filename)
			fmt.Printf("File Size: %+v\n", handler.Size)
			fmt.Printf("MIME Header: %s\n", handler.Header["Content-Type"])
			aName := u.GetRequestValue(r, "a"+cStr, "")
			aDesc := u.GetRequestValue(r, "desc"+cStr, "")
			if aName == "" {
				aName = handler.Filename
			}
			uploadPath := u.GetRequestValue(r, "upload_path"+cStr)
			attachedFileDir := u.Ternary(uploadPath == "", models.UpLoadPath, filepath.Join(models.UpLoadPath, uploadPath)).(string)
			a := models.Attachment{
				Name:         aName,
				Description:  aDesc,
				AttachedFile: filepath.Join(attachedFileDir, handler.Filename),
				FileSize:     handler.Size,
			}
			os.MkdirAll(attachedFileDir, os.ModePerm)

			f, err := os.OpenFile(a.AttachedFile, os.O_WRONLY|os.O_CREATE, 0666)
			mlog.FatalIfError(err)

			copyFile := func(f *os.File, file io.Reader) {
				bWritten, e := io.Copy(f, file)
				if e != nil {
					mlog.Error(fmt.Errorf("copied %d and got error %s", bWritten, e.Error()))
					return
				}
				f.Close()
			}
			copyFile(f, file)

			a.AuthorID = user.ID
			a.GroupID = user.Groups[0].ID
			mimetype := fmt.Sprintf("%s", handler.Header["Content-Type"])
			mimetype = strings.ReplaceAll(mimetype, "[", "")
			mimetype = strings.ReplaceAll(mimetype, "]", "")
			a.Mimetype = mimetype
			_p, _ := strconv.Atoi(u.GetRequestValue(r, "permission", "1"))
			a.Permission = int8(_p)
			a.Save()
			aList = append(aList, &a)
		}
		//Cleanup temp files
		r.MultipartForm.RemoveAll()

		tmplStr := `<html><body>OK Attachment(s) created
		{{ $settings := .settings }}
		{{ range $idx, $a := .attachments }}
			<ul>
				<li>
					Name: {{ $a.Name }} - Path: {{ $a.AttachedFile }}<br>
					<a href="{{ $settings.BASE_URL }}/streamfile?id={{ $a.ID }}&action=stream">view</a>
					-
					<a href="{{ $settings.BASE_URL }}/streamfile?id={{ $a.ID }}&action=download">download</a>
					-
					<a href="{{ $settings.BASE_URL }}/edit_attachment?id={{ $a.ID }}">edit</a>
				</li>
			</ul>
		{{ end }}
		<hr>
		<ul>
			<li><a href="/">Home</a></li>
			<li><a href="/upload">More uploads</a></li>
			<li><a href="/list_attachment">List files</a></li>
		</ul>
		</body></html>`
		t := template.Must(template.New("a").Parse(tmplStr))
		t.Execute(w, map[string]interface{}{
			"settings":    models.Settings,
			"attachments": aList,
		})
		return
	}
}

func GetCurrentNote(w http.ResponseWriter, r *http.Request) *models.Note {
	noteIDStr := u.GetRequestValue(r, "note_id", "")
	if noteIDStr != "" {
		user := GetCurrentUser(&w, r)
		noteID, _ := strconv.ParseInt(noteIDStr, 10, 64)
		n := models.GetNoteByID(noteID)
		if models.CheckPerm(n.Object, user.ID, "r") {
			return n
		}
	}
	return nil
}

func DoListAttachment(w http.ResponseWriter, r *http.Request) {
	kw := u.GetRequestValue(r, "keyword", "")
	aNote := GetCurrentNote(w, r)
	user := GetCurrentUser(&w, r)
	aList := models.SearchAttachment(kw, user)
	data := map[string]interface{}{
		"title":       "Webnote - List attachements",
		"page":        "list_attachement",
		"msg":         "",
		"attachments": aList,
	}
	if aNote != nil {
		data["note"] = aNote
	}
	CommonRenderTemplate("list_attachment.html", &w, r, &data)
}

func DoDeleteAttachment(w http.ResponseWriter, r *http.Request) {
	aIDStr := u.GetRequestValue(r, "id", "")
	if aIDStr == "" {
		return
	}
	aID, _ := strconv.ParseInt(aIDStr, 10, 64)
	user := GetCurrentUser(&w, r)
	a := models.GetAttachementByID(aID)
	if e := a.DeleteAttachment(user); e != nil {
		msg := fmt.Sprintf("ERROR Can not delete attachment - %v", e)
		http.Error(w, msg, http.StatusOK)
		return
	}
	is_ajax := u.GetRequestValue(r, "is_ajax", "0")
	if is_ajax == "1" {
		fmt.Fprintf(w, "Deleted attachement ID %d", aID)
	}
}

func DoStreamfile(w http.ResponseWriter, r *http.Request) {
	aIDStr := u.GetRequestValue(r, "id", "")
	if aIDStr == "" {
		http.Error(w, "No aID provided", http.StatusBadRequest)
		return
	}
	aID, _ := strconv.ParseInt(aIDStr, 10, 64)
	a := models.GetAttachementByID(aID)
	user := GetCurrentUser(&w, r)

	if a.Permission < 5 {
		isAuth := models.GetSessionVal(r, "authenticated", nil)
		if isAuth == nil || !isAuth.(bool) {
			mlog.Error(fmt.Errorf("no session"))
			from_uri := r.URL.RequestURI()
			from_uri = strings.TrimPrefix(from_uri, "/")
			http.Redirect(w, r, "/login?from_uri="+from_uri, http.StatusTemporaryRedirect)
			return
		}
		if !models.CheckPerm(a.Object, user.ID, "r") {
			mlog.Error(fmt.Errorf("no permission user %s - attachment name: %s", user, a.Name))
			fmt.Fprintf(w, "Permission denied")
			return
		}
	}

	file, err := os.Open(a.AttachedFile)
	if err != nil {
		http.Error(w, "ERROR", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	// stream straight to client(browser)
	w.Header().Set("Content-type", a.Mimetype)
	action := u.GetRequestValue(r, "action", "")

	if action == "download" {
		w.Header().Set("Content-Disposition", "attachment; filename="+path.Base(a.AttachedFile))
		http.ServeFile(w, r, a.AttachedFile)

	} else {
		http.ServeContent(w, r, "playing", time.Time{}, file)
	}
}

func DoAttachmentToNote(w http.ResponseWriter, r *http.Request) {
	action := u.GetRequestValue(r, "action", "")
	user := GetCurrentUser(&w, r)
	n := GetCurrentNote(w, r)
	attachmentIDStr := u.GetRequestValue(r, "attachment_id", "")
	if attachmentIDStr == "" {
		return
	}
	attachmentID, _ := strconv.ParseInt(attachmentIDStr, 10, 64)
	if action == "unlink" {
		if e := n.UnlinkAttachment(attachmentID, user); e != nil {
			msg := fmt.Sprintf("ERROR unlink attachment to note - %v\n", e)
			fmt.Fprint(w, msg)
			return
		}
	} else {
		if e := n.LinkAttachment(attachmentID, user); e != nil {
			mlog.Error(fmt.Errorf("link attachment to note - %s", e.Error()))
			fmt.Fprintf(w, "ERROR")
			return
		}
	}
	fmt.Fprintf(w, "OK")
}

func DoEditUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		allGroups := []string{}
		for _, gn := range models.GetAllGroups() {
			allGroups = append(allGroups, gn.Name)
		}
		CommonRenderTemplate("edituser.html", &w, r, &map[string]interface{}{
			"title":     "Webnote - Edit User",
			"page":      "edituser",
			"allgroups": strings.Join(allGroups, ","),
		})
	case "POST":
		action := u.GetRequestValue(r, "submit", "")
		cUser := GetCurrentUser(&w, r)
		userEmail := u.GetRequestValue(r, "email", "")
		password := u.GetRequestValue(r, "cur_password")
		newPassword := u.GetRequestValue(r, "password")
		userData := map[string]interface{}{
			"FirstName":   u.GetRequestValue(r, "f_name", ""),
			"LastName":    u.GetRequestValue(r, "l_name", ""),
			"Email":       u.GetRequestValue(r, "email"),
			"HomePhone":   u.GetRequestValue(r, "h_phone"),
			"WorkPhone":   u.GetRequestValue(r, "w_phone"),
			"MobilePhone": u.GetRequestValue(r, "m_phone"),
			"ExtraInfo":   u.GetRequestValue(r, "extra_info"),
			"Address":     u.GetRequestValue(r, "address"),
			"Password":    newPassword,
			"GroupNames":  u.GetRequestValue(r, "group_names", "default"),
		}
		if cUser.Email != userEmail {
			//We are updating other user, not current user. We need to be admin to do so
			if cUser.Email != models.Settings.ADMIN_EMAIL {
				mlog.Error(fmt.Errorf("permission denied. Only admin has right to update other user"))
				fmt.Fprintf(w, "Permission denied")
				return
			}
		}
		//Checking admin password to confirm
		if !u.VerifyHash(password, cUser.PasswordHash, int(cUser.SaltLength)) {
			mlog.Error(fmt.Errorf("permission denied. Old/Admin password provided incorrect"))
			fmt.Fprintf(w, "Permission denied")
			return
		}
		switch action {
		case "Add/Edit User":
			user := models.UserNew(userData)
			fmt.Fprintf(w, "OK user detail updated for %s. You need to generate TOP QR code using previous form", user.Email)
		case "Generate new OTP QR image":
			user := models.GetUser(userEmail)
			pngImageBuff := models.SetUserOTP(user.Email)
			w.Write(pngImageBuff.Bytes())
			return
		case "Add Groups":
			if cUser.Email != models.Settings.ADMIN_EMAIL {
				mlog.Error(fmt.Errorf("permission denied. Only admin has right to add more groups"))
				fmt.Fprintf(w, "Permission denied")
				return
			} else {
				groups := strings.Split(u.GetRequestValue(r, "new_group_names"), `,`)
				for _, gn := range groups {
					gn = strings.TrimSpace(gn)
					newGroup := models.Group{
						Name: gn,
					}
					newGroup.Save()
				}
				fmt.Fprintf(w, "Groups added %s", groups)
			}
			return
		case "Unlock Account":
			user := models.UserNew(userData)
			user.AttemptCount = 1
			user.Save()
			fmt.Fprintf(w, "Account %s is unlocked", user.Email)
			return
		}
	}
}

func DoSearchUser(w http.ResponseWriter, r *http.Request) {
	user := GetCurrentUser(&w, r)
	if user.Email != models.Settings.ADMIN_EMAIL {
		fmt.Fprint(w, "Permission denied")
		return
	}
	kw := u.GetRequestValue(r, "kw", "")
	foundUsers := models.SearchUser(kw)
	if len(foundUsers) == 0 {
		fmt.Fprint(w, "No user found")
		return
	}
	foundUser := foundUsers[0]
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	o, e := json.MarshalToString(foundUser)
	if e != nil {
		mlog.Error(fmt.Errorf("can not create json %s", e.Error()))
		return
	}
	fmt.Fprint(w, o)
}

func DoEditAttachment(w http.ResponseWriter, r *http.Request) {
	_aID := u.GetRequestValue(r, "id")
	aID, _ := strconv.ParseInt(_aID, 10, 64)
	a := models.GetAttachementByID(aID)
	switch r.Method {
	case "GET":
		CommonRenderTemplate("editattachment.html", &w, r, &map[string]interface{}{
			"title":      "Webnote - Edit Attachment",
			"page":       "editattachment",
			"attachment": a,
		})
	case "POST":
		user := GetCurrentUser(&w, r)
		if models.CheckPerm(a.Object, user.ID, "w") {
			action := u.GetRequestValue(r, "submit")
			a.Name = u.GetRequestValue(r, "a_name")
			a.Description = u.GetRequestValue(r, "a_desc")
			_perm, _ := strconv.Atoi(u.GetRequestValue(r, "permission"))
			a.Permission = int8(_perm)
			g := models.GetGroup(u.GetRequestValue(r, "ngroup"))
			a.GroupID = g.ID

			switch action {
			case "Edit Attachment":
				a.Save()
				fmt.Fprint(w, "OK Attachment updated")
				return
			case "Encrypt with zip":
				key := u.ZipEncript(a.AttachedFile, a.AttachedFile+".zip")
				os.Remove(a.AttachedFile)
				a.AttachedFile = a.AttachedFile + ".zip"
				a.Save()
				fmt.Fprintf(w, "OK Attachment encrypted with key: '%s'", key)
				return
			case "Decrypt with zip":
				key := u.GetRequestValue(r, "zipkey", "")
				if e := u.ZipDecrypt(a.AttachedFile, key); e == nil {
					os.Remove(a.AttachedFile)
					a.AttachedFile = strings.TrimSuffix(a.AttachedFile, ".zip")
					a.Save()
					fmt.Fprintf(w, "OK Attachment decrypted. File '%s'", a.AttachedFile)
				} else {
					fmt.Fprintf(w, "%v", e)
				}
				return
			}
		} else {
			fmt.Fprint(w, "Permission denied")
			return
		}
	}
}

func DoAutoScanAttachment(w http.ResponseWriter, r *http.Request) {
	u := GetCurrentUser(&w, r)
	if u.Email != models.Settings.ADMIN_EMAIL {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}
	o := models.ScanAttachment("uploads/", u)
	fmt.Fprintf(w, "Add / Update list: %v", o)
}

func OnSelectedRunSql(w http.ResponseWriter, r *http.Request) {
	sql := u.GetRequestValue(r, "sql")
	ptn := regexp.MustCompile(`^[\s]*(UPDATE|update|DELETE|delete) `)
	if !ptn.MatchString(sql) {
		fmt.Fprintf(w, "Invalid sql provided")
		mlog.Warning("Invalid sql provided - %s", sql)
		return
	}
	if !strings.Contains(sql, " WHERE id in ") {
		fmt.Fprintf(w, "Invalid sql provided. No WHERE clause")
		mlog.Warning("Invalid sql provided No WHERE clause - %s", sql)
		return
	}
	user := GetCurrentUser(&w, r)
	sql = fmt.Sprintf("%s AND author_id = %d", sql, user.ID)
	mlog.Info("going to run sql: %s\n", sql)
	DB := models.GetDB("")
	defer DB.Close()
	tx, err := DB.Begin()
	if err != nil {
		mlog.IfError(err)
		fmt.Fprintf(w, "Can not start TX. See app log for details")
		return
	}
	_, err = tx.Exec(sql)
	if err != nil {
		mlog.IfError(err)
		fmt.Fprintf(w, "Can not exec sql. See app log for details")
		return
	}
	if err := tx.Commit(); err != nil {
		mlog.IfError(err)
		fmt.Fprintf(w, "Can not commit sql. See app log for details")
		return
	}
	fmt.Fprintf(w, "Success!")
}

// Sync with gnote. gnote will send a request to get notes with titles and ids - and check itself to see if it has these titles
// and then make a decision to pull or not. Then it will send to get full notes by ids
// Search notes but only return a list of note titles and ids. Search based on time range, and/or some other condition
func DoGetNoteTitles(w http.ResponseWriter, r *http.Request) {
	duration := u.GetRequestValue(r, "duration", "")
	tzString := u.GetRequestValue(r, "tz", "Australia/Brisbane")
	starttime, endtime := u.ParseTimeRange(duration, tzString)
	sqlwhere := fmt.Sprintf("datelog >= %d AND datelog <= %d", starttime.UnixNano(), endtime.UnixNano())
	user := GetCurrentUser(&w, r)
	notes := models.Query(sqlwhere, user, true)
	fmt.Fprint(w, u.JsonDump(notes, ""))
}

func DoGetNotesByIds(w http.ResponseWriter, r *http.Request) {
	listIdStr := u.GetRequestValue(r, "ids", "") // should get example (1,4,5,6)
	sqlwhere := "id in " + listIdStr
	user := GetCurrentUser(&w, r)
	notes := models.Query(sqlwhere, user, false)
	fmt.Fprint(w, u.JsonDump(notes, ""))
}

func HandleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	base_url := models.GetConfig("base_url", "")
	_u, _e := url.Parse(base_url)
	if _e != nil {
		panic(_e)
	}
	app_domain, _, _ := net.SplitHostPort(_u.Host)
	// router := StaticRouter()
	CSRF_TOKEN := u.MakePassword(32)
	csrf.MaxAge(4 * 3600)

	// CSRF := csrf.Protect(
	// 	[]byte(CSRF_TOKEN),
	// 	// instruct the browser to never send cookies during cross site requests
	// 	csrf.SameSite(csrf.SameSiteStrictMode),
	// 	csrf.TrustedOrigins([]string{"note.inxuanthuy.com", "note.xvt.technology"}),
	// 	// csrf.RequestHeader("X-CSRF-Token"),
	// 	// csrf.FieldName("authenticity_token"),
	// 	// csrf.ErrorHandler(http.HandlerFunc(serverError(403))),
	// )
	// csrf.Secure(true)
	// router.Use(CSRF)
	//By pass csrf for /view See https://stackoverflow.com/questions/53271241/disable-csrf-on-json-api-calls
	protectionMiddleware := func(handler http.Handler) http.Handler {
		protectionFn := csrf.Protect(
			[]byte(CSRF_TOKEN),
			// instruct the browser to never send cookies during cross site requests
			csrf.SameSite(csrf.SameSiteStrictMode),
			csrf.TrustedOrigins([]string{app_domain}),
			// csrf.RequestHeader("X-CSRF-Token"),
			// csrf.FieldName("authenticity_token"),
			// csrf.ErrorHandler(http.HandlerFunc(serverError(403))),
			csrf.Secure(false),
		)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use some kind of condition here to see if the router should use
			// the CSRF protection. we'll check the path prefix.
			if !strings.HasPrefix(r.URL.Path, "/nocsrf") {
				protectionFn(handler).ServeHTTP(w, r)
				return
			}
			handler.ServeHTTP(w, r)
		})
	}

	staticFS := http.FileServer(http.Dir("./assets"))
	//Not sure why this line wont work but the one after that works for serving static
	// router.Handle("/assets/", http.StripPrefix("/assets/", staticFS)).Methods("GET")
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", staticFS))

	router.HandleFunc("/login", DoLogin).Methods("GET", "POST")
	router.Handle("/", isAuthorized(HomePage)).Methods("GET")

	//All routes handlers
	router.Handle("/savenote", isAuthorized(DoSaveNote)).Methods("POST")
	router.Handle("/search", isAuthorized(DoSearchNote)).Methods("POST", "GET")
	//some note is universal viewable thus we dont put isAuthorized here but check perms at the handler func
	router.HandleFunc("/view", DoViewNote).Methods("GET", "POST")
	router.Handle("/view_rev", isAuthorized(DoViewRevNote)).Methods("GET")
	router.Handle("/view_diff", isAuthorized(DoViewDiffNote)).Methods("GET")
	router.Handle("/delete", isAuthorized(DoDeleteNote)).Methods("POST", "GET")
	router.Handle("/logout", isAuthorized(DoLogout)).Methods("POST", "GET")
	router.Handle("/upload", isAuthorized(DoUpload)).Methods("POST", "GET")
	router.Handle("/list_attachment", isAuthorized(DoListAttachment)).Methods("GET")
	router.Handle("/edit_attachment", isAuthorized(DoEditAttachment)).Methods("GET", "POST")
	router.Handle("/auto_scan_attachment", isAuthorized(DoAutoScanAttachment)).Methods("GET", "POST")
	router.Handle("/delete_attachment", isAuthorized(DoDeleteAttachment)).Methods("GET")
	router.HandleFunc("/streamfile", DoStreamfile).Methods("GET")
	router.Handle("/add_attachment_to_note", isAuthorized(DoAttachmentToNote)).Methods("GET")
	router.Handle("/delete_note_attachment", isAuthorized(DoAttachmentToNote)).Methods("GET")
	router.Handle("/get_notes_titles", isAuthorized(DoGetNoteTitles)).Methods("GET")
	router.Handle("/get_notes_by_id", isAuthorized(DoGetNotesByIds)).Methods("GET")

	//User management
	router.Handle("/edituser", isAuthorized(DoEditUser)).Methods("GET", "POST")
	router.Handle("/searchuser", isAuthorized(DoSearchUser)).Methods("GET")
	// With selected - handle bulk ops from the search result
	router.Handle("/on_selected_run_sql", isAuthorized(OnSelectedRunSql)).Methods("POST")

	//SinglePage (as note content) handler. Per app the controller file is in app-controllers folder. The javascript app needs to get the token and send it with its post request. Eg. var csrfToken = document.getElementsByName("gorilla.csrf.Token")[0].value
	router.Handle("/cred", isAuthorized(app.DoCredApp)).Methods("POST", "GET")

	//kodi send. Dont need to authenticate via note app but its own (via IP)
	router.Handle("/kodi/add", app.KodiIsAuthorized(app.HandleAddToPlayList)).Methods("POST")
	router.Handle("/kodi/play", app.KodiIsAuthorized(app.HandlePlay)).Methods("POST")
	router.Handle("/kodi/loadlist", app.KodiIsAuthorized(app.HandleLoadList)).Methods("POST")
	router.Handle("/kodi/savelist", app.KodiIsAuthorized(app.HandleSaveList)).Methods("POST")
	router.Handle("/kodi/playlist", app.KodiIsAuthorized(app.HandlePlayList)).Methods("POST")
	//Save bookmark
	router.Handle("/savebookmark", isAuthorized(app.SaveBookMark)).Methods("GET")
	router.Handle("/delbookmark", isAuthorized(app.DeleteBookMark)).Methods("GET")
	//A random generator
	router.HandleFunc("/rand", app.GenRandNumber).Methods("GET")
	// Onetime secret share
	router.HandleFunc("/nocsrf/onetimesec/generate", app.GenerateOnetimeSecURL).Methods("POST")
	router.HandleFunc("/nocsrf/onetimesec/{secret_id}", app.GetOnetimeSecret).Methods("GET")

	srv := &http.Server{
		Addr: ":" + ServerPort,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15000,
		ReadTimeout:  time.Second * 15000,
		IdleTimeout:  time.Second * 60,
		//Handler: handlers.CompressHandler(router), // Pass our instance of gorilla/mux in.
	}
	if (*EnableCompression == "") && (os.Getenv("HTTP_ENABLE_COMPRESSION") == "true") {
		*EnableCompression = "yes"
	}
	if *EnableCompression == "yes" {
		srv.Handler = handlers.CompressHandler(protectionMiddleware(router))
	} else {
		srv.Handler = protectionMiddleware(router)
	}

	if SSLKey != "" {
		mlog.Info("Start SSL/TLS server on port %s\n", ServerPort)
		mlog.FatalIfError(srv.ListenAndServeTLS(SSLCert, SSLKey))
	} else {
		mlog.Info("Start server on port %s\n", ServerPort)
		mlog.FatalIfError(srv.ListenAndServe())
	}
}

func main() {
	tz := os.Getenv("TZ")
	if tz != "" {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			fmt.Printf("[WARN] can not load timezone %s\n", tz)
		} else {
			time.Local = loc // -> this is setting the global timezone
		}
	}
	getVersion := flag.Bool("v", false, "Get build version")
	dbPath := flag.String("db", "", "Application DB path")
	sessionKey := flag.String("sessionkey", "", "Session Key")
	setup := flag.Bool("setup", false, "Run initial setup DB and config")
	sslKey := flag.String("key", "", "SSL Key path")
	sslCert := flag.String("cert", "", "SSL Cert path")
	port := flag.String("p", "", "Port")
	base_url := flag.String("baseurl", "", "baseurl")
	cmd := flag.String("cmd", "", "Command utils to manage config")
	//Adduser data command
	useremail := flag.String("email", "", "User email")
	userpassword := flag.String("password", "", "User password")
	usergroup := flag.String("group", "", "User Group. Any of default|family|friend or coma separated ")
	EnableCompression = flag.String("comp", "", "Enable server compression. Dont use it for https")
	AttachmentDir := flag.String("attachmentdir", "", "Directory path to scan attachments for auto add attachment command '-cmd scan_attachment'")
	rootUploadDir := flag.String("upload-dir", "uploads/", "Default Root dir for uploads files")

	flag.Usage = func() {
		flag.PrintDefaults()
		msg := `
		Quick start
		To setup initial database use option '-db path-to-your-new-db.db -key mykey.key -cert mykey.crt -p port -setup'

		The ssl cert and key if it does not exist then will be created automatically.

		The default admin email to login is admin@admin.comodels. To change this use option '-cmd set_admin_email'
		The default password for admin@admin.com is 1qa2ws. When logged in follow the instructions on screen to change password and generate QR OTP image for MFA.

		Next run remove the option -setup to start the app

		The command after -cmd can be: set_admin_password, set_admin_otp, add_user

		You have to run set_admin_otp to get the OTP QR image produced in the current directory and use it for admin login.

		If a user is blacklisted from an IP (account lock) run sqlite3 path-to-your-db
		Then run this sql 'DELETE FROM appconfig WHERE key="blacklist_ips";'

		Try to select * from appconfig to see what key you can do. The whitelist_ip is a list of IP that user does not need to use OTP to login.
		`
		fmt.Fprint(os.Stderr, msg)
	}

	flag.Parse()
	models.UpLoadPath = *rootUploadDir

	if *getVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	ServerPort = *port

	os.Setenv("DBPATH", *dbPath)
	if _, err := os.Stat(*dbPath); errors.Is(err, os.ErrNotExist) {
		*setup = true
	}
	switch os.Getenv("RESET_STORAGE") {
	case "yes", "true":
		journalPath := fmt.Sprintf("%s-journal", *dbPath)
		mlog.Info("RESET_STORAGE is set going to remove %s and %s\n", *dbPath, journalPath)
		if err := os.Remove(*dbPath); err != nil {
			mlog.Info("remove dbPath error: %v\n", err)
		}
		if err := os.Remove(journalPath); err != nil {
			mlog.Info("remove journalPath error: %v\n", err)
		}
		*setup = true
	}
	if *setup {
		mlog.Info("SETUP is called and started\n")
		models.SetupAppDatabase()
		models.SetupDefaultConfig()
		models.CreateAdminUser()
		models.CreatePublicReadUser()
		if _, err := os.Stat(*sslKey); os.IsNotExist(err) {
			keyFileName := u.FileNameWithoutExtension(*sslKey)
			u.GenSelfSignedKey(keyFileName)
			*sslCert = fmt.Sprintf("%s.crt", keyFileName)
		}
	}

	SSLKey = models.GetConfigSave("ssl_key", *sslKey)
	SSLCert = models.GetConfigSave("ssl_cert", *sslCert)
	models.GetConfigSave("base_url", *base_url)

	models.InitConfig()

	if *cmd != "" {
		//Run command utils
		switch *cmd {
		case "list":
			fmt.Printf(`List of commands:
			set_admin_password
			set_admin_otp
			set_admin_email
			add_user - take more opt username, password, group
			scan_attachment - take option attachmentdir or leave it empty to use the default 'uploads' folder`)
		case "set_admin_password":
			models.SetAdminPassword()
		case "set_admin_otp":
			models.SetAdminOTP()
		case "set_admin_email":
			models.SetAdminEmail("")
		case "add_user":
			models.AddUser(map[string]interface{}{
				"username": *useremail,
				"password": *userpassword,
				"group":    *usergroup,
			})
		case "scan_attachment":
			aDir := u.Ternary(AttachmentDir == nil, "uploads", *AttachmentDir).(string)
			u := models.GetUser(models.Settings.ADMIN_EMAIL)
			models.ScanAttachment(aDir, u)
		}
	} else { //Server mode
		if *sessionKey == "" {
			*sessionKey = models.GetConfig("session-key", "")
			if *sessionKey == "" {
				*sessionKey = u.MakePassword(64)
				models.SetConfig("session-key", *sessionKey)
			}
		}
		models.SessionStore = sessions.NewCookieStore([]byte(*sessionKey))
		models.SessionStore.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   3600 * 4,
			HttpOnly: true,
		}
		// log.Println(*sessionKey)
		models.LoadAllTemplates()
		HandleRequests()
	}
}

func DoLogin(w http.ResponseWriter, r *http.Request) {
	trycount := models.GetSessionVal(r, "trycount", 0).(int)
	_userIP := models.ReadUserIP(r)
	portPtn := regexp.MustCompile(`\:[\d]+$`)
	userIP := portPtn.ReplaceAllString(_userIP, "")
	ses, _ := models.SessionStore.Get(r, "auth-session")

	currentBlackList := models.GetConfig("blacklist_ips", "")

	if trycount >= 30 {
		currentBlackList = currentBlackList + "," + userIP
		models.SetConfig("blacklist_ips", currentBlackList)
	}
	if strings.Contains(currentBlackList, userIP) {
		ses.Values = make(map[interface{}]interface{}, 1) //Empty session values
		ses.Save(r, w)
		http.Error(w, "Account locked", http.StatusForbidden)
		//To remove the lock use command below on the terminal - adjust the val properly. And clear the browser cookie as well
		//ql -db testwebnote.db  'update appconfig set val = "" where key = "blacklist_ips"'
		return
	}

	trycount = trycount + 1
	models.SaveSessionVal(r, &w, "trycount", trycount)

	isAuthenticated := models.GetSessionVal(r, "authenticated", nil)

	switch r.Method {
	case "GET":
		if isAuthenticated == nil || !isAuthenticated.(bool) {
			from_uri := r.URL.RequestURI()
			from_uri = strings.TrimPrefix(from_uri, "/login?from_uri=")
			data := map[string]interface{}{
				csrf.TemplateTag: csrf.TemplateField(r),
				"title":          "Webnote",
				"page":           "login",
				"msg":            "",
				"client_ip":      userIP,
				"from_uri":       from_uri,
			}
			if err := models.AllTemplates.ExecuteTemplate(w, "login.html", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			from_uri := u.GetRequestValue(r, "from_uri", "")
			from_uri = strings.TrimPrefix(from_uri, "/")
			http.Redirect(w, r, "/"+from_uri, http.StatusFound)
		}
	case "POST":
		r.ParseForm()
		useremail := r.FormValue("username")
		password := r.FormValue("password")

		var user *models.User

		totop := r.FormValue("totp_number")
		mlog.Info("user input totp %s\n", totop)
		user, err := models.VerifyLogin(useremail, password, totop, userIP)
		ses, _ := models.SessionStore.Get(r, "auth-session")
		if user != nil {
			if strings.Contains(user.ExtraInfo, " First Time Login ") {
				ses.Values["first_time_login"] = "yes"
				user.ExtraInfo = strings.ReplaceAll(user.ExtraInfo, " First Time Login ", "")
				user.Save()
			} else {
				ses.Values["first_time_login"] = "no"
			}
			mlog.Info("Verified user %v\n", user)
			ses.Values["authenticated"] = true
			ses.Values["trycount"] = 0
			ses.Values["useremail"] = useremail
			ses.Save(r, w)
			from_uri := u.GetRequestValue(r, "from_uri", "")
			from_uri = strings.TrimPrefix(from_uri, "/")
			http.Redirect(w, r, "/"+from_uri, http.StatusFound)
			return
		} else {
			mlog.Info("Failed To Verify user %s - %s\n", useremail, err.Error())
			ses.Values["authenticated"] = false
			ses.Values["useremail"] = ""
			ses.Save(r, w)
			http.Error(w, "Failed login", http.StatusForbidden)
			return
		}
	}
}

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuth := models.GetSessionVal(r, "authenticated", nil)
		if isAuth == nil || !isAuth.(bool) {
			mlog.Error(fmt.Errorf("no session"))
			uri := r.RequestURI
			// uri = url.PathEscape(uri)
			uri = strings.TrimPrefix(uri, "/")
			http.Redirect(w, r, "/login?from_uri="+uri, http.StatusTemporaryRedirect)
			return
		}
		// w.Header().Set("X-CSRF-Token", csrf.Token(r))
		endpoint(w, r)
	})
}

func CommonRenderTemplate(tmplName string, w *http.ResponseWriter, r *http.Request, mapData *map[string]interface{}) {
	user := GetCurrentUser(w, r)
	keyword := u.GetRequestValue(r, "keyword", "")

	var uGroups []*models.Group
	var commonMapData map[string]interface{}

	if user != nil {
		uGroups = user.Groups
		commonMapData = map[string]interface{}{
			csrf.TemplateTag:  csrf.TemplateField(r),
			"keyword":         keyword,
			"settings":        models.Settings,
			"user":            user,
			"groups":          uGroups,
			"permission_list": models.PermissionList,
			"date_layout":     models.DateLayout,
		}
	} else {
		commonMapData = map[string]interface{}{
			csrf.TemplateTag:  csrf.TemplateField(r),
			"keyword":         keyword,
			"settings":        models.Settings,
			"permission_list": models.PermissionList,
			"date_layout":     models.DateLayout,
		}
	}
	for _k, _v := range *mapData {
		commonMapData[_k] = _v
	}
	if err := models.AllTemplates.ExecuteTemplate(*w, tmplName, commonMapData); err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
	}
}
