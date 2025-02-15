package app

import (
	"context"
	jsonstd "encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ollama/ollama/api"
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
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	fmt.Println(string(jsonData))
	if err := json.Unmarshal(jsonData, &ollamaRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	req := &api.ChatRequest{
		Model:    ollamaRequest.Model,
		Messages: ollamaRequest.Messages,
		Stream:   &ollamaRequest.Stream,
		Options:  ollamaRequest.Options,
		Format:   jsonstd.RawMessage(ollamaRequest.Format),
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	respFunc := func(resp api.ChatResponse) error {
		// fmt.Print(resp.Message.Content)
		fmt.Fprint(w, resp.Message.Content)
		flusher.Flush()
		return nil
	}

	err = client.Chat(ctx, req, respFunc)
	if err != nil {
		http.Error(w, "Failed to process chat request", http.StatusInternalServerError)
		return
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
