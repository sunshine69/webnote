package main

import (
	"bytes"
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
	"html/template"
	"github.com/yuin/goldmark"
	"github.com/microcosm-cc/bluemonday"
	m "github.com/sunshine69/webnote-go/models"
)

var ServerPort, SSLKey, SSLCert string

func init() {
	ServerPort = m.GetConfig("server_port", "8080")
	SSLKey = m.GetConfig("ssl_key", "")
	SSLCert = m.GetConfig("ssl_cert", "")
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	noteID, _ := strconv.ParseInt(GetRequestValue(r, "id", "0"), 10, 64)
	raw_editor, _ := strconv.Atoi(GetRequestValue(r, "raw_editor", "1"))

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
	msg := "Note saved"
	useremail := m.GetSessionVal(r, "useremail", "").(string)
	user := m.GetUser(useremail)

	r.ParseForm()
	noteID, _ := strconv.ParseInt(GetRequestValue(r, "id", "0"), 10, 64)
	ngroup := m.GetGroup(GetFormValue(r, "ngroup", "default"))
	_permission, _ := strconv.Atoi(GetFormValue(r, "permission", "0"))
	permission := int8(_permission)

	_raw_editor, _ := strconv.Atoi(GetFormValue(r, "raw_editor", "0"))
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
	isAjax := GetRequestValue(r, "is_ajax", "0")
	if isAjax == "1" {
		fmt.Fprintf(w, msg)
	} else {
		if msg != "Note saved" {
			fmt.Fprintf(w, msg)
		} else {
			http.Redirect(w, r, fmt.Sprintf("/?id=%d", aNote.ID), http.StatusFound)
		}
	}
}

func ReadUserIP(r *http.Request) string {
    IPAddress := r.Header.Get("X-Real-Ip")
    if IPAddress == "" {
        IPAddress = r.Header.Get("X-Forwarded-For")
    }
    if IPAddress == "" {
		IPAddress = r.RemoteAddr
    }
    return IPAddress
}

func DoSearchNote(w http.ResponseWriter, r *http.Request) {
	keyword := GetRequestValue(r, "keyword", "")
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
	viewType := GetRequestValue(r, "t", "1")
	tName := "noteview"  + viewType + ".html"

	noteID, _ := strconv.ParseInt(GetRequestValue(r, "id", "0"), 10, 64)
	aNote := m.GetNoteByID(noteID)
	CommonRenderTemplate(tName, &w, r, &map[string]interface{}{
		"title": "Webnote - " + aNote.Title,
		"page": "noteview",
		"msg":  "",
		"note": aNote,
	})
}

func DoDeleteNote(w http.ResponseWriter, r *http.Request) {
	noteIDStr := GetRequestValue(r, "id", "0")
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
	isAjax := GetRequestValue(r, "is_ajax", "0")
	page := GetRequestValue(r, "page", "0")
	keyword := GetRequestValue(r, "keyword", "")
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

//HandleRequests -
func HandleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	// router := StaticRouter()
	csrfKey := m.MakePassword(32)
	CSRF := csrf.Protect(
		[]byte(csrfKey),
		// instruct the browser to never send cookies during cross site requests
		csrf.SameSite(csrf.SameSiteStrictMode),
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
	router.Handle("/view", isAuthorized(DoViewNote)).Methods("GET")
	router.Handle("/delete", isAuthorized(DoDeleteNote)).Methods("POST", "GET")
	router.Handle("/logout", isAuthorized(DoLogout)).Methods("POST", "GET")

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

//TemplateFuncMap - custom template func map
var TemplateFuncMap *template.FuncMap

func main() {
	dbPath := flag.String("db", "", "Application DB path")
	sessionKey := flag.String("sessionkey", "", "Session Key")
	setup := flag.Bool("setup", false, "Run initial setup DB and config")
	sslKey := flag.String("key", "", "SSL Key path")
	sslCert := flag.String("cert", "", "SSL Cert path")
	flag.Parse()

	os.Setenv("DBPATH", *dbPath)
	if *setup {
		m.SetupDefaultConfig()
		m.SetupAppDatabase()
		m.CreateAdminUser()
	}

	m.InitConfig()

	SSLKey = m.GetConfigSave("ssl_key", *sslKey)
	SSLCert = m.GetConfigSave("ssl_cert", *sslCert)

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
	//Template custom functions
	_TemplateFuncMap := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"inc": func(i int) int {
			return i + 1
		},
		"add": func(x, y int) int {
			return x + y
		},
		"time_fmt": func(timelayout string, timeticks int64) string {
			return m.NsToTime(timeticks).Format(timelayout)
		},
		"raw_html": func(html string) template.HTML {
			cleanupBytes := bluemonday.UGCPolicy().SanitizeBytes([]byte(html))
			return template.HTML(cleanupBytes)
		},
		"unsafe_raw_html": func(html string) template.HTML {
			return template.HTML(html)
		},
		"if_ie": func() template.HTML {
			return template.HTML("<!--[if IE]>")
		},
		"end_if_ie": func() template.HTML {
			return template.HTML("<![endif]-->")
		},
		"truncatechars": func(length int, in string) template.HTML {
			return template.HTML(m.ChunkString(in, length)[0])
		},
		"cycle": func(idx int, vals ...string) template.HTML {
			_idx := idx % len(vals)
			return template.HTML(vals[_idx])
		},
		"md2html": func(md string) template.HTML {
			var buf bytes.Buffer
			if err := goldmark.Convert([]byte(md), &buf); err != nil {
				panic(err)
			}
			cleanupBytes := bluemonday.UGCPolicy().SanitizeBytes(buf.Bytes())
			return template.HTML( cleanupBytes )
		},
	}
	TemplateFuncMap = &_TemplateFuncMap
	LoadAllTemplates()
	HandleRequests()
}

// func LoadTemplate(tFilePath ...string) (string) {
// 	var o bytes.Buffer
// 	for _, f := range(tFilePath) {
// 		tStringb, _ := Asset(f)
// 		o.Write(tStringb)
// 	}
// 	return o.String()
// }

var AllTemplates *template.Template

func LoadAllTemplates() {
	t, err := template.New("templ").Funcs(*TemplateFuncMap).ParseGlob("assets/templates/*.html")
	if err != nil {
		log.Fatalf("ERROR can not parse templates %v\n", err)
	}
	AllTemplates = t
}

func DoLogin(w http.ResponseWriter, r *http.Request) {
	trycount := m.GetSessionVal(r, "trycount", 0)
	_userIP := ReadUserIP(r)
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
			data := map[string]interface{}{
				csrf.TemplateTag: csrf.TemplateField(r),
				"title": "Webnote",
				"page": "login",
				"msg":  "",
				"client_ip": userIP,
			}
			if err := AllTemplates.ExecuteTemplate(w, "login.html", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			fmt.Fprintf(w, "Already logged in")
		}
	case "POST":
		r.ParseForm()
		useremail := r.FormValue("username")
		m.SaveSessionVal(r, &w, "useremail", useremail)

		password := r.FormValue("password")

		var user *m.User

		whitelistIP := m.GetConfigSave("white_list_ips", "192.168.0.0/24, 127.0.0.1/8")
		if ! m.CheckUserIPInWhiteList(userIP, whitelistIP){
			totop := r.FormValue("totp_number")
			log.Printf("INFO input totp %s\n", totop)
			user = m.VerifyLogin(useremail, password, totop)
		} else {
			user = m.VerifyLogin(useremail, password, "")
		}
		ses, _ := m.SessionStore.Get(r, "auth-session")
		if user != nil {
			log.Printf("INFO Verified user %v\n", user)
			ses.Values["authenticated"] = true
			ses.Save(r, w)
			http.Redirect(w, r, "/", http.StatusFound)
			m.SaveSessionVal(r, &w, "trycount", 0)
			return
		} else {
			log.Printf("INFO Failed To Verify user %s\n", useremail)
			if trycount.(int) >= 3 {
				ses.Values["authenticated"] = nil
				ses.Save(r, w)
				currentBlackList := m.GetConfig("blacklist_ips", "")
				m.SetConfig("blacklist_ips", currentBlackList + "," + userIP)
			}
			fmt.Fprintf(w, "Failed login")
		}
	}
}
//GetRequestValue - Attempt to get a val by key from the request in all cases.
//First from the mux variables in the route path such as /dosomething/{var1}/{var2}
//Then check the query string values such as /dosomething?var1=x&var2=y
//Then check the form values if any
//Then check the default value if supplied to use as return value
//For performance we split each type into each function so it can be called independantly
func GetRequestValue(r *http.Request, key ...string) string {
	o := GetMuxValue(r, key[0], "")
	if o == "" {
		o = GetQueryValue(r, key[0], "")
	}
	if o == "" {
		o = GetFormValue(r, key[0], "")
	}
	if o == "" {
		if len(key) > 1 {
			o = key[1]
		} else {
			o = ""
		}
	}
	return o
}

//GetMuxValue -
func GetMuxValue(r *http.Request, key ...string) string {
	vars := mux.Vars(r)
	val, ok := vars[key[0]]
	if !ok {
		if len(key) > 1 {
			return key[1]
		}
		return ""
	}
	return val
}

//GetFormValue -
func GetFormValue(r *http.Request, key ...string) string {
	val := r.FormValue(key[0])
	if val == "" {
		if len(key) > 1 {
			return key[1]
		}
	}
	return val
}

//GetQueryValue -
func GetQueryValue(r *http.Request, key ...string) string {
	vars := r.URL.Query()
	val, ok := vars[key[0]]
	if !ok {
		if len(key) > 1 {
			return key[1]
		}
		return ""
	}
	return val[0]
}

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		useremail := m.GetSessionVal(r, "useremail", nil)
		if useremail == nil {
			log.Printf("ERROR - No session\n")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		endpoint(w, r)
    })
}

func CommonRenderTemplate(tmplName string, w *http.ResponseWriter, r *http.Request, mapData *map[string]interface{}) {
	useremail := m.GetSessionVal(r, "useremail", "").(string)
	user := m.GetUser(useremail)
	uGroups := user.Groups
	keyword := GetRequestValue(r, "keyword", "")
	commonMapData := map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
		"keyword": keyword,
		"settings": m.Settings,
		"user": user,
		"groups": uGroups,
		"permission_list": m.PermissionList,
		"date_layout": m.DateLayout,
	}

	for _k, _v := range(*mapData) {
		commonMapData[_k] = _v
	}

	if err := AllTemplates.ExecuteTemplate(*w, tmplName, commonMapData); err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
	}
}