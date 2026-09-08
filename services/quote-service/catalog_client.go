package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Garanzia deve corrispondere a quella esposta da catalog-service.
//
// Qui si vede un problema reale delle architetture a microservizi: il modello
// è duplicato in due servizi. Le soluzioni sono generare il client dallo
// schema OpenAPI (è quello che fa il progetto Poste con swagger-codegen)
// oppure pubblicare una libreria condivisa. Per un esercizio, duplicare e
// saperlo è onesto; in produzione la duplicazione silenziosa fa danni.
type Garanzia struct {
	Codice       string  `json:"codice"`
	Nome         string  `json:"nome"`
	Descrizione  string  `json:"descrizione"`
	PrezzoBase   float64 `json:"prezzoBase"`
	Obbligatoria bool    `json:"obbligatoria"`
}

// CatalogClient parla con catalog-service.
//
// Tiene una cache perché il catalogo cambia molto di rado ma viene letto a
// ogni preventivo. Senza cache, ogni richiesta genererebbe una chiamata di
// rete inutile e legherebbe la disponibilità di quote-service a quella di
// catalog-service.
type CatalogClient struct {
	baseURL string
	http    *http.Client

	mu        sync.RWMutex
	cache     []Garanzia
	scadenza  time.Time
	durataTTL time.Duration
}

func NewCatalogClient(baseURL string, ttl time.Duration) *CatalogClient {
	return &CatalogClient{
		baseURL: baseURL,
		// Non si usa mai http.DefaultClient in produzione: non ha timeout,
		// quindi una chiamata può restare appesa per sempre.
		http:      &http.Client{Timeout: 5 * time.Second},
		durataTTL: ttl,
	}
}

// Garanzie restituisce il catalogo, dalla cache se ancora valida.
//
// Il context permette al chiamante di annullare la richiesta: se l'utente
// chiude la pagina, la chiamata a valle viene interrotta invece di occupare
// risorse per una risposta che nessuno leggerà.
func (c *CatalogClient) Garanzie(ctx context.Context) ([]Garanzia, error) {
	// Lettura con lock condiviso: più goroutine possono leggere insieme.
	c.mu.RLock()
	if c.cache != nil && time.Now().Before(c.scadenza) {
		cached := c.cache
		c.mu.RUnlock()
		return cached, nil
	}
	c.mu.RUnlock()

	url := c.baseURL + "/api/v1/garanzie"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("costruzione richiesta al catalogo: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chiamata al catalogo fallita: %w", err)
	}
	// defer esegue alla fine della funzione, qualunque sia l'uscita.
	// Dimenticare di chiudere il body è la causa più comune di perdita di
	// connessioni in Go.
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("il catalogo ha risposto %d", resp.StatusCode)
	}

	var garanzie []Garanzia
	if err := json.NewDecoder(resp.Body).Decode(&garanzie); err != nil {
		return nil, fmt.Errorf("risposta del catalogo non leggibile: %w", err)
	}

	// Scrittura con lock esclusivo.
	c.mu.Lock()
	c.cache = garanzie
	c.scadenza = time.Now().Add(c.durataTTL)
	c.mu.Unlock()

	return garanzie, nil
}
