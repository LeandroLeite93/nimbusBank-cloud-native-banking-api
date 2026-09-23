package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Customer struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /customers", listCustomersHandler)

	const address = "127.0.0.1:8080"
	log.Printf("Iniciando NimbusBank em http://%s", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatalf("Erro ao executar o servidor: %v", err)
	}
}

func listCustomersHandler(w http.ResponseWriter, r *http.Request) {
	customers := []Customer{
		{ID: 1, Name: "Ana Silva", Email: "ana@example.com"},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(customers); err != nil {
		log.Printf("Erro ao escrever lista de clientes: %v", err)
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
