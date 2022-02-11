package app

import (
	"fmt"
	"github.com/json-iterator/go"
	"github.com/sunshine69/kodirpc"
	m "github.com/sunshine69/webnote-go/models"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var json jsoniter.API
var kodiURL string

func init() {
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	kodiURL = `127.0.0.1:9090`
}

func GetKodiClient() *kodirpc.Client {
	client, err := kodirpc.NewClient(kodiURL, kodirpc.NewConfig())
	if err != nil || client == nil {
		panic(err)
	}
	return client
}

//GetCurrentPlayList -
func GetCurrentPlayList(playerID int) int {
	client := GetKodiClient()
	defer client.Close()

	res, err := client.Call(`Player.GetProperties`, map[string]interface{}{
		"playerid":   playerID,
		"properties": []string{"playlistid"},
	})
	if err != nil {
		panic(err)
	}
	o, _ := json.Marshal(res)
	return json.Get(o, "playlistid").ToInt()
	// return res.(map[string]interface{})["playlistid"].(float64)
}

func ClearCurrentList(listID int) {
	client := GetKodiClient()
	defer client.Close()
	res, err := client.Call(`Playlist.Clear`, map[string]interface{}{
		"playlistid": listID,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func AddToPlayList(listID int, entry string) {
	kodiYoutubeUrl := ParseVideoURL(entry)

	client := GetKodiClient()
	defer client.Close()
	res, err := client.Call(`Playlist.Add`, map[string]interface{}{
		"playlistid": listID,
		"item": []map[string]string{
			{
				"file": kodiYoutubeUrl,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func InsertToPlayList(listID int, entry string, position int) {
	kodiYoutubeUrl := ParseVideoURL(entry)

	client := GetKodiClient()
	defer client.Close()
	res, err := client.Call(`Playlist.Insert`, map[string]interface{}{
		"playlistid": listID,
		"position":   position,
		"item": []map[string]string{
			{
				"file": kodiYoutubeUrl,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

//GetActivePlayer -
func GetActivePlayer() int {
	client := GetKodiClient()
	defer client.Close()

	res, err := client.Call(`Player.GetActivePlayers`, map[string]interface{}{})
	if err != nil {
		panic(err)
	}
	//First use json Marshal and print the string. Then define this type. Not sure what is better and cleaner way to convert cast it though.
	type Player struct {
		Playerid   int
		Playertype string
		Type       string
	}

	o, _ := json.Marshal(res)
	o1 := json.Get(o, 0, "playerid")
	//If player not started it return 0
	return o1.ToInt()
}

func ParseVideoURL(url string) string {
	outputURL := ""
	found := false

	patterns := map[string]*regexp.Regexp{
		"youtube": regexp.MustCompile(`youtube\.com\/watch\?v\=([^\=\&]+)`),
		"vimeo":   regexp.MustCompile(`vimeo\.com\/([\d]+)`),
	}
	for vtype, ptn := range patterns {
		match := ptn.FindStringSubmatch(url)
		if len(match) > 0 {
			found = true
			vid := match[1]
			switch vtype {
			case "youtube":
				//old
				// outputURL = "plugin://plugin.video.youtube/?action=play_video&videoid=" + vid
				//new
				outputURL = "plugin://plugin.video.youtube/play/?video_id=" + vid
			case "vimeo":
				outputURL = "plugin://plugin.video.vimeo/play/?video_id=" + vid
			}
			break
		}
	}
	if !found {
		outputURL = url
	}
	// fmt.Println(outputURL)
	return outputURL
}

func PlayYoutube(url string) {
	client := GetKodiClient()
	defer client.Close()
	youtubeUrl := ParseVideoURL(url)

	if youtubeUrl != "" {
		params := map[string]interface{}{
			"item": map[string]string{
				"file": youtubeUrl,
			},
		}
		res, err := client.Call(`Player.Open`, params)
		if err != nil {
			panic(err)
		}
		fmt.Println(res)
	}
}

func HandleAddToPlayList(w http.ResponseWriter, r *http.Request) {
	url := ParseCommon(w, r)
	positionstr := r.FormValue("position")
	SaveRecentList(url)

	playerID := GetActivePlayer()
	listID := GetCurrentPlayList(playerID)
	if playerID == 0 {
		// log.Printf("DEBUG playserid %d\n", playerID)
		PlayYoutube(url)
	} else {
		if positionstr == "" {
			AddToPlayList(listID, url)
		} else {
			position, _ := strconv.Atoi(positionstr)
			InsertToPlayList(listID, url, position)
		}
	}
	fmt.Fprintf(w, "OK")
}

func ParseCommon(w http.ResponseWriter, r *http.Request) string {
	r.ParseForm()
	if _kodiURL := r.FormValue("kodi_addr"); _kodiURL != "" {
		kodiURL = _kodiURL
	}
	return r.FormValue("url")
}

func HandlePlay(w http.ResponseWriter, r *http.Request) {
	url := ParseCommon(w, r)
	SaveRecentList(url)

	PlayYoutube(url)
	fmt.Fprintf(w, "OK")
}

func HandleLoadList(w http.ResponseWriter, r *http.Request) {
	ParseCommon(w, r)
	listName := r.FormValue("list_name")
	data, e := ioutil.ReadFile(listName + ".list")
	if e != nil {
		fmt.Fprintf(w, "ERROR list does not exists")
	} else {
		fmt.Fprintf(w, string(data))
	}
}

func HandleSaveList(w http.ResponseWriter, r *http.Request) {
	ParseCommon(w, r)
	list_text := r.FormValue("list_text")
	list_name := r.FormValue("list_name")
	ioutil.WriteFile(list_name+".list", []byte(list_text), 0755)
	fmt.Fprintf(w, "OK")
}

func HandlePlayList(w http.ResponseWriter, r *http.Request) {
	ParseCommon(w, r)
	data := r.FormValue("list_text")
	action := r.FormValue("action")
	listUrls := strings.Split(data, "\n")
	playerID := GetActivePlayer()
	listID := GetCurrentPlayList(playerID)
	// log.Printf("DEBUG: ListID %d\n", listID)
	if action == "play" {
		ClearCurrentList(listID)
	}
	for _, url := range listUrls {
		_tmp := strings.Split(url, " ")
		if _tmp[0] == "" {
			continue
		}
		if playerID == 0 {
			PlayYoutube(_tmp[0])
			time.Sleep(5 * time.Second)
			playerID = GetActivePlayer()
			listID = GetCurrentPlayList(playerID)
		} else {
			AddToPlayList(listID, _tmp[0])
		}
	}
	fmt.Fprintf(w, "OK")
}

func SaveRecentList(url string) {
	f, err := os.OpenFile("recent.list", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(url + "\n"); err != nil {
		log.Println(err)
	}
}

func KodiIsAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIP := m.ReadUserIP(r)
		kodiNetwork := m.GetConfigSave("kodi_network", "127.0.0.1/8, 192.168.0.0/24")
		if !m.CheckUserIPInWhiteList(userIP, kodiNetwork) {
			fmt.Fprintf(w, "ERROR NOT ALLOWED NETWORK")
			return
		}
		// w.Header().Set("X-CSRF-Token", csrf.Token(r))
		endpoint(w, r)
	})
}
