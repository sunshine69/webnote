package app

import (
	"fmt"
	"io"
	"net/http"

	"github.com/sunshine69/ollama-ui-go/lib"
)

func OllamaGetTags(w http.ResponseWriter, r *http.Request) {
	models, err := lib.GetOllamaModels()
	if err != nil {
		println("[DEBUG] [ERROR]: " + err.Error())
		http.Error(w, "Failed to call Ollama API", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(models)
}

func OllamaAsk(w http.ResponseWriter, r *http.Request) {
	var ollamaRequest lib.OllamaRequest
	jsonData, err := io.ReadAll(r.Body)
	if err != nil {
		// Handle error
	}
	fmt.Println(string(jsonData))
	if err := json.Unmarshal(jsonData, &ollamaRequest); err != nil {
		fmt.Printf("[DEBUG] Error: %s\n", err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	requestString, err := json.Marshal(ollamaRequest)
	if err != nil {
		http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
		return
	}
	response, err := lib.AskOllamaAPI(string(requestString))
	if err != nil {
		http.Error(w, "Failed to call Ollama API", http.StatusInternalServerError)
		return
	}
	fmt.Println("[DEBUG] AI response " + string(response))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		fmt.Println("[DEBUG] Error writing response: " + err.Error())
	}
}

func OllamaGetModel(w http.ResponseWriter, r *http.Request) {
	modelName := r.PathValue("model_name")
	model, err := lib.GetOllamaModel(modelName)
	if err != nil {
		http.Error(w, "Failed to call Ollama API", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(model)
}
