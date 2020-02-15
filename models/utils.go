package models

import (
	"bufio"
	"io/ioutil"
	"bytes"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/ssh/terminal"
	"github.com/pquerna/otp/totp"
	"net/http"
	"time"
	"regexp"
	"net"
	"strings"
	"encoding/base64"
	"crypto/sha512"
	"log"
	"encoding/binary"
	crand "crypto/rand"
	rand "math/rand"
)

//GetMapByKey -
func GetMapByKey(in map[string]interface{}, key string, defaultValue interface{}) interface{} {
	// log.Printf("%v - %v - %v\n", in, key, defaultValue )
	var o interface{}
	v, ok := in[key]
	if !ok {
		o = defaultValue
	} else {
		o = v
	}
	// log.Printf("RETURN: %v\n", o)
	return o
}

//MakeRandNum -
func MakeRandNum(max int) int {
    var src cryptoSource
    rnd := rand.New(src)
	// fmt.Println(rnd.Intn(1000)) // a truly random number 0 to 999
	return rnd.Intn(max)
}

type cryptoSource struct{}

func (s cryptoSource) Seed(seed int64) {}

func (s cryptoSource) Int63() int64 {
    return int64(s.Uint64() & ^uint64(1<<63))
}

func (s cryptoSource) Uint64() (v uint64) {
    err := binary.Read(crand.Reader, binary.BigEndian, &v)
    if err != nil {
        log.Fatal(err)
    }
    return v
}

//MakePassword -
func MakePassword(length int) string {
	b := make([]byte, length)
	// seededRand := rand.New(rand.NewSource(time.Now().UnixNano() ))
	const charset = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_-+=`
	for i := range b {
	  b[i] = charset[MakeRandNum(len(charset))]
	}
	return string(b)
}

func ComputeHash(plainText string , salt []byte) (string) {
	plainTextWithSalt := []byte(plainText)
	plainTextWithSalt =  append(plainTextWithSalt, salt...)
	sha_512 := sha512.New()
	sha_512.Write(plainTextWithSalt)
	out := sha_512.Sum(nil)
	out = append(out, []byte(salt)...)
	return base64.StdEncoding.EncodeToString(out)
}

func VerifyHash(password string, passwordHashString string, saltLength int) bool {
	// log.Printf("DEBUG VerifyHash input pass: %s - Hash %s s_len %d\n", password, passwordHashString, saltLength)
	passwordHash, _ := base64.StdEncoding.DecodeString(passwordHashString)
	saltBytes := []byte(passwordHash[len(passwordHash) - saltLength:len(passwordHash)])
	result := ComputeHash(password, saltBytes)
	return result == passwordHashString
}

func MakeSalt(length int8) (salt *[]byte) {
	asalt := make([]byte, length)
	crand.Read(asalt)
	return &asalt
}

//CheckUserIPInWhiteList - whitelist is a string coma sep list of network
func CheckUserIPInWhiteList(ip, whitelist string) (bool) {
	listNetwork := strings.Split(whitelist, ",")
	portPtn := regexp.MustCompile(`\:[\d]+$`)
	host := portPtn.ReplaceAllString(ip, "")
	ipA := net.ParseIP(host)
	if len(listNetwork) == 0 {return false}
	for _, nwStr := range(listNetwork) {
		nwStr = strings.TrimSpace(nwStr)
		_, netB, _ := net.ParseCIDR(nwStr)
		if (netB != nil) && netB.Contains(ipA) { return true }
	}
	return false
}

//Time handling
const (
	millisPerSecond     = int64(time.Second / time.Millisecond)
	nanosPerMillisecond = int64(time.Millisecond / time.Nanosecond)
	nanosPerSecond      = int64(time.Second / time.Nanosecond)
)

//NsToTime -
func NsToTime(ns int64) time.Time  {
	secs := ns/nanosPerSecond
	nanos := ns - secs * nanosPerSecond
	return time.Unix(secs, nanos)
}

//ChunkString -
func ChunkString(s string, chunkSize int) []string {
	var chunks []string
	runes := []rune(s)

	if len(runes) == 0 {
		return []string{s}
	}
	for i := 0; i < len(runes); i += chunkSize {
		nn := i + chunkSize
		if nn > len(runes) {
			nn = len(runes)
		}
		chunks = append(chunks, string(runes[i:nn]))
	}
	return chunks
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

func AddUser(in map[string]interface{}) {
	reader := bufio.NewReader(os.Stdin)
	useremail := GetMapByKey(in, "Email", "").(string)
	password := GetMapByKey(in, "Password", "").(string)
	groupStr := GetMapByKey(in, "Group", "").(string)

	if useremail == "" {
		fmt.Printf("\nEnter user email: ")
		useremail, _ = reader.ReadString('\n')
		useremail = strings.Replace(useremail, "\n", "", -1)
	}
	if password == "" {
		fmt.Printf("\nEnter user password: ")
		passwordByte, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
		password = string(passwordByte)
	}
	if groupStr == "" {
		fmt.Printf("\nEnter user group separated by coma (default|family|friend): ")
		groupStr, _ = reader.ReadString('\n')
		groupStr = strings.Replace(groupStr, "\n", "", -1)
	}
	groups := strings.Split(groupStr, ",")

	user := UserNew(map[string]interface{} {
		"Email": useremail,
	})
	user.Save()
	user.SetGroup(groups...)
	user.SetUserPassword(password)
	SetUserOTP(useremail)
}

func SetAdminPassword() {
	u := GetUser(Settings.ADMIN_EMAIL)
	fmt.Printf("please type in the password (mandatory): ")
	password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
	// log.Printf("DEBUG pass %s\n", string(password))
	u.SetUserPassword(string(password))
}

func SetAdminEmail(email string) {
	if (email == ""){
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("\nEnter user email: ")
		email, _ := reader.ReadString('\n')
		email = strings.Replace(email, "\n", "", -1)
	}
	SetConfig("admin_email", email)
}

func SetUserOTP(username string) *bytes.Buffer {
	u := GetUser(username)
	Issuer := Settings.BASE_URL
	if Issuer == "" {
		Issuer =  strings.Split(u.Email, `@`)[1]
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer: Issuer,
		AccountName: u.Email,
	})
	if err != nil {
		panic(err)
	}
	// Convert TOTP key into a PNG
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		panic(err)
	}
	png.Encode(&buf, img)

	// display the QR code to the user.
	filename := fmt.Sprintf("%s@%s-OTP.png", u.Email, Issuer)
	ioutil.WriteFile(filename, buf.Bytes(), 0600)
	fmt.Printf("PNG QR encoded file name %s has been generated in the current folder\n", filename)
	// Now Validate that the user's successfully added the passcode.
	fmt.Printf("The OTP Sec is: '%s'\n", key.Secret())
	u.TotpPassword = key.Secret()
	u.Save()
	u.SaveUserOTP()
	return &buf
}

func SetAdminOTP() {
	SetUserOTP(Settings.ADMIN_EMAIL)
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

func ZipEncript(filePath ...string) string {
	src, dest, key := filePath[0], "", ""
	argCount := len(filePath)
	if argCount > 1 {
		dest = filePath[1]
	} else {
		dest = src + ".zip"
	}
	if argCount > 2 {
		key = filePath[2]
	} else {
		key = MakePassword(42)
	}
	os.Remove(dest)
	srcDir := filepath.Dir(src)
	srcName := filepath.Base(src)
	absDest, _ := filepath.Abs(dest)

	fmt.Printf("DEBUG srcDir %s - srcName %s\n", srcDir, srcName)
	cmd := exec.Command("/bin/sh", "-c", "cd " + srcDir + "; /usr/bin/zip -r -e -P '" +  key + "' " + absDest + " " + srcName)
	// fmt.Println(cmd.String())
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	return key
}

func Ternary(cond bool, first, second interface{}) interface{} {
	if cond {
		return first
	} else {
		return second
	}
}