package main

import (
	"net/url"
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
	"github.com/gorilla/csrf"
	m "github.com/sunshine69/webnote-go/models"
	"github.com/sunshine69/webnote-go/app"
)

var ServerPort, SSLKey, SSLCert string

func init() {
	SSLKey = m.GetConfig("ssl_key", "")
	SSLCert = m.GetConfig("ssl_cert", "")
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	noteID, _ := strconv.ParseInt(m.GetRequestValue(r, "id", "0"), 10, 64)
	raw_editor, _ := strconv.Atoi(m.GetRequestValue(r, "raw_editor", "1"))

	var aNote *m.Note
	if noteID == 0 {
		aNote = m.NoteNew(map[string]interface{}{
			"ID": noteID,
			"group_id": int8(1),
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
	useremail := m.GetSessionVal(r, "useremail", "").(string)
	user := m.GetUser(useremail)

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
			"group_id": ngroup.Group_id,
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
			aNote.GroupID = ngroup.Group_id
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
	notes := m.SearchNote(keyword)

	CommonRenderTemplate("search_result.html", &w, r, &map[string]interface{}{
		"title": "Webnote - Search result",
		"page": "search_result",
		"msg":  "",
		"notes": notes,
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
			from_uri := r.RequestURI
			http.Redirect(w, r, "/login?from_uri=" + from_uri, http.StatusTemporaryRedirect)
			return
		}
	}
	CommonRenderTemplate(tName, &w, r, &map[string]interface{}{
		"title": "Webnote - " + aNote.Title,
		"page": "noteview",
		"msg":  "",
		"note": aNote,
		"revisions": m.GetNoteRevisions(aNote.ID),
	})
}

func DoDeleteNote(w http.ResponseWriter, r *http.Request) {
	noteIDStr := m.GetRequestValue(r, "id", "0")
	msg := "OK"
	if noteIDStr != "0" {
		useremail := m.GetSessionVal(r, "useremail", "").(string)
		user := m.GetUser(useremail)
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
		r.ParseMultipartForm(10 << 20)

		useremail := m.GetSessionVal(r, "useremail", "").(string)
		var user *m.User
		if useremail == "" {
			fmt.Fprintf(w, "ERROR")
			return
		}
		user = m.GetUser(useremail)

		var aList [3]*m.Attachment

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
				AttachedFile: "assets/media/attachments/" + handler.Filename,
			}

			f, err := os.OpenFile(a.AttachedFile, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				log.Fatalf("ERROR can not open file to save attachement - %v\n", err)
			}
			defer f.Close()
			io.Copy(f, file)

			a.AuthorID = user.ID
			a.GroupID = user.Groups[0].Group_id
			a.Mimetype = fmt.Sprintf("+%v", handler.Header["Content-Type"])
			_p, _ := strconv.Atoi( m.GetRequestValue(r, "permission", "1"))
			a.Permission = int8(_p)
			a.Save()
			aList[count - 1] = &a
		}
		fmt.Fprintf(w, "OK Attachment created - +%v", aList)
	}
}

func DoListAttachment(w http.ResponseWriter, r *http.Request) {
	kw := m.GetRequestValue(r, "keyword", "")
	
}

func HandleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	// router := StaticRouter()
	CSRF_TOKEN := m.MakePassword(32)
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
        WriteTimeout: time.Second * 15,
        ReadTimeout:  time.Second * 15,
        IdleTimeout:  time.Second * 60,
		Handler: router, // Pass our instance of gorilla/mux in.

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
		fmt.Fprintf(w, "Account locked")
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
			from_uri := m.GetRequestValue(r, "from_uri", "")
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
			fmt.Fprintf(w, "Failed login")
		}
	}
}

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuth := m.GetSessionVal(r, "authenticated", nil)
		if isAuth == nil || ! isAuth.(bool) {
			log.Printf("ERROR - No session\n")
			uri := r.RequestURI
			http.Redirect(w, r, "/login?from_uri=" + url.PathEscape(uri), http.StatusTemporaryRedirect)
			return
		}
		// w.Header().Set("X-CSRF-Token", csrf.Token(r))
		endpoint(w, r)
    })
}

func CommonRenderTemplate(tmplName string, w *http.ResponseWriter, r *http.Request, mapData *map[string]interface{}) {
	useremail := m.GetSessionVal(r, "useremail", "").(string)
	keyword := m.GetRequestValue(r, "keyword", "")

	var user *m.User
	var uGroups []*m.Group
	var commonMapData map[string]interface{}

	if useremail != "" {
		user = m.GetUser(useremail)
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
