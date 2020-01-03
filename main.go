package main

import (
	"strconv"
	"regexp"
	"strings"
	"bytes"
	"github.com/gorilla/sessions"
	"fmt"
	"os"
	"flag"
	"time"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/gorilla/csrf"
	"github.com/arschles/go-bindata-html-template"
	m "github.com/sunshine69/webnote-go/models"
)

var ServerPort, SSLKey, SSLCert string

func init() {
	ServerPort = m.GetConfig("server_port", "8080")
	SSLKey = m.GetConfig("ssl_key", "")
	SSLCert = m.GetConfig("ssl_cert", "")
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

func HomePage(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("home", Asset).Funcs(*TemplateFuncMap).ParseFiles("assets/templates/header.html", "assets/templates/head_menu.html", "assets/templates/list_note_attachment.html", "assets/templates/footer.html", "assets/templates/frontpage.html")

	if err != nil {
		panic(err)
	}
	useremail := m.GetSessionVal(r, "useremail", "").(string)
	user := m.GetUser(useremail)
	uGroups := user.Groups

	noteID, _ := strconv.ParseInt(GetRequestValue(r, "id", "0"), 10, 64)

	raw_editor, _ := strconv.Atoi(GetRequestValue(r, "raw_editor", "1"))

	var aNote *m.Note
	if noteID == 0 {
		aNote = m.NoteNew(map[string]interface{}{
			"ID": noteID,
			"author_id": user.ID,
			"group_id": int8(1),
			"raw_editor": int8(raw_editor),
		})
	} else {
		aNote = m.GetNoteByID(noteID)
	}

	if err := t.Execute(w, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
		"title": "Webnote",
		"page": "FrontPage",
		"msg":  "",
		"settings": m.Settings,
		"note": aNote,
		"user": user,
		"groups": uGroups,
		"permission_list": m.PermissionList,
		"date_layout": m.DateLayout,

	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func LoadTemplate(tFilePath ...string) (string) {
	var o bytes.Buffer
	for _, f := range(tFilePath) {
		tStringb, _ := Asset(f)
		o.Write(tStringb)
	}
	return o.String()
}

func DoLogin(w http.ResponseWriter, r *http.Request) {
	trycount := m.GetSessionVal(r, "trycount", 0)
	_userIP := ReadUserIP(r)
	portPtn := regexp.MustCompile(`\:[\d]+$`)
	userIP := portPtn.ReplaceAllString(_userIP, "")
	ses, _ := m.SessionStore.Get(r, "auth-session")

	blIP := m.GetConfig("blacklist_ips", "")

	if trycount.(int) > 3 || (strings.Contains(blIP, userIP)) {
		fmt.Fprintf(w, "Account locked")
		ses.Values = make(map[interface{}]interface{}, 1)
		ses.Save(r, w)
		return
	}

	trycount = trycount.(int) + 1
	m.SaveSessionVal(r, &w, "trycount", trycount)

	isAuthenticated := m.GetSessionVal(r, "authenticated", nil)

	switch r.Method {
	case "GET":
		if isAuthenticated == nil || ! isAuthenticated.(bool) {
			t, err := template.New("login", Asset).Funcs(*TemplateFuncMap).ParseFiles("assets/templates/header.html", "assets/templates/login.html")
			if err != nil {
				panic(err)
			}
			if err := t.Execute(w, map[string]interface{}{
				csrf.TemplateTag: csrf.TemplateField(r),
				"title": "Webnote",
				"page": "login",
				"msg":  "",
				"client_ip": userIP,
			}); err != nil {
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

	aNote := m.NoteNew(map[string]interface{} {
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
	if noteID == 0 {//New note created by current user
		aNote.Save()
	} else {//Existing note loaded. Need to check permmission
		aNote.ID = noteID
		if m.CheckPerm(aNote.Object, user.ID, "w") {
			aNote.Save()
		} else {
			msg = "Permission denied."
		}
	}
	is_ajax := GetRequestValue(r, "is_ajax", "0")
	if is_ajax == "1" {
		fmt.Fprintf(w, msg)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/?id=%d", aNote.ID), http.StatusFound)
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

//HandleRequests -
func HandleRequests() {
	router := mux.NewRouter()
	csrfKey := m.MakePassword(32)
	CSRF := csrf.Protect(
		[]byte(csrfKey),
		// instruct the browser to never send cookies during cross site requests
		csrf.SameSite(csrf.SameSiteStrictMode),
	)
	csrf.Secure(true)
	router.Use(CSRF)

	router.PathPrefix("/assets").Handler(http.FileServer(AssetFile()))
	router.HandleFunc("/login", DoLogin).Methods("GET", "POST")
	router.Handle("/", isAuthorized(HomePage)).Methods("GET")

	//All routes handlers
	router.Handle("/savenote", isAuthorized(DoSaveNote)).Methods("POST")

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
	log.Println(*sessionKey)
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
		"template_html": func(html string) template.HTML {
			return template.HTML(html)
		},
		"if_ie": func() template.HTML {
			return template.HTML("<!--[if IE]>")
		},
		"end_if_ie": func() template.HTML {
			return template.HTML("<![endif]-->")
		},
	}
	TemplateFuncMap = &_TemplateFuncMap
	HandleRequests()
}
