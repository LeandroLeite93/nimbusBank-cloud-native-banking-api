package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	const address = "127.0.0.1:8080"
	log.Printf("Iniciando NimbusBank em http://%s", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatalf("Erro ao executar o servidor: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Status string `json:"status"`
	}{Status: "ok"}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Erro ao escrever resposta de health: %v", err)
	}
}
