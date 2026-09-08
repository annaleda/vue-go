// quote-service calcola il preventivo.
//
// È il servizio più interessante dei due perché non è isolato: per lavorare
// deve chiamare catalog-service. Serve a vedere:
//   - come si legge un corpo JSON da una richiesta POST
//   - come si chiama un altro servizio (vedi catalog_client.go)
//   - come si distingue un errore dell'utente (400) da un guasto (502)
//   - come si tiene la logica di business separata dal trasporto HTTP
package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// server tiene le dipendenze del servizio.
//
// Metterle in una struct invece che in variabili globali è ciò che permette
// di sostituirle: in un test si passa un client finto e non serve la rete.
type server struct {
	catalog *CatalogClient
}

func main() {
	port := getenv("PORT", "8080")
	catalogURL := getenv("CATALOG_URL", "http://localhost:8081")
	log.Printf("catalogo configurato su %s", catalogURL)

	s := &server{
		catalog: NewCatalogClient(catalogURL, 60*time.Second),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/preventivi", s.handlePreventivo)
	mux.HandleFunc("/health/liveness", handleHealth)
	mux.HandleFunc("/health/readiness", s.handleReadiness)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      cors(logging(mux)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("quote-service in ascolto sulla porta %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("errore del server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("spegnimento in corso...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("spegnimento forzato: %v", err)
	}
	log.Println("quote-service terminato")
}

// handlePreventivo è il cuore del servizio.
func (s *server) handlePreventivo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "usare POST")
		return
	}

	// Limite sul corpo: senza, un client può inviare gigabyte e far esaurire
	// la memoria del processo.
	body := http.MaxBytesReader(w, r.Body, 64*1024)
	defer r.Body.Close()

	var req RichiestaPreventivo
	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields() // campi non previsti = errore, non silenzio
	if err := dec.Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "corpo della richiesta mancante")
			return
		}
		writeError(w, http.StatusBadRequest, "JSON non valido: "+err.Error())
		return
	}

	// Il context della richiesta viene propagato al client HTTP: se il
	// browser chiude la connessione, anche la chiamata al catalogo si ferma.
	garanzie, err := s.catalog.Garanzie(r.Context())
	if err != nil {
		log.Printf("catalogo non raggiungibile: %v", err)
		// 502 e non 500: il guasto non è qui, è a valle. La distinzione
		// conta per chi guarda i grafici e deve capire chi è rotto.
		writeError(w, http.StatusBadGateway, "catalogo garanzie non disponibile")
		return
	}

	prev, err := Calcola(req, garanzie)
	if err != nil {
		if errors.Is(err, ErrValidazione) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("errore nel calcolo: %v", err)
		writeError(w, http.StatusInternalServerError, "errore nel calcolo del preventivo")
		return
	}

	writeJSON(w, http.StatusOK, prev)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "UP"})
}

// handleReadiness dice se il servizio è pronto a ricevere traffico.
//
// La differenza con liveness è importante in Kubernetes:
//   - liveness fallita  -> il pod viene RIAVVIATO
//   - readiness fallita -> il pod resta vivo ma non riceve traffico
//
// Qui la readiness dipende dal catalogo: se non è raggiungibile il servizio
// non può fare il suo lavoro, ma riavviarlo non servirebbe a nulla.
func (s *server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if _, err := s.catalog.Garanzie(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "DOWN",
			"causa":  "catalog-service non raggiungibile",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "UP"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("errore serializzando la risposta: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"errore": message})
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}

// cors permette al frontend di chiamare questo servizio da un'altra origine.
//
// Serve solo in sviluppo, quando Vue gira sulla porta 5173 e il servizio
// sull'8080: il browser considera due porte diverse due origini diverse.
// In produzione frontend e API stanno dietro lo stesso gateway, quindi
// l'origine è la stessa e questo middleware non farebbe nulla.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", getenv("CORS_ORIGIN", "*"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
