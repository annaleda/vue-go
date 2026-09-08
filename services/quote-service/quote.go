package main

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// RichiestaPreventivo è quello che arriva dal frontend.
//
// I puntatori non servono: in Go i valori zero (0, "", false) sono validi,
// e la validazione esplicita è preferibile ai campi opzionali.
type RichiestaPreventivo struct {
	EtaConducente  int      `json:"etaConducente"`
	Provincia      string   `json:"provincia"`
	ClasseDiMerito int      `json:"classeDiMerito"`
	AnniPatente    int      `json:"anniPatente"`
	Garanzie       []string `json:"garanzie"`
}

// VocePreventivo è una riga del preventivo: una garanzia con il suo prezzo
// finale, dopo l'applicazione dei coefficienti.
type VocePreventivo struct {
	Codice     string  `json:"codice"`
	Nome       string  `json:"nome"`
	PrezzoBase float64 `json:"prezzoBase"`
	Prezzo     float64 `json:"prezzo"`
}

// Preventivo è la risposta.
//
// Coefficienti è esposto di proposito: in un preventivatore vero il cliente
// ha diritto di capire perché paga quella cifra, e in fase di sviluppo
// aiuta a verificare che il calcolo faccia quello che deve.
type Preventivo struct {
	Voci         []VocePreventivo   `json:"voci"`
	Imponibile   float64            `json:"imponibile"`
	Imposte      float64            `json:"imposte"`
	Totale       float64            `json:"totale"`
	Coefficienti map[string]float64 `json:"coefficienti"`
}

// ErrValidazione segnala una richiesta non valida.
//
// errors.New crea un errore "sentinella": il chiamante può confrontarlo con
// errors.Is per distinguere un input sbagliato (400) da un guasto (500).
var ErrValidazione = errors.New("richiesta non valida")

// Valida controlla i dati in ingresso.
//
// Restituire un errore invece di lanciare un'eccezione è la differenza più
// visibile rispetto a Java: in Go gli errori sono valori di ritorno normali,
// e vanno controllati esplicitamente a ogni chiamata.
func (r RichiestaPreventivo) Valida() error {
	if r.EtaConducente < 18 || r.EtaConducente > 100 {
		return fmt.Errorf("%w: età conducente fuori intervallo (18-100)", ErrValidazione)
	}
	if r.ClasseDiMerito < 1 || r.ClasseDiMerito > 18 {
		return fmt.Errorf("%w: classe di merito fuori intervallo (1-18)", ErrValidazione)
	}
	if r.AnniPatente < 0 || r.AnniPatente > 80 {
		return fmt.Errorf("%w: anni di patente fuori intervallo", ErrValidazione)
	}
	if len(strings.TrimSpace(r.Provincia)) != 2 {
		return fmt.Errorf("%w: la provincia deve essere una sigla di 2 lettere", ErrValidazione)
	}
	if len(r.Garanzie) == 0 {
		return fmt.Errorf("%w: selezionare almeno una garanzia", ErrValidazione)
	}
	return nil
}

// coefficienteEta: i conducenti molto giovani costano di più.
func coefficienteEta(eta int) float64 {
	switch {
	case eta < 23:
		return 1.60
	case eta < 26:
		return 1.35
	case eta < 30:
		return 1.15
	case eta > 75:
		return 1.20
	default:
		return 1.00
	}
}

// coefficienteClasse traduce la classe di merito in un moltiplicatore.
//
// Nel sistema Bonus/Malus italiano si va dalla classe 1 (migliore) alla 18
// (peggiore). Si entra in 14, si scende di una classe ogni anno senza
// sinistri e si sale di due dopo un sinistro con responsabilità principale.
func coefficienteClasse(classe int) float64 {
	// Da 0.55 in classe 1 fino a 2.10 in classe 18, con passo costante.
	return 0.55 + (float64(classe-1) * (2.10 - 0.55) / 17.0)
}

// coefficienteProvincia: la sinistrosità cambia molto per area geografica.
//
// Le province non elencate usano il valore di default. È una mappa, che in
// Go si scrive con `map[chiave]valore`.
var coefficientiProvincia = map[string]float64{
	"NA": 1.45, "CE": 1.40, "RM": 1.20, "MI": 1.15,
	"TO": 1.10, "BO": 1.00, "FI": 1.00, "AO": 0.85, "PN": 0.85,
}

func coefficienteProvincia(sigla string) float64 {
	if c, ok := coefficientiProvincia[strings.ToUpper(sigla)]; ok {
		return c
	}
	return 1.00
}

// coefficientePatente: chi ha preso la patente da poco paga di più.
func coefficientePatente(anni int) float64 {
	if anni < 2 {
		return 1.25
	}
	if anni < 5 {
		return 1.10
	}
	return 1.00
}

// aliquotaImposte è l'imposta sulle assicurazioni RCA, semplificata.
const aliquotaImposte = 0.1550

// Calcola produce il preventivo a partire dalla richiesta e dal catalogo.
//
// Il catalogo arriva da un altro servizio: questa funzione non lo sa e non
// gli importa. Tenere la logica di calcolo separata dalle chiamate di rete
// la rende leggibile e verificabile senza avviare nulla.
func Calcola(req RichiestaPreventivo, catalogo []Garanzia) (Preventivo, error) {
	if err := req.Valida(); err != nil {
		return Preventivo{}, err
	}

	// Indicizzo il catalogo per codice: cercare in una mappa è immediato,
	// scorrere una slice per ogni garanzia scelta no.
	perCodice := make(map[string]Garanzia, len(catalogo))
	for _, g := range catalogo {
		perCodice[g.Codice] = g
	}

	coeff := map[string]float64{
		"eta":       coefficienteEta(req.EtaConducente),
		"classe":    coefficienteClasse(req.ClasseDiMerito),
		"provincia": coefficienteProvincia(req.Provincia),
		"patente":   coefficientePatente(req.AnniPatente),
	}
	moltiplicatore := coeff["eta"] * coeff["classe"] * coeff["provincia"] * coeff["patente"]

	prev := Preventivo{
		Voci:         []VocePreventivo{},
		Coefficienti: coeff,
	}

	for _, codice := range req.Garanzie {
		g, ok := perCodice[codice]
		if !ok {
			return Preventivo{}, fmt.Errorf("%w: garanzia sconosciuta %q", ErrValidazione, codice)
		}
		// I coefficienti di rischio si applicano solo alla RCA: le garanzie
		// accessorie hanno un prezzo che non dipende dalla storia del guidatore.
		prezzo := g.PrezzoBase
		if g.Codice == "RCA" {
			prezzo = g.PrezzoBase * moltiplicatore
		}
		prev.Voci = append(prev.Voci, VocePreventivo{
			Codice:     g.Codice,
			Nome:       g.Nome,
			PrezzoBase: arrotonda(g.PrezzoBase),
			Prezzo:     arrotonda(prezzo),
		})
		prev.Imponibile += prezzo
	}

	prev.Imponibile = arrotonda(prev.Imponibile)
	prev.Imposte = arrotonda(prev.Imponibile * aliquotaImposte)
	prev.Totale = arrotonda(prev.Imponibile + prev.Imposte)
	return prev, nil
}

// arrotonda a due decimali.
//
// I float non rappresentano esattamente i decimali: 0.1+0.2 non fa 0.3.
// Per un esercizio va bene arrotondare in uscita, ma in un sistema che
// muove denaro davvero si usano interi in centesimi o un tipo decimale.
func arrotonda(v float64) float64 {
	return math.Round(v*100) / 100
}
