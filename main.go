package main

import (
	"path"
	"io"
	"html/template"
	"strconv"
	"regexp"
	"strings"
	"github.com/gorilla/sessions"
	"fmt"
	"os"
	"flag"
	"time"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"github.com/gorilla/csrf"
	m "github.com/sunshine69/webnote-go/models"
	"github.com/sunshine69/webnote-go/app"
)

var ServerPort, SSLKey, SSLCert string
var EnableCompression *string

func init() {
	SSLKey = m.GetConfig("ssl_key", "")
	SSLCert = m.GetConfig("ssl_cert", "")
}

func GetCurrentUser(w *http.ResponseWriter, r *http.Request) *m.User {
	isAuth := m.GetSessionVal(r, "authenticated", nil)
	if isAuth == nil || ! isAuth.(bool) {
		return nil
	}
	useremail := m.GetSessionVal(r, "useremail", "").(string)
	if useremail == "" {
		return nil
	}
	return m.GetUser(useremail)
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	noteID, _ := strconv.ParseInt(m.GetRequestValue(r, "id", "0"), 10, 64)
	raw_editor, _ := strconv.Atoi(m.GetRequestValue(r, "raw_editor", "1"))

	var aNote *m.Note
	u := GetCurrentUser(&w, r)
	if noteID == 0 {
		aNote = m.NoteNew(map[string]interface{}{
			"ID": noteID,
			"group_id": u.Groups[0].ID,
			"raw_editor": int8(raw_editor),
		})
	} else {
		aNote = m.GetNoteByID(noteID)
	}
	CommonRenderTemplate("frontpage.html", &w, r, &map[string]interface{}{
		"title": "Webnote - note " + aNote.Title,
		"page": "frontpage",
		"msg":  "",
		"note": aNote,
	})
}

func DoSaveNote(w http.ResponseWriter, r *http.Request) {
	msg := "OK note saved"
	user := GetCurrentUser(&w, r)

	r.ParseForm()
	noteID, _ := strconv.ParseInt(m.GetRequestValue(r, "id", "0"), 10, 64)
	ngroup := m.GetGroup(m.GetFormValue(r, "ngroup", "default"))
	_permission, _ := strconv.Atoi(m.GetFormValue(r, "permission", "0"))
	permission := int8(_permission)

	_raw_editor, _ := strconv.Atoi(m.GetFormValue(r, "raw_editor", "0"))
	raw_editor := int8(_raw_editor)

	var aNote *m.Note

	if noteID == 0 {//New note created by current user
		aNote = m.NoteNew(map[string]interface{} {
			"title": r.FormValue("title"),
			"datelog" : r.FormValue("datelog"),
			"flags": r.FormValue("flags"),
			"content": r.FormValue("content"),
			"url": r.FormValue("url"),
			"raw_editor": raw_editor, //If checked return string 1, otherwise empty string
			"permission": permission,
			"author_id": user.ID,
			"group_id": ngroup.ID,
			},
		)
		aNote.Save()
	} else {//Existing note loaded. Need to check permmission
		aNote = m.GetNoteByID(noteID)
		if m.CheckPerm(aNote.Object, user.ID, "w") {
			aNote.Title = r.FormValue("title")
			aNote.Flags = r.FormValue("flags")
			aNote.Content = r.FormValue("content")
			aNote.URL = r.FormValue("url")
			aNote.RawEditor = raw_editor //If checked return string 1, otherwise empty string
			aNote.Permission = permission
			aNote.GroupID = ngroup.ID
			aNote.Save()
		} else {
			msg = "Permission denied."
		}
	}
	isAjax := m.GetRequestValue(r, "is_ajax", "0")
	if isAjax == "1" {
		fmt.Fprintf(w, msg)
	} else {
		if msg != "OK note saved" {
			fmt.Fprintf(w, msg)
		} else {
			http.Redirect(w, r, fmt.Sprintf("/?id=%d", aNote.ID), http.StatusFound)
		}
	}
}

func DoSearchNote(w http.ResponseWriter, r *http.Request) {
	keyword := m.GetRequestValue(r, "keyword", "")
	u := GetCurrentUser(&w, r)
	var attachments []*m.Attachment
	if u != nil {
		attachments = m.SearchAttachement(keyword, u)
	} else {
		attachments = []*m.Attachment{}
	}
	notes := m.SearchNote(keyword, GetCurrentUser(&w, r))
	CommonRenderTemplate("search_result.html", &w, r, &map[string]interface{}{
		"title": "Webnote - Search result",
		"page": "search_result",
		"msg":  "",
		"notes": notes,
		"attachments": attachments,
	})
}

//DoViewNote -
func DoViewNote(w http.ResponseWriter, r *http.Request) {
	viewType := m.GetRequestValue(r, "t", "1")
	tName := "noteview"  + viewType + ".html"
	noteID, _ := strconv.ParseInt(m.GetRequestValue(r, "id", "0"), 10, 64)
	aNote := m.GetNoteByID(noteID)

	if aNote.Permission < 5 {
		isAuth := m.GetSessionVal(r, "authenticated", nil)
		if isAuth == nil || ! isAuth.(bool) {
			log.Printf("ERROR - No session\n")
			from_uri := r.URL.RequestURI()
			from_uri = strings.TrimPrefix(from_uri, "/")
			http.Redirect(w, r, "/login?from_uri=" + from_uri, http.StatusTemporaryRedirect)
			return
		}
		u := GetCurrentUser(&w, r)
		if ! m.CheckPerm(aNote.Object, u.ID, "r") {
			log.Printf("ERROR - No Permission\n")
			fmt.Fprintf(w, "Permission denied")
			return
		}
	}
	data := map[string]interface{}{
		"title": "Webnote - " + aNote.Title,
		"page": "noteview",
		"msg":  "",
		"note": aNote,
		"revisions": m.GetNoteRevisions(aNote.ID),
	}
	if len(aNote.Attachments) > 0 { data["attachments"] = aNote.Attachments }
	CommonRenderTemplate(tName, &w, r, &data)
}

func DoDeleteNote(w http.ResponseWriter, r *http.Request) {
	noteIDStr := m.GetRequestValue(r, "id", "0")
	msg := "OK"
	if noteIDStr != "0" {
		user := GetCurrentUser(&w, r)
		noteID, _ := strconv.ParseInt(noteIDStr, 10, 64)
		aNote := m.GetNoteByID(noteID)
		if m.CheckPerm(aNote.Object, user.ID, "d") {
			aNote.Delete()
		} else {
			msg = "Permission denied."
		}
	} else {
		msg = "Invalid noteID"
	}
	isAjax := m.GetRequestValue(r, "is_ajax", "0")
	page := m.GetRequestValue(r, "page", "0")
	keyword := m.GetRequestValue(r, "keyword", "")
	var redirectURL string
	switch page{
	case "frontpage":
		redirectURL = "/"
	default:
		redirectURL = "/search?keyword=" + keyword
	}
	if isAjax == "1" {
		fmt.Fprintf(w, msg)
	} else {
		if msg != "OK" {
			fmt.Fprintf(w, msg)
		} else {
			http.Redirect(w, r, fmt.Sprintf(redirectURL), http.StatusFound)
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
	http.Redirect(w, r, fmt.Sprintf("/"), http.StatusFound)
}

var CSRF_TOKEN string
//HandleRequests -

func DoViewRevNote(w http.ResponseWriter, r *http.Request) {
	viewType := m.GetRequestValue(r, "t", "1")
	tName := "noteview"  + viewType + ".html"

	noteID, _ := strconv.ParseInt(m.GetRequestValue(r, "id", "0"), 10, 64)
	aNote := m.GetNoteRevisionByID(noteID)
	CommonRenderTemplate(tName, &w, r, &map[string]interface{}{
		"title": "Webnote - " + aNote.Title,
		"page": "noteview",
		"msg":  "",
		"note": aNote,
	})
}

func DoViewDiffNote(w http.ResponseWriter, r *http.Request) {
	noteID, _ := strconv.ParseInt(m.GetRequestValue(r, "id", "0"), 10, 64)
	revNoteID, _ := strconv.ParseInt(m.GetRequestValue(r, "rev_id", "0"), 10, 64)
	n := m.GetNoteByID(noteID)
	n1 := m.GetNoteRevisionByID(revNoteID)
	nd := n1.Diff(n)
	t, _ := template.New("diff").Parse(`<html>
	<body>{{ .htmlText }}</body>
	</html>
	`)
	if e := t.ExecuteTemplate(w, "diff", map[string]interface{} {
		"htmlText": template.HTML(nd.String()),
	} ); e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
}

func DoUpload(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		CommonRenderTemplate("upload.html", &w, r, &map[string]interface{}{
			"title": "Webnote - Upload attachment",
			"page": "upload",
			"msg":  "",
		})
	case "POST":
		user := GetCurrentUser(&w, r)
		if user == nil {
			http.Error(w, "ERROR", http.StatusForbidden)
			return
		}
		var aList [3]*m.Attachment

		if err := r.ParseMultipartForm(m.MaxUploadSizeInMemory); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for count := 1; count <= len(aList) ; count++ {
			cStr := strconv.Itoa(count)
			file, handler, err := r.FormFile("myFile" + cStr )
			if err != nil {
				fmt.Printf("Error Retrieving the File %v\n", err)
				continue
			}
			defer file.Close()
			fmt.Printf("Uploaded File: %+v\n", handler.Filename)
			fmt.Printf("File Size: %+v\n", handler.Size)
			fmt.Printf("MIME Header: %s\n", handler.Header["Content-Type"])
			aName := m.GetRequestValue(r, "a" + cStr, "")
			aDesc := m.GetRequestValue(r, "desc" + cStr, "")
			if aName == "" {
				aName = handler.Filename
			}
			a := m.Attachment{
				Name: aName,
				Description: aDesc,
				AttachedFile: m.UpLoadPath + handler.Filename,
				FileSize: handler.Size,
			}

			f, err := os.OpenFile(a.AttachedFile, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				log.Fatalf("ERROR can not open file to save attachement - %v\n", err)
			}
			copyFile := func(f *os.File, file io.Reader) {
				bWritten, e := io.Copy(f, file)
				if e != nil {
					log.Printf("ERROR Copied %d and got error %v\n", bWritten, e)
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
			_p, _ := strconv.Atoi( m.GetRequestValue(r, "permission", "1"))
			a.Permission = int8(_p)
			a.Save()
			aList[count - 1] = &a
		}
		//Cleanup temp files
		r.MultipartForm.RemoveAll()

		fmt.Fprintf(w, `<html><body><pre>OK Attachment created - +%v</pre>
		<ul>
			<li><a href="/">Home</a></li>
			<li><a href="/upload">More uploads</a></li>
			<li><a href="/list_attachment">List files</a></li>
		</ul>
		</body></html>`, aList)
		return
	}
}

func GetCurrentNote(r *http.Request) *m.Note {
	noteIDStr := m.GetRequestValue(r, "note_id", "")
	if noteIDStr != "" {
		noteID, _ := strconv.ParseInt(noteIDStr, 10, 64)
		return m.GetNoteByID(noteID)
	}
	return nil
}

func DoListAttachment(w http.ResponseWriter, r *http.Request) {
	kw := m.GetRequestValue(r, "keyword", "")
	aNote := GetCurrentNote(r)
	u := GetCurrentUser(&w, r)
	aList := m.SearchAttachement(kw, u)
	data := map[string]interface{}{
		"title": "Webnote - List attachements",
		"page": "list_attachement",
		"msg":  "",
		"attachments": aList,
	}
	if aNote != nil { data["note"] = aNote }
	CommonRenderTemplate("list_attachment.html", &w, r, &data)
}

func DoDeleteAttachment(w http.ResponseWriter, r *http.Request) {
	aIDStr := m.GetRequestValue(r, "id", "")
	if aIDStr == "" { return }
	aID, _ := strconv.ParseInt(aIDStr, 10, 64)
	u := GetCurrentUser(&w, r)
	a := m.GetAttachementByID(aID)
	if e := a.DeleteAttachment(aID, u); e != nil {
		msg := fmt.Sprintf("ERROR Can not delete attachment - %v",e)
		http.Error(w, msg, http.StatusOK)
		return
	}
	is_ajax := m.GetRequestValue(r, "is_ajax", "0")
	if is_ajax == "1" {
		fmt.Fprintf(w, "Deleted attachement ID %d", aID)
	}
}

func DoStreamfile(w http.ResponseWriter, r *http.Request) {
	aIDStr := m.GetRequestValue(r, "id", "")
	if aIDStr == "" {
		http.Error(w, "No aID provided", http.StatusBadRequest)
		return
	}
	aID, _ := strconv.ParseInt(aIDStr, 10, 64)
	a := m.GetAttachementByID(aID)
	u := GetCurrentUser(&w, r)
	if pok := m.CheckPerm(a.Object, u.ID, "r"); ! pok { http.Error(w, "Permission denied", http.StatusBadRequest); return }

	file, err := os.Open(a.AttachedFile)
	if err != nil {
		http.Error(w, "ERROR", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	// stream straight to client(browser)
	w.Header().Set("Content-type", a.Mimetype)
	action := m.GetRequestValue(r, "action", "")

	if action == "download" {
		w.Header().Set("Content-Disposition", "attachment; filename=" + path.Base(a.AttachedFile) )
		http.ServeFile(w, r, a.AttachedFile)

	} else {
		http.ServeContent(w, r, "playing", time.Time{}, file)
	}
}

func DoAttachmentToNote(w http.ResponseWriter, r *http.Request) {
	action := m.GetRequestValue(r, "action", "")
	u := GetCurrentUser(&w, r)
	n := GetCurrentNote(r)
	attachmentIDStr := m.GetRequestValue(r, "attachment_id", "")
	if attachmentIDStr == "" { return }
	attachmentID, _ := strconv.ParseInt(attachmentIDStr, 10, 64)
	if action == "unlink" {
		if e := n.UnlinkAttachment(attachmentID, u); e != nil {
			msg := fmt.Sprintf("ERROR unlink attachment to note - %v\n", e)
			fmt.Fprintf(w, msg)
			return
		}
	} else{
		if e := n.LinkAttachment(attachmentID, u); e != nil {
			log.Printf("ERROR link attachment to note - %v\n", e)
			fmt.Fprintf(w, "ERROR")
			return
		}
	}
	fmt.Fprintf(w, "OK")
}

func HandleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	// router := StaticRouter()
	CSRF_TOKEN := m.MakePassword(32)
	csrf.MaxAge(4 * 3600)
	CSRF := csrf.Protect(
		[]byte(CSRF_TOKEN),
		// instruct the browser to never send cookies during cross site requests
		csrf.SameSite(csrf.SameSiteStrictMode),
		csrf.TrustedOrigins([]string{"note.inxuanthuy.com", "note.xvt.technology"}),
		// csrf.RequestHeader("X-CSRF-Token"),
		// csrf.FieldName("authenticity_token"),
		// csrf.ErrorHandler(http.HandlerFunc(serverError(403))),
	)
	csrf.Secure(true)
	// log.Printf("DEBUG temporary disable csrf %v\n", CSRF)
	router.Use(CSRF)

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
	router.HandleFunc("/view", DoViewNote).Methods("GET")
	router.Handle("/view_rev", isAuthorized(DoViewRevNote)).Methods("GET")
	router.Handle("/view_diff", isAuthorized(DoViewDiffNote)).Methods("GET")
	router.Handle("/delete", isAuthorized(DoDeleteNote)).Methods("POST", "GET")
	router.Handle("/logout", isAuthorized(DoLogout)).Methods("POST", "GET")
	router.Handle("/upload", isAuthorized(DoUpload)).Methods("POST", "GET")
	router.Handle("/list_attachment", isAuthorized(DoListAttachment)).Methods("GET")
	router.Handle("/delete_attachment", isAuthorized(DoDeleteAttachment)).Methods("GET")
	router.Handle("/streamfile", isAuthorized(DoStreamfile)).Methods("GET")
	router.Handle("/add_attachment_to_note", isAuthorized(DoAttachmentToNote)).Methods("GET")
	router.Handle("/delete_note_attachment", isAuthorized(DoAttachmentToNote)).Methods("GET")

	//SinglePage (as note content) handler. Per app the controller file is in app-controllers folder. The javascript app needs to get the token and send it with its post request. Eg. var csrfToken = document.getElementsByName("gorilla.csrf.Token")[0].value
	router.Handle("/cred", isAuthorized(app.DoCredApp)).Methods("POST", "GET")

	//kodi send. Dont need to authenticate via note app but its own (via IP)
	router.Handle("/kodi/add",      app.KodiIsAuthorized(app.HandleAddToPlayList)).Methods("POST")
	router.Handle("/kodi/play",     app.KodiIsAuthorized(app.HandlePlay)).Methods("POST")
	router.Handle("/kodi/loadlist", app.KodiIsAuthorized(app.HandleLoadList)).Methods("POST")
	router.Handle("/kodi/savelist", app.KodiIsAuthorized(app.HandleSaveList)).Methods("POST")
	router.Handle("/kodi/playlist", app.KodiIsAuthorized(app.HandlePlayList)).Methods("POST")
	//Save bookmark
	router.Handle("/savebookmark", isAuthorized(app.SaveBookMark)).Methods("GET")
	router.Handle("/delbookmark", isAuthorized(app.DeleteBookMark)).Methods("GET")

	srv := &http.Server{
		Addr:  ":" + ServerPort,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15000,
		ReadTimeout:  time.Second * 15000,
		IdleTimeout:  time.Second * 60,
		//Handler: handlers.CompressHandler(router), // Pass our instance of gorilla/mux in.
	}
	if *EnableCompression == "yes" {
		srv.Handler = handlers.CompressHandler(router)
	} else {
		srv.Handler = router
	}

	if SSLKey != "" {
		log.Printf("Start SSL/TLS server on port %s\n", ServerPort)
		log.Fatal(srv.ListenAndServeTLS(SSLCert, SSLKey))
	} else {
		log.Printf("Start server on port %s\n", ServerPort)
		log.Fatal(srv.ListenAndServe())
	}
}

func main() {
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

	flag.Usage = func() {
		flag.PrintDefaults()
		msg := `
		Quick start
		To setup initial database use option '-db path-to-your-new-db.db -key path-to-ssl-key -cert path-to-cert-file -p port -setup'

		Next run remove the option -setup
		The default admin email to login is admin@admin.com. To change this use option '-cmd set_admin_email'

		The command after -cmd can be: set_admin_password, set_admin_otp, add_user
		You have to run set_admin_otp to get the OTP QR image produced in the current directory and use it for admin login.

		If a user is blacklisted form an IP (account lock) run sqlite3 path-to-your-db
		Then run this sql 'DELETE FROM appconfig WHERE key="blacklist_ips";'

		Try to select * from appconfig to see what key you can do. The whitelist_ip is a list of IP that user does not need to use OTP to login.
		`
		fmt.Fprintf(os.Stderr, msg)
	}

	flag.Parse()

    ServerPort = *port

	os.Setenv("DBPATH", *dbPath)
	if *setup {
		m.SetupDefaultConfig()
		m.SetupAppDatabase()
		m.CreateAdminUser()
	}

	SSLKey = m.GetConfigSave("ssl_key", *sslKey)
	SSLCert = m.GetConfigSave("ssl_cert", *sslCert)
    m.GetConfigSave("base_url", *base_url)

	m.InitConfig()

	if *cmd != "" {
		//Run command utils
		switch *cmd {
		case "set_admin_password":
			m.SetAdminPassword()
		case "set_admin_otp":
			m.SetAdminOTP()
		case "set_admin_email":
			m.SetAdminEmail("")
		case "add_user":
			m.AddUser(map[string]interface{} {
				"username": *useremail,
				"password": *userpassword,
				"group": *usergroup,
			})
		}
	} else {//Server mode
		if *sessionKey == "" {
			*sessionKey = m.GetConfig("session-key", "")
			if *sessionKey == "" {
				*sessionKey = m.MakePassword(64)
				m.SetConfig("session-key", *sessionKey)
			}
		}
		m.SessionStore = sessions.NewCookieStore([]byte(*sessionKey))
		m.SessionStore.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   3600 * 4,
			HttpOnly: true,
		}
		// log.Println(*sessionKey)
		m.LoadAllTemplates()
		HandleRequests()
	}
}

func DoLogin(w http.ResponseWriter, r *http.Request) {
	trycount := m.GetSessionVal(r, "trycount", 0)
	_userIP := m.ReadUserIP(r)
	portPtn := regexp.MustCompile(`\:[\d]+$`)
	userIP := portPtn.ReplaceAllString(_userIP, "")
	ses, _ := m.SessionStore.Get(r, "auth-session")

	blIP := m.GetConfig("blacklist_ips", "")

	if trycount.(int) > 3 || (strings.Contains(blIP, userIP)) {
		ses.Values = make(map[interface{}]interface{}, 1)
		ses.Save(r, w)
		_ip := strings.Split(_userIP, ":")
		if ! strings.Contains(blIP, _ip[0]){
			m.SetConfig("blacklist_ips", blIP + "," + _ip[0])
		}
		http.Error(w, "Account locked", http.StatusForbidden)
		//To remove the lock use command below on the terminal - adjust the val properly. And clear the browser cookie as well
		//ql -db testwebnote.db  'update appconfig set val = "" where key = "blacklist_ips"'
		return
	}

	trycount = trycount.(int) + 1
	m.SaveSessionVal(r, &w, "trycount", trycount)

	isAuthenticated := m.GetSessionVal(r, "authenticated", nil)

	switch r.Method {
	case "GET":
		if isAuthenticated == nil || ! isAuthenticated.(bool) {
			from_uri := r.URL.RequestURI()
			from_uri = strings.TrimPrefix(from_uri, "/login?from_uri=")
			data := map[string]interface{}{
				csrf.TemplateTag: csrf.TemplateField(r),
				"title": "Webnote",
				"page": "login",
				"msg":  "",
				"client_ip": userIP,
				"from_uri": from_uri,
			}
			if err := m.AllTemplates.ExecuteTemplate(w, "login.html", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			fmt.Fprintf(w, "Already logged in")
		}
	case "POST":
		r.ParseForm()
		useremail := r.FormValue("username")
		password := r.FormValue("password")

		var user *m.User

		// whitelistIP := m.GetConfigSave("white_list_ips", "192.168.0.0/24, 127.0.0.1/8")
		whitelistIP := m.GetConfigSave("white_list_ips", "")
		if ! m.CheckUserIPInWhiteList(userIP, whitelistIP){
			totop := r.FormValue("totp_number")
			log.Printf("INFO user input totp %s\n", totop)
			if totop == "" {
				user = nil
			} else{
				user = m.VerifyLogin(useremail, password, totop)
			}
		} else {
			log.Printf("INFO user IP whitelisted, ignore OTP - ip %s - list: %s\n", userIP, whitelistIP)
			user = m.VerifyLogin(useremail, password, "")
		}
		ses, _ := m.SessionStore.Get(r, "auth-session")
		if user != nil {
			log.Printf("INFO Verified user %v\n", user)
			ses.Values["authenticated"] = true
			m.SaveSessionVal(r, &w, "useremail", useremail)
			ses.Save(r, w)
			from_uri := m.GetRequestValue(r, "from_uri", "")
			from_uri = strings.TrimPrefix(from_uri, "/")
			http.Redirect(w, r, "/" + from_uri, http.StatusFound)
			m.SaveSessionVal(r, &w, "trycount", 0)
			return
		} else {
			log.Printf("INFO Failed To Verify user %s\n", useremail)
			ses.Values["authenticated"] = false
			m.SaveSessionVal(r, &w, "useremail", "")
			ses.Save(r, w)
			if trycount.(int) >= 3 {
				currentBlackList := m.GetConfig("blacklist_ips", "")
				m.SetConfig("blacklist_ips", currentBlackList + "," + userIP)
			}
			http.Error(w, "Failed login", http.StatusForbidden)
			return
		}
	}
}

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuth := m.GetSessionVal(r, "authenticated", nil)
		if isAuth == nil || ! isAuth.(bool) {
			log.Printf("ERROR - No session\n")
			uri := r.RequestURI
			// uri = url.PathEscape(uri)
			uri = strings.TrimPrefix(uri, "/")
			http.Redirect(w, r, "/login?from_uri=" + uri, http.StatusTemporaryRedirect)
			return
		}
		// w.Header().Set("X-CSRF-Token", csrf.Token(r))
		endpoint(w, r)
    })
}

func CommonRenderTemplate(tmplName string, w *http.ResponseWriter, r *http.Request, mapData *map[string]interface{}) {
	user := GetCurrentUser(w, r)
	keyword := m.GetRequestValue(r, "keyword", "")

	var uGroups []*m.Group
	var commonMapData map[string]interface{}

	if user != nil {
		uGroups = user.Groups
		commonMapData = map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(r),
			"keyword": keyword,
			"settings": m.Settings,
			"user": user,
			"groups": uGroups,
			"permission_list": m.PermissionList,
			"date_layout": m.DateLayout,
		}
	} else {
		commonMapData = map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(r),
			"keyword": keyword,
			"settings": m.Settings,
			"permission_list": m.PermissionList,
			"date_layout": m.DateLayout,
		}
	}
	for _k, _v := range(*mapData) {
		commonMapData[_k] = _v
	}

	if err := m.AllTemplates.ExecuteTemplate(*w, tmplName, commonMapData); err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
	}
}
