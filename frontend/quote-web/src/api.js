// Tutte le chiamate HTTP passano da qui.
//
// Tenere la rete in un modulo separato dai componenti è una buona abitudine:
// i componenti restano concentrati sulla presentazione, e il giorno che
// cambia un endpoint si tocca un file solo.
//
// Uso fetch, che è nativo del browser. Nei progetti Poste si usa axios, che
// aggiunge interceptor, timeout e una gestione degli errori più comoda:
// concettualmente fa la stessa cosa.

// In sviluppo il proxy di Vite gira /api verso i due servizi Go.
// In produzione il gateway fa la stessa cosa, quindi il percorso non cambia.
const BASE = import.meta.env.VITE_API_BASE || '/api'

async function leggiRisposta(res) {
  // Una risposta HTTP arrivata non significa che sia andata bene:
  // fetch non lancia eccezioni sui 4xx/5xx, va controllato a mano.
  // È la differenza più insidiosa rispetto ad axios per chi arriva da lì.
  if (!res.ok) {
    let messaggio = `Errore ${res.status}`
    try {
      const corpo = await res.json()
      if (corpo && corpo.errore) messaggio = corpo.errore
    } catch (e) {
      // il corpo non era JSON: teniamo il messaggio generico
    }
    throw new Error(messaggio)
  }
  return res.json()
}

export async function caricaGaranzie() {
  const res = await fetch(`${BASE}/garanzie`, {
    headers: { Accept: 'application/json' },
  })
  return leggiRisposta(res)
}

export async function calcolaPreventivo(richiesta) {
  const res = await fetch(`${BASE}/preventivi`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
    body: JSON.stringify(richiesta),
  })
  return leggiRisposta(res)
}
