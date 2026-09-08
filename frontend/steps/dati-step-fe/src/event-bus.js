import Vue from 'vue'

// L'EventBus condiviso fra micro-frontend.
//
// QUESTO È IL PUNTO PIÙ DELICATO DI TUTTA L'ARCHITETTURA, e vale la pena
// capirlo bene perché è controintuitivo.
//
// Ogni step è un bundle separato, compilato e servito da un servizio diverso.
// Se ogni bundle facesse semplicemente:
//
//     export const bus = new Vue()
//
// ...ognuno avrebbe la PROPRIA istanza. Verrebbero eseguiti nella stessa
// pagina, sembrerebbe tutto a posto, ma gli eventi emessi da uno non
// arriverebbero mai all'altro: sono due oggetti diversi che non si conoscono.
//
// È l'errore classico dei micro-frontend, e il sintomo è insidioso —
// funziona tutto tranne la comunicazione, senza nessun errore in console.
//
// La soluzione è avere UN SOLO oggetto per pagina. L'unico posto davvero
// condiviso fra bundle diversi è `window`, quindi il bus ci si registra
// sopra e chi arriva dopo riusa quello già presente.
//
// (Nel progetto Poste il bus arriva da `feu-ui-components-lib`, importata da
//  tutti gli step; qui si risolve con lo stesso principio ma senza dover
//  pubblicare una libreria npm condivisa.)

const CHIAVE = '__FEU_EVENT_BUS__'

function creaBus() {
  // Un'istanza vuota di Vue basta e avanza: ha già $emit, $on e $off.
  // È il modo idiomatico di fare un bus di eventi in Vue 2.
  return new Vue()
}

// Lato server (SSR) `window` non esiste: si restituisce un bus locale usa e
// getta. Non è un problema, perché durante il rendering server-side nessuno
// deve comunicare con nessuno — gli eventi servono solo dopo l'hydration,
// quando i componenti sono vivi nel browser.
function getBus() {
  if (typeof window === 'undefined') {
    return creaBus()
  }
  if (!window[CHIAVE]) {
    window[CHIAVE] = creaBus()
  }
  return window[CHIAVE]
}

export const bus = getBus()

// I nomi degli eventi come costanti, non come stringhe sparse nel codice.
//
// In un'architettura dove i pezzi si parlano solo per evento, questi nomi
// SONO il contratto fra i micro-frontend. Un errore di battitura non dà
// nessun errore: semplicemente l'evento non arriva. Tenerli in un posto solo
// è la difesa minima.
export const EVENTI = {
  DATI_CONDUCENTE: 'dati-conducente-cambiati',
  GARANZIE_SCELTE: 'garanzie-scelte',
  PREVENTIVO_CALCOLATO: 'preventivo-calcolato',
  STEP_COMPLETATO: 'step-completato',
}
