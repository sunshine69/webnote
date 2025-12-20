package models

import (
	"bufio"
	"bytes"
	"fmt"
	"image/png"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/pquerna/otp/totp"
	u "github.com/sunshine69/golang-tools/utils"
	terminal "golang.org/x/term"
)

// CheckUserIPInWhiteList - whitelist is a string coma sep list of network
func CheckUserIPInWhiteList(ip, whitelist string) bool {
	listNetwork := strings.Split(whitelist, ",")
	portPtn := regexp.MustCompile(`\:[\d]+$`)
	host := portPtn.ReplaceAllString(ip, "")
	ipA := net.ParseIP(host)
	if len(listNetwork) == 0 {
		return false
	}
	for _, nwStr := range listNetwork {
		nwStr = strings.TrimSpace(nwStr)
		_, netB, _ := net.ParseCIDR(nwStr)
		if (netB != nil) && netB.Contains(ipA) {
			return true
		}
	}
	return false
}

func AddUser(in map[string]interface{}) {
	reader := bufio.NewReader(os.Stdin)
	useremail := u.MapLookup(in, "Email", "").(string)
	password := u.MapLookup(in, "Password", "").(string)
	groupStr := u.MapLookup(in, "Group", "").(string)

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

	user := User{
		Email: useremail,
	}
	user.SetUserPassword(password)
	user.SetGroup(groups...)
	SetUserOTP(useremail)
}

func SetAdminPassword() {
	u := GetUser(Settings.ADMIN_EMAIL)
	fmt.Printf("please type in the password (mandatory): ")
	password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
	u.SetUserPassword(string(password))
}

func SetAdminEmail(email string) {
	if email == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("\nEnter user email: ")
		email, _ = reader.ReadString('\n')
		email = strings.Replace(email, "\n", "", -1)
	}
	SetConfig("admin_email", email)
}

func SetUserOTP(username string) *bytes.Buffer {
	u := GetUser(username)
	Issuer := Settings.BASE_URL
	if Issuer == "" {
		Issuer = strings.Split(u.Email, `@`)[1]
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      Issuer,
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
	os.WriteFile(filename, buf.Bytes(), 0600)
	fmt.Printf("PNG QR encoded file name %s has been generated in the current folder\n", filename)
	// Now Validate that the user's successfully added the passcode.
	// fmt.Printf("The OTP Sec is: '%s'\n", key.Secret())
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

func GetRequestValue(r *http.Request, key ...string) (value string) {
	value = r.FormValue(key[0])
	if value == "" {
		value = r.PathValue(key[0])
	}
	if value == "" && len(key) > 1 {
		value = key[1]
	}
	return
}
