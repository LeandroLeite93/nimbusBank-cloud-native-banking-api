package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// Os testes usam estado global e não devem chamar t.Parallel().
func resetCustomers(t *testing.T) {
	t.Helper()
	customersMu.Lock()
	originalCustomers, originalID := customers, nextCustomerID
	customers = []Customer{{ID: 1, Name: "Ana Silva", Email: "ana@example.com"}}
	nextCustomerID = 2
	customersMu.Unlock()
	t.Cleanup(func() {
		customersMu.Lock()
		customers, nextCustomerID = originalCustomers, originalID
		customersMu.Unlock()
	})
}

func TestCreateAndListCustomers(t *testing.T) {
	resetCustomers(t)
	req := httptest.NewRequest(http.MethodPost, "/customers", strings.NewReader(`{"name":" Bruno Lima ","email":" bruno@example.com "}`))
	w := httptest.NewRecorder()
	createCustomerHandler(w, req)
	if w.Code != http.StatusCreated || w.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("resposta inesperada: %d %v", w.Code, w.Header())
	}
	var created Customer
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created != (Customer{ID: 2, Name: "Bruno Lima", Email: "bruno@example.com"}) {
		t.Fatalf("cliente inesperado: %+v", created)
	}
	w = httptest.NewRecorder()
	listCustomersHandler(w, httptest.NewRequest(http.MethodGet, "/customers", nil))
	var listed []Customer
	if err := json.Unmarshal(w.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || len(listed) != 2 || listed[1] != created {
		t.Fatalf("listagem inesperada: %d %+v", w.Code, listed)
	}
}

func TestInvalidCustomerDoesNotChangeState(t *testing.T) {
	for _, body := range []string{
		``, `{`, `null`, `[]`, `{}`, `{"name":123,"email":"a@example.com"}`,
		`{"name":"   ","email":"a@example.com"}`, `{"name":"Ana","email":" "}`,
		`{"name":"Ana"}`, `{"email":"a@example.com"}`,
		`{"id":99,"name":"Ana","email":"a@example.com"}`,
		`{"name":"Ana","email":"a@example.com"} {}`,
		`{"name":"Ana","email":"a@example.com"} lixo`,
	} {
		t.Run(body, func(t *testing.T) {
			resetCustomers(t)
			w := httptest.NewRecorder()
			createCustomerHandler(w, httptest.NewRequest(http.MethodPost, "/customers", strings.NewReader(body)))
			if w.Code != http.StatusBadRequest {
				t.Fatalf("esperado 400, recebido %d", w.Code)
			}
			if len(customers) != 1 || nextCustomerID != 2 {
				t.Fatal("requisição inválida alterou o estado")
			}
		})
	}
}

func TestConcurrentCreateAndListCustomers(t *testing.T) {
	resetCustomers(t)
	const count = 20
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body := fmt.Sprintf(`{"name":"Cliente %d","email":"cliente%d@example.com"}`, i, i)
			w := httptest.NewRecorder()
			createCustomerHandler(w, httptest.NewRequest(http.MethodPost, "/customers", strings.NewReader(body)))
			if w.Code != http.StatusCreated {
				t.Errorf("esperado 201, recebido %d", w.Code)
			}
			w = httptest.NewRecorder()
			listCustomersHandler(w, httptest.NewRequest(http.MethodGet, "/customers", nil))
			var listed []Customer
			if err := json.Unmarshal(w.Body.Bytes(), &listed); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	if len(customers) != count+1 {
		t.Fatalf("cadastros perdidos: %d clientes", len(customers))
	}
	seen := make(map[int]bool)
	for _, customer := range customers {
		if seen[customer.ID] {
			t.Fatalf("ID repetido: %d", customer.ID)
		}
		seen[customer.ID] = true
	}
}
