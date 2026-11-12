package app

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"

	u "github.com/sunshine69/golang-tools/utils"
	m "github.com/sunshine69/webnote-go/models"
)

// onetimeDisplayTemplate renders the secret for the browser.
//
// #gotmpl:{{ }}
//
// The secret is injected as a base64 string (alphabet A-Za-z0-9+/= only), which
// is inert in an HTML text node, an HTML attribute and a JavaScript string
// literal. The exact UTF-8 bytes are reconstructed client-side and rendered via
// textContent, so special characters (<, >, &, ", ', quotes, non-ASCII, etc.)
// are never consumed as markup and never silently lost.
const onetimeDisplayTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>One-Time Secret</title>
    <style>
        body {
            margin: 0;
            font-family: sans-serif;
            background-color: #f5f5f5;
        }

        .container {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: flex-start;
            min-height: 100vh;
            padding-top: 40px;
        }

        #secretDisplay {
            width: 80%;
            max-width: 600px;
            min-height: 100px;
            font-size: 16px;
            padding: 15px;
            border: 1px solid #ccc;
            border-radius: 4px;
            white-space: pre-wrap;
            word-wrap: break-word;
            word-break: break-all;
            overflow-wrap: break-word;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
            background-color: white;
            box-sizing: border-box;
            line-height: 1.4;
        }

        .btn-row {
            display: flex;
            gap: 12px;
            flex-wrap: wrap;
            justify-content: center;
        }

        button {
            margin-top: 20px;
            padding: 10px 20px;
            font-size: 16px;
            background-color: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            transition: background-color 0.3s ease;
        }

        button:hover {
            background-color: #0056b3;
        }

        .copy-message {
            margin-top: 10px;
            padding: 8px 16px;
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
            border-radius: 4px;
            font-size: 14px;
            animation: fadeIn 0.3s ease-in;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(-10px); }
            to { opacity: 1; transform: translateY(0); }
        }
    </style>
</head>
<body>
    <div class="container">
        <div id="secret-b64" style="display:none">{{ .SecretB64 }}</div>
        <div id="secretDisplay"></div>
        <div class="btn-row">
            <button id="toggleBtn" onclick="toggleSecret()">Reveal</button>
            <button onclick="copyToClipboard()">Copy to Clipboard</button>
        </div>
    </div>

    <script>
        // Secret bytes are delivered as standard base64 in a hidden element.
        // Standard base64 (A-Za-z0-9+/=) is inert in HTML/JS; HTML entity
        // escaping (e.g. + -> &#43;) is transparently decoded by textContent,
        // so the browser sees the exact original UTF-8 bytes of the secret.
        var secretB64 = document.getElementById("secret-b64").textContent.trim();
        var secretText = atob(secretB64);
        var secretBytes = Uint8Array.from(secretText, function (c) { return c.charCodeAt(0); });
        var secretRaw = new TextDecoder("utf-8").decode(secretBytes);
        var secretMask = "\u2022".repeat(Math.min(secretRaw.length, 12));
        var revealed = true;

        var display = document.getElementById("secretDisplay");
        var toggleBtn = document.getElementById("toggleBtn");

        function applyState() {
            display.textContent = revealed ? secretRaw : secretMask;
            display.setAttribute("aria-live", "polite");
            toggleBtn.textContent = revealed ? "Hide" : "Reveal";
        }

        function toggleSecret() {
            revealed = !revealed;
            applyState();
        }

        // Show the raw secret by default so nothing is ever lost on reveal.
        applyState();

        function copyToClipboard() {
            navigator.clipboard.writeText(secretRaw).then(function () {
                var existingMessage = document.querySelector('.copy-message');
                if (existingMessage) {
                    existingMessage.remove();
                }

                var messageDiv = document.createElement('div');
                messageDiv.textContent = 'Content copied to clipboard!';
                messageDiv.className = 'copy-message';
                document.querySelector('.container').appendChild(messageDiv);

                setTimeout(function () {
                    if (messageDiv.parentNode) {
                        messageDiv.remove();
                    }
                }, 3000);
            }).catch(function (err) {
                console.error('Failed to copy: ', err);

                var errorDiv = document.createElement('div');
                errorDiv.textContent = 'Failed to copy to clipboard';
                errorDiv.style.marginTop = '10px';
                errorDiv.style.color = '#dc3545';
                document.querySelector('.container').appendChild(errorDiv);

                setTimeout(function () {
                    if (errorDiv.parentNode) {
                        errorDiv.remove();
                    }
                }, 3000);
            });
        }
    </script>
</body>
</html>
`

func GenerateOnetimeSecURL(w http.ResponseWriter, r *http.Request) {
	submit_type := m.GetRequestValue(r, "submit", "")
	base_url := m.GetRequestValue(r, "base_url", "")
	var secret string
	if submit_type == "submit_genpass" {
		length_str := m.GetRequestValue(r, "password_len", "12")
		length_str = strings.TrimSpace(length_str)
		password_len, err := strconv.Atoi(length_str)
		if u.CheckErrNonFatal(err, "GenerateOnetimeSecURL") != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"ERROR length should be a integer"}`)
			return
		}
		secret = u.GenRandomString(password_len)
	} else {
		secret = m.GetRequestValue(r, "sec_content", "")
	}
	if secret == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"ERROR secret content is empty"}`)
		return
	}
	var anote *m.Note = nil
	var note_title string
	for {
		gen_number, _ := rand.Int(rand.Reader, big.NewInt(922337203685477580))
		note_title = fmt.Sprintf("%d", gen_number)
		// check if we already have a note with this title - if we do then loop to generate a new title; otherwise exit this loop
		anote = m.GetNote(note_title)
		if anote == nil {
			break
		}
	}
	secnote := m.NoteNew(m.Note{
		Title:     note_title,
		Content:   secret,
		RawEditor: 1,
		Object: m.Object{
			Permission: 0,
			AuthorID:   1,
			GroupID:    1,
		}})
	secnote.Save()

	secURL := fmt.Sprintf("%s/nocsrf/onetimesec/display-%s", base_url, note_title)
	resp := struct {
		Link   string `json:"link"`
		Secret string `json:"password"`
	}{
		Link:   secURL,
		Secret: secret,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] GenerateOnetimeSecURL NewEncoder %s\n", err.Error())
	}
}

func GetOnetimeSecret(w http.ResponseWriter, r *http.Request) {
	note_id_str := r.PathValue("secret_id")

	if note_id_str == "-1" {
		http.Error(w, "ERROR got -1 in id", http.StatusBadRequest)
		return
	}

	if strings.HasPrefix(note_id_str, "display-") {
		note_id_str = strings.TrimPrefix(note_id_str, "display-")
		base_url := m.GetConfig("base_url", "")
		html := fmt.Sprintf(`<html><body align="center"><h2><a href="%s/nocsrf/onetimesec/%s">Click to get the secret:</a></h2></body></html>`, base_url, note_id_str)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
		return
	}

	note_sec := m.GetNote(note_id_str)
	if note_sec == nil {
		http.Error(w, `{"error":"ERROR"}`, http.StatusNotFound)
		return
	}
	sec := note_sec.Content
	note_sec.Delete()

	// Encode to base64 so arbitrary bytes (including <, >, &, quotes, unicode)
	// round-trip to the browser intact and are never parsed as markup.
	secretB64 := base64.StdEncoding.EncodeToString([]byte(sec))

	tmpl := template.Must(template.New("onetime").Parse(onetimeDisplayTemplate))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, map[string]any{"SecretB64": secretB64}); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] GetOnetimeSecret template execute %s\n", err.Error())
	}
}
