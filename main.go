package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

type Customer struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Estado compartilhado, mantido somente durante a execução do programa.
var (
	customers = []Customer{
		{ID: 1, Name: "Ana Silva", Email: "ana@example.com"},
	}
	nextCustomerID = 2
	customersMu    sync.Mutex
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /customers", listCustomersHandler)
	mux.HandleFunc("POST /customers", createCustomerHandler)

	const address = "127.0.0.1:8080"
	log.Printf("Iniciando NimbusBank em http://%s", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatalf("Erro ao executar o servidor: %v", err)
	}
}

func listCustomersHandler(w http.ResponseWriter, r *http.Request) {
	customersMu.Lock()
	response := make([]Customer, len(customers))
	copy(response, customers)
	customersMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Erro ao escrever lista de clientes: %v", err)
	}
}

func createCustomerHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "Envie um objeto JSON com name e email", http.StatusBadRequest)
		return
	}
	// O corpo deve conter um único valor JSON.
	if err := decoder.Decode(new(any)); err != io.EOF {
		http.Error(w, "Envie apenas um objeto JSON", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	if input.Name == "" || input.Email == "" {
		http.Error(w, "name e email são obrigatórios", http.StatusBadRequest)
		return
	}

	customersMu.Lock()
	customer := Customer{ID: nextCustomerID, Name: input.Name, Email: input.Email}
	customers = append(customers, customer)
	nextCustomerID++
	customersMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(customer); err != nil {
		log.Printf("Erro ao escrever cliente criado: %v", err)
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
