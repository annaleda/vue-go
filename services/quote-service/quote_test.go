package main

import (
	"errors"
	"testing"
)

// In Go i test stanno accanto al codice, nello stesso package, in file che
// finiscono per _test.go. Non serve nessuna libreria: `go test ./...` basta.
//
// La convenzione è: funzione che inizia per Test, riceve *testing.T, e
// segnala i fallimenti con t.Errorf (continua) o t.Fatalf (si ferma).

var catalogoDiProva = []Garanzia{
	{Codice: "RCA", Nome: "Responsabilità Civile Auto", PrezzoBase: 400, Obbligatoria: true},
	{Codice: "CRISTALLI", Nome: "Cristalli", PrezzoBase: 50},
}

func richiestaValida() RichiestaPreventivo {
	return RichiestaPreventivo{
		EtaConducente:  40,
		Provincia:      "BO",
		ClasseDiMerito: 1,
		AnniPatente:    20,
		Garanzie:       []string{"RCA"},
	}
}

func TestCalcolaConDatiValidi(t *testing.T) {
	prev, err := Calcola(richiestaValida(), catalogoDiProva)
	if err != nil {
		t.Fatalf("non mi aspettavo un errore: %v", err)
	}
	if len(prev.Voci) != 1 {
		t.Fatalf("mi aspettavo 1 voce, ne ho %d", len(prev.Voci))
	}
	if prev.Totale <= prev.Imponibile {
		t.Errorf("il totale (%v) deve essere maggiore dell'imponibile (%v)", prev.Totale, prev.Imponibile)
	}
}

// La classe di merito peggiore deve costare più della migliore.
// È il tipo di test che vale la pena scrivere: verifica una regola di
// business, non l'implementazione.
func TestClassePeggioreCostaDiPiu(t *testing.T) {
	migliore := richiestaValida()
	migliore.ClasseDiMerito = 1

	peggiore := richiestaValida()
	peggiore.ClasseDiMerito = 18

	pMigliore, err := Calcola(migliore, catalogoDiProva)
	if err != nil {
		t.Fatalf("errore inatteso: %v", err)
	}
	pPeggiore, err := Calcola(peggiore, catalogoDiProva)
	if err != nil {
		t.Fatalf("errore inatteso: %v", err)
	}
	if pPeggiore.Totale <= pMigliore.Totale {
		t.Errorf("classe 18 (%v) dovrebbe costare più di classe 1 (%v)",
			pPeggiore.Totale, pMigliore.Totale)
	}
}

// I coefficienti si applicano solo alla RCA: una garanzia accessoria deve
// costare uguale a chiunque.
func TestGaranziaAccessoriaNonDipendeDalRischio(t *testing.T) {
	giovane := richiestaValida()
	giovane.EtaConducente = 19
	giovane.Garanzie = []string{"CRISTALLI"}

	esperto := richiestaValida()
	esperto.Garanzie = []string{"CRISTALLI"}

	a, _ := Calcola(giovane, catalogoDiProva)
	b, _ := Calcola(esperto, catalogoDiProva)

	if a.Imponibile != b.Imponibile {
		t.Errorf("i cristalli dovrebbero costare uguale: %v vs %v", a.Imponibile, b.Imponibile)
	}
}

// I test tabellari sono molto usati in Go: un caso per riga, un solo corpo.
func TestValidazione(t *testing.T) {
	casi := []struct {
		nome     string
		modifica func(*RichiestaPreventivo)
	}{
		{"età troppo bassa", func(r *RichiestaPreventivo) { r.EtaConducente = 15 }},
		{"classe fuori scala", func(r *RichiestaPreventivo) { r.ClasseDiMerito = 25 }},
		{"provincia non valida", func(r *RichiestaPreventivo) { r.Provincia = "BOL" }},
		{"nessuna garanzia", func(r *RichiestaPreventivo) { r.Garanzie = nil }},
	}

	for _, c := range casi {
		// t.Run crea un sottotest con un nome proprio: nell'output si vede
		// esattamente quale caso è fallito.
		t.Run(c.nome, func(t *testing.T) {
			req := richiestaValida()
			c.modifica(&req)
			_, err := Calcola(req, catalogoDiProva)
			if !errors.Is(err, ErrValidazione) {
				t.Errorf("mi aspettavo ErrValidazione, ho avuto: %v", err)
			}
		})
	}
}

func TestGaranziaSconosciuta(t *testing.T) {
	req := richiestaValida()
	req.Garanzie = []string{"NON_ESISTE"}

	if _, err := Calcola(req, catalogoDiProva); !errors.Is(err, ErrValidazione) {
		t.Errorf("mi aspettavo ErrValidazione, ho avuto: %v", err)
	}
}
