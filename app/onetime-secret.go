package app

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	u "github.com/sunshine69/golang-tools/utils"
	m "github.com/sunshine69/webnote-go/models"
)

func GenerateOnetimeSecURL(w http.ResponseWriter, r *http.Request) {
	submit_type := m.GetRequestValue(r, "submit", "")
	base_url := m.GetRequestValue(r, "base_url", "")
	var secret string
	if submit_type == "submit_genpass" {
		length_str := m.GetRequestValue(r, "password_len", "12")
		password_len, err := strconv.Atoi(length_str)
		if u.CheckErrNonFatal(err, "GenerateOnetimeSecURL") != nil {
			fmt.Fprintf(w, "ERROR length should be a integer")
			return
		}
		secret = u.GenRandomString(password_len)
	} else {
		secret = m.GetRequestValue(r, "sec_content", "")
		fmt.Printf("DEBUG sec is %s\n", secret)
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
	secnote := m.NoteNew(map[string]interface{}{
		"content":    secret,
		"title":      note_title,
		"permission": int8(0),
	})
	secnote.Save()

	secURL := fmt.Sprintf("%s/nocsrf/onetimesec/display-%s", base_url, note_title)
	fmt.Fprintf(w, `{"link": "%s", "password": "%s"}`, secURL, secret)
}

func GetOnetimeSecret(w http.ResponseWriter, r *http.Request) {
	note_id_str := r.PathValue("secret_id")

	if note_id_str == "-1" {
		fmt.Fprintf(w, "ERROR got -1 in id")
		return
	}

	if strings.HasPrefix(note_id_str, "display-") {
		note_id_str = strings.TrimPrefix(note_id_str, "display-")
		base_url := m.GetConfig("base_url", "")
		html := fmt.Sprintf(`<html><body align="center"><h2><a href="%s/nocsrf/onetimesec/%s">Click to get the secret:</a></h2></body></html>`, base_url, note_id_str)
		w.Write([]byte(html))
		return
	}

	note_sec := m.GetNote(note_id_str)
	if note_sec == nil {
		fmt.Fprintf(w, "ERROR")
		return
	}
	sec := note_sec.Content
	note_sec.Delete()
	html := u.GoTemplateString(`
	<html>
<head>
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
            height: 100vh;
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
        <div id="secretDisplay">{{ .secret }}</div>
        <button onclick="copyToClipboard()">Copy to Clipboard</button>
    </div>

    <script>
        function copyToClipboard() {
            const text = document.getElementById("secretDisplay").textContent;
            navigator.clipboard.writeText(text).then(() => {
                // Remove any existing message
                const existingMessage = document.querySelector('.copy-message');
                if (existingMessage) {
                    existingMessage.remove();
                }

                const messageDiv = document.createElement('div');
                messageDiv.textContent = 'Content copied to clipboard!';
                messageDiv.className = 'copy-message';
                document.querySelector('.container').appendChild(messageDiv);

                // Remove message after 3 seconds
                setTimeout(() => {
                    if (messageDiv.parentNode) {
                        messageDiv.remove();
                    }
                }, 3000);
            }).catch(err => {
                console.error('Failed to copy: ', err);

                // Show error message
                const errorDiv = document.createElement('div');
                errorDiv.textContent = 'Failed to copy to clipboard';
                errorDiv.style.marginTop = '10px';
                errorDiv.style.color = '#dc3545';
                document.querySelector('.container').appendChild(errorDiv);

                setTimeout(() => {
                    if (errorDiv.parentNode) {
                        errorDiv.remove();
                    }
                }, 3000);
            });
        }
    </script>
</body>
</html>
	`, map[string]any{"secret": sec})
	fmt.Fprint(w, html)
}
