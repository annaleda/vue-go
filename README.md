# vue-go

Preventivatore auto a microservizi: due servizi **Go**, un frontend **Vue 3**,
distribuiti su **Kubernetes** con **Helm** e Gateway API.

È un progetto didattico. Il dominio (calcolo di un premio RCA) è abbastanza
piccolo da stare in testa e abbastanza reale da non essere finto: coefficienti
per età, classe di merito, provincia e anzianità di patente.

---

## Cosa c'è dentro

```
vue-go/
├── services/
│   ├── catalog-service/     Go — espone il catalogo delle garanzie
│   └── quote-service/       Go — calcola il preventivo, chiama il catalogo
├── frontend/
│   └── quote-web/           Vue 3 + Vite, servito da nginx
├── infrastructure/
│   └── helm/vue-go/         chart Helm: Deployment, Service, Gateway, HTTPRoute
├── documentation/           spiegazioni, un file per argomento
└── docker-compose.yml       lo stesso stack senza Kubernetes
```

### Il flusso

```
browser
   │
   ▼
quote-web (nginx + Vue)
   │  /api/garanzie  ──────────────►  catalog-service
   │  /api/preventivi ─────────────►  quote-service
                                          │
                                          └──►  catalog-service
```

`quote-service` non ha il catalogo: se lo fa dare da `catalog-service` e lo
tiene in cache per un minuto. È la parte interessante — due servizi che si
parlano, con tutto quello che ne consegue: timeout, errori a valle, cache,
e una readiness che dipende da qualcun altro.

---

## Provarlo

### Con Docker Compose (il modo più veloce)

```bash
docker compose up -d --build
```

Poi apri **http://localhost:3000**

| Cosa | Indirizzo |
|---|---|
| Frontend | http://localhost:3000 |
| quote-service | http://localhost:8080/api/v1/preventivi |
| catalog-service | http://localhost:8081/api/v1/garanzie |

Se quelle porte sono occupate:

```bash
WEB_PORT=13000 QUOTE_PORT=18080 CATALOG_PORT=18081 docker compose up -d
```

Una prova rapida da riga di comando:

```bash
curl -s http://localhost:8081/api/v1/garanzie

curl -s -X POST http://localhost:8080/api/v1/preventivi \
  -H "Content-Type: application/json" \
  -d '{"etaConducente":40,"provincia":"BO","classeDiMerito":1,
       "anniPatente":20,"garanzie":["RCA","CRISTALLI"]}'
```

Prova a cambiare i parametri: un ventiduenne a Napoli in classe 16 con un anno
di patente paga circa **dieci volte** un quarantenne bolognese in classe 1.

Per fermare tutto:

```bash
docker compose down
```

### Su Kubernetes

Serve un cluster locale — kind, minikube o quello di Docker Desktop.

```bash
# 1. costruisci le immagini
docker compose build

# 2. rendile visibili al cluster
#    (con kind; con Docker Desktop questo passo non serve)
kind load docker-image vue-go/catalog-service:0.1.0
kind load docker-image vue-go/quote-service:0.1.0
kind load docker-image vue-go/quote-web:0.1.0

# 3. installa il chart
helm upgrade --install vue-go infrastructure/helm/vue-go \
  -f infrastructure/helm/vue-go/values-local.yaml

# 4. guarda cosa succede
kubectl -n vue-go get pods -w

# 5. raggiungi l'applicazione
kubectl -n vue-go port-forward svc/quote-web 8080:8080
```

Poi **http://localhost:8080**

Per disinstallare:

```bash
helm uninstall vue-go
```

`values-local.yaml` disattiva Gateway API, perché richiede che un controller
sia già installato nel cluster. Per provarla davvero:

```bash
helm install eg oci://docker.io/envoyproxy/gateway-helm \
  --version v1.0.1 -n envoy-gateway-system --create-namespace

helm upgrade --install vue-go infrastructure/helm/vue-go   # senza values-local
```

### Sviluppo, senza container

```bash
# terminale 1
cd services/catalog-service && PORT=8081 go run .

# terminale 2
cd services/quote-service && PORT=8080 CATALOG_URL=http://localhost:8081 go run .

# terminale 3
cd frontend/quote-web && npm install && npm run dev
```

Il frontend parte su http://localhost:5173 con il ricaricamento a caldo:
salvi un `.vue` e il browser si aggiorna senza perdere lo stato.

> Se `npm install` non trova i pacchetti, probabilmente npm punta a un
> registry aziendale. Si aggira così:
> `npm install --registry https://registry.npmjs.org`

I test dei servizi Go:

```bash
cd services/quote-service && go test -v ./...
```

---

## Documentazione

Le spiegazioni stanno in [`documentation/`](documentation/), un file per
argomento. Non ripetono il codice: raccontano il **perché** delle scelte, e
i punti dove è facile sbagliare.

| | |
|---|---|
| [01 — Go](documentation/01-go.md) | La sintassi che serve, il server HTTP, gli errori come valori, la concorrenza |
| [02 — Vue](documentation/02-vue.md) | Reattività, componenti, props ed eventi, le due sintassi a confronto |
| [03 — Docker](documentation/03-docker.md) | Multi-stage build, perché le immagini Go pesano 7 MB |
| [04 — Kubernetes e Helm](documentation/04-kubernetes-e-helm.md) | Deployment, Service, le tre probe, il chart, Gateway API |

---

## Note

- I due servizi Go non usano **nessuna dipendenza esterna**: solo la standard
  library. Non è una scelta di stile, è quanto basta davvero.
- Le immagini Go pesano **meno di 8 MB** perché partono da `distroless`, che
  non contiene nemmeno una shell.
- Non c'è un database: il catalogo è una slice in memoria. Il punto qui è la
  comunicazione tra servizi e il ciclo di vita in Kubernetes, non la persistenza.
- Il modello `Garanzia` è **duplicato** nei due servizi. È un problema vero
  delle architetture a microservizi, ed è commentato nel codice
  (`services/quote-service/catalog_client.go`) invece di essere nascosto.
