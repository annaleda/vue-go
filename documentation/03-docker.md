# 03 — Docker

[← Vue](02-vue.md) · [Indice](../README.md#documentazione) · [Kubernetes e Helm →](04-kubernetes-e-helm.md)

---

## 1. Immagine e container

Sono i due concetti da non confondere.

| **Immagine** | **Container** |
|---|---|
| Pacchetto immutabile: filesystem + comando di avvio | Istanza in esecuzione di un'immagine |
| La classe | L'oggetto |
| Permanente, versionata | Effimero: quando muore, le sue modifiche spariscono |

L'immutabilità è il punto: la stessa identica immagine che gira sul portatile
è quella che gira in produzione. Non si "ricostruisce per l'ambiente di
produzione", si costruisce una volta e si promuove.

E **Docker non sostituisce il server**: un container gira sempre su una
macchina. Docker sostituisce la *procedura di installazione* su quella
macchina.

---

## 2. Multi-stage build: il concetto centrale

Guarda il Dockerfile di `catalog-service`:

```dockerfile
FROM golang:1.21-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/catalog-service .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/catalog-service /catalog-service
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/catalog-service"]
```

Due `FROM`, due stadi indipendenti. Solo l'ultimo produce l'immagine finale.

- **Stadio 1**: serve tutto il toolchain Go, circa 800 MB.
- **Stadio 2**: serve solo il binario compilato.

`COPY --from=build` prende il risultato dello stadio precedente e **butta via
tutto il resto**. Il compilatore, i sorgenti, la cache dei moduli: niente di
tutto questo finisce nell'immagine.

I numeri di questo progetto, misurati:

```
vue-go/catalog-service:0.1.0   6.85 MB
vue-go/quote-service:0.1.0     7.40 MB
vue-go/quote-web:0.1.0        48.50 MB
```

Senza multi-stage sarebbero circa 800 MB e 250 MB.

---

## 3. L'ordine delle istruzioni conta

```dockerfile
COPY go.mod ./
RUN go mod download      # ← layer messo in cache
COPY . .
RUN go build ...
```

Ogni istruzione crea un **layer**. Docker riusa i layer dalla cache finché
l'input non cambia.

Copiare prima solo i file delle dipendenze e poi il codice significa che
`go mod download` (o `npm ci`) si rifà **solo quando cambiano le dipendenze**,
non a ogni modifica di una riga di codice.

Invertire l'ordine — `COPY . .` prima di tutto — funziona lo stesso ma
reinstalla ogni volta. Su un progetto Node significa minuti invece di secondi.

Stessa cosa nel Dockerfile del frontend:

```dockerfile
COPY package.json package-lock.json* ./
RUN npm ci 2>/dev/null || npm install
COPY . .
RUN npm run build
```

> `npm ci` invece di `npm install`: installa esattamente le versioni del
> lockfile e fallisce se il lockfile non è coerente. In una build
> riproducibile è quello che si vuole.

---

## 4. Perché le immagini Go sono così piccole

Tre scelte che si sommano.

**`CGO_ENABLED=0`** produce un binario **statico**, senza dipendenze da libc.
È ciò che permette di partire da un'immagine praticamente vuota. Con CGO
attivo servirebbe almeno una distribuzione minimale con le librerie di sistema.

**`-ldflags="-s -w"`** rimuove la tabella dei simboli e le informazioni di
debug. Binario più piccolo, in cambio di stack trace meno leggibili.

**`distroless`** contiene solo i certificati CA e il minimo indispensabile:
niente shell, niente package manager, niente utenti extra. Meno superficie
d'attacco — se un attaccante riesce a eseguire codice, non trova nemmeno `sh`.

Il prezzo: **niente `docker exec ... sh` per guardarci dentro**. Si perde in
comodità di debug, si guadagna in sicurezza. È un compromesso consapevole, e
in questo progetto si vede anche altrove (vedi §6).

---

## 5. Non girare da root

```dockerfile
USER nonroot:nonroot
```

Un container che gira da root è un rischio concreto: se un attaccante evade
l'isolamento, si ritrova root sull'host.

Nel frontend è servito un passaggio in più, perché l'immagine nginx ufficiale
gira da root e ascolta sulla porta 80:

```dockerfile
RUN chown -R nginx:nginx /usr/share/nginx/html /var/cache/nginx \
    && touch /var/run/nginx.pid \
    && chown nginx:nginx /var/run/nginx.pid
USER nginx
EXPOSE 8080
```

Il cambio di porta non è cosmetico: **sotto la 1024 servono privilegi di
root**, sopra no. È il motivo per cui i container ben fatti ascoltano su 8080
e non su 80.

---

## 6. Il limite di distroless che abbiamo incontrato

Nel `docker-compose.yml` non c'è nessun `healthcheck` per i servizi Go, ed è
una scelta consapevole documentata nel file:

```yaml
# L'immagine distroless non ha shell né curl, quindi il classico
# test: ["CMD", "curl", "-f", "http://localhost:8080/health/liveness"]
# non può funzionare: dentro il container non esiste curl.
```

Le alternative sarebbero tre: usare un'immagine base più grassa, aggiungere al
binario un flag `--healthcheck` che chiama se stesso, oppure rinunciare.

Qui si rinuncia, perché in Kubernetes le probe le esegue il kubelet
**dall'esterno** e il problema non si pone. Vale la pena conoscere questo
limite invece di scoprirlo in produzione.

---

## 7. docker-compose

Compose descrive un insieme di container che lavorano insieme su una macchina.

```yaml
services:
  quote-service:
    build:
      context: ./services/quote-service
    environment:
      - CATALOG_URL=http://catalog-service:8080
    ports:
      - "8080:8080"
    depends_on:
      - catalog-service
```

Due cose da notare.

**`http://catalog-service:8080`** — il nome del servizio è risolto dal DNS
interno della rete di compose. Non servono indirizzi IP. È lo stesso
meccanismo che in Kubernetes fornisce il Service.

**`"8080:8080"`** — il primo numero è la porta sull'host, il secondo quella
nel container. Sono indipendenti: nel progetto le porte host sono
configurabili proprio perché sono una risorsa condivisa della macchina.

```bash
WEB_PORT=13000 QUOTE_PORT=18080 CATALOG_PORT=18081 docker compose up -d
```

> Attenzione: compose fa sostituzione delle variabili su **tutto il file,
> commenti compresi**. Scrivere un esempio con il dollaro e le graffe dentro
> un commento fa comparire un warning a ogni avvio. Ci siamo cascati.

### `depends_on` e la differenza con Kubernetes

`depends_on` controlla solo l'**ordine di avvio**, non la disponibilità: il
container parte, ma il servizio dentro può non essere ancora pronto.

E in Kubernetes **non esiste affatto**: i pod partono in ordine arbitrario e
possono riavviarsi in qualunque momento. È il motivo per cui `quote-service`
non fallisce all'avvio se il catalogo non risponde — si dichiara solo non
pronto finché non lo raggiunge.

Questa è la differenza concettuale più importante fra i due strumenti: in
compose puoi ancora ragionare per sequenze, in Kubernetes devi progettare
servizi che sopportano che le dipendenze non ci siano.

---

## 8. Comandi

```bash
docker compose up -d --build      # costruisce e avvia in background
docker compose ps                 # stato dei servizi
docker compose logs -f quote-service
docker compose down               # ferma e rimuove

docker images                     # elenca le immagini
docker build -t nome:tag .        # costruisce
docker run --rm -p 8080:8080 nome:tag

docker system df                  # quanto spazio sta occupando
docker system prune -a            # rimuove tutto ciò che non è in uso
```

L'ultimo serve più spesso di quanto si creda: Docker riempie il disco in fretta.

---

[← Vue](02-vue.md) · [Indice](../README.md#documentazione) · [Kubernetes e Helm →](04-kubernetes-e-helm.md)
