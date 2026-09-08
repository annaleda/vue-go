// catalog-service espone il catalogo delle garanzie assicurative.
//
// È il servizio più semplice dei due: non chiama nessuno, tiene i dati in
// memoria e li restituisce in JSON. Serve a prendere confidenza con:
//   - come si dichiara un tipo in Go (struct) e come lo si serializza in JSON
//   - come si scrive un server HTTP con la sola libreria standard
//   - come si gestiscono errori, timeout e spegnimento pulito
//
// Non ci sono dipendenze esterne: tutto quello che serve è nella standard
// library. È una caratteristica di Go che si nota subito arrivando da Java.
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Garanzia è una copertura acquistabile.
//
// I tag `json:"..."` dicono all'encoder come si chiamano i campi una volta
// serializzati. Senza tag, Go userebbe il nome del campo Go (Codice, Nome...):
// i tag servono a rispettare il contratto REST senza rinunciare alle
// convenzioni del linguaggio.
//
// Nota: i campi esportati (che iniziano con la maiuscola) sono gli unici
// visibili fuori dal package, e gli unici che l'encoder JSON può leggere.
type Garanzia struct {
	Codice      string  `json:"codice"`
	Nome        string  `json:"nome"`
	Descrizione string  `json:"descrizione"`
	PrezzoBase  float64 `json:"prezzoBase"`
	Obbligatoria bool   `json:"obbligatoria"`
}

// catalogo è il nostro "database": una slice in memoria.
//
// In un servizio vero qui ci sarebbe una query. Il punto di questo esercizio
// è il trasporto HTTP, non la persistenza, quindi i dati sono fissi.
var catalogo = []Garanzia{
	{
		Codice:       "RCA",
		Nome:         "Responsabilità Civile Auto",
		Descrizione:  "Copre i danni causati a terzi. Obbligatoria per legge.",
		PrezzoBase:   380.00,
		Obbligatoria: true,
	},
	{
		Codice:      "FURTO_INCENDIO",
		Nome:        "Furto e Incendio",
		Descrizione: "Copre il furto totale o parziale e i danni da incendio.",
		PrezzoBase:  120.00,
	},
	{
		Codice:      "KASKO",
		Nome:        "Kasko",
		Descrizione: "Copre i danni al proprio veicolo anche in caso di colpa.",
		PrezzoBase:  260.00,
	},
	{
		Codice:      "CRISTALLI",
		Nome:        "Cristalli",
		Descrizione: "Copre la rottura di parabrezza e finestrini.",
		PrezzoBase:  45.00,
	},
	{
		Codice:      "ASSISTENZA",
		Nome:        "Assistenza stradale",
		Descrizione: "Soccorso stradale e traino 24 ore su 24.",
		PrezzoBase:  35.00,
	},
	{
		Codice:      "TUTELA_LEGALE",
		Nome:        "Tutela legale",
		Descrizione: "Spese legali per controversie legate alla circolazione.",
		PrezzoBase:  30.00,
	},
}

func main() {
	port := getenv("PORT", "8080")

	// ServeMux è il router della standard library: associa un percorso a una
	// funzione. È volutamente minimale — niente path parameter, niente
	// middleware. Per progetti veri si usa un router esterno (chi, gorilla),
	// ma per capire come funziona HTTP in Go questo è il punto di partenza.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/garanzie", handleGaranzie)
	mux.HandleFunc("/health/liveness", handleHealth)
	mux.HandleFunc("/health/readiness", handleHealth)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: logging(mux),
		// Questi timeout non sono un dettaglio: senza, una connessione lenta
		// o malevola può tenere occupata una goroutine per sempre.
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Il server si avvia in una goroutine così il main può restare in ascolto
	// dei segnali di terminazione. `go` davanti a una chiamata la esegue in
	// modo concorrente: è la primitiva su cui è costruito tutto Go.
	go func() {
		log.Printf("catalog-service in ascolto sulla porta %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("errore del server: %v", err)
		}
	}()

	// Spegnimento pulito: Kubernetes manda SIGTERM prima di uccidere il pod.
	// Se lo ignoriamo, le richieste in corso vengono troncate. Qui invece
	// diamo 15 secondi per finire quelle già accettate.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("spegnimento in corso...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("spegnimento forzato: %v", err)
	}
	log.Println("catalog-service terminato")
}

// handleGaranzie restituisce l'elenco completo delle garanzie.
//
// Ogni handler ha sempre questa firma: riceve dove scrivere la risposta
// (ResponseWriter) e cosa è stato chiesto (*Request).
func handleGaranzie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo non consentito")
		return
	}
	writeJSON(w, http.StatusOK, catalogo)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "UP"})
}

// writeJSON serializza v e la scrive nella risposta.
//
// L'ordine conta: prima gli header, poi lo status, poi il corpo. Scrivere
// un header dopo WriteHeader non ha effetto, ed è un errore facile da fare.
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

// logging è un middleware: avvolge un handler e ne aggiunge un comportamento.
//
// In Go un middleware è semplicemente una funzione che prende un Handler e
// ne restituisce un altro. Non serve un framework.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}

// getenv legge una variabile d'ambiente con un valore di default.
//
// Tutta la configurazione passa da qui: è la regola dei container, dove
// l'immagine è immutabile e cambia solo l'ambiente in cui gira.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
