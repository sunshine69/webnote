package main

import (
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
	t, err := template.New("home", Asset).ParseFiles("assets/templates/header.html", "assets/templates/head_menu.html", "assets/templates/list_note_attachment.html", "assets/templates/footer.html", "assets/templates/frontpage.html")
	if err != nil {
		panic(err)
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
			t, err := template.New("login", Asset).ParseFiles("assets/templates/header.html", "assets/templates/login.html")
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
		// m.SetConfig("white_list_ips", "192.168.0.0/24, 127.0.0.1/8")

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
	}

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
	//Template custom functions
	_TemplateFuncMap := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"inc": func(i int) int {
			return i + 1
		},
		"add": func(x, y int) int {
			return x + y
		},
	}
	TemplateFuncMap = &_TemplateFuncMap
	HandleRequests()
}
