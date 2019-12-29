package models

import (
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
	for _, nwStr := range(listNetwork) {
		nwStr = strings.TrimSpace(nwStr)
		_, netB, _ := net.ParseCIDR(nwStr)
		if netB.Contains(ipA) { return true }
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
