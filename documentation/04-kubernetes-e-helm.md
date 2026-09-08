# 04 — Kubernetes e Helm

[← Docker](03-docker.md) · [Indice](../README.md#documentazione)

---

## 1. Cosa aggiunge Kubernetes

Docker fa girare *un* container su *una* macchina. Non decide su quale server
metterlo, non lo riavvia se muore, non bilancia il carico fra dieci copie, non
gestisce aggiornamenti progressivi.

Kubernetes fa esattamente quello. Il modo in cui lo fa è **dichiarativo**: non
gli dici *"avvia due container"*, gli dici *"voglio due repliche"* e lui si
occupa di mantenere quella condizione. Se un pod muore, ne crea un altro. Se
un nodo cade, li sposta.

| Docker Compose | Kubernetes |
|---|---|
| `service` | `Deployment` + `Service` |
| nome del servizio | nome DNS del `Service` |
| `environment` | `ConfigMap` / `Secret` |
| `ports` | `Gateway` (o `port-forward`) |
| `healthcheck` | probe di liveness / readiness / startup |
| `depends_on` | **nessun equivalente** |

---

## 2. Gli oggetti che servono

### Pod

L'unità minima: uno o più container che condividono rete e volumi. **Non si
crea mai un Pod a mano**: lo crea un Deployment.

### Deployment

Dichiara quanti pod vuoi e con quale immagine, e gestisce gli aggiornamenti
sostituendoli progressivamente.

```yaml
spec:
  replicas: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: quote-service
```

**Il `selector` è immutabile dopo la creazione.** Cambiarlo rompe l'upgrade
del chart. È il motivo per cui in `_helpers.tpl` le etichette di selezione
sono separate da quelle complete: le prime devono restare stabili, le seconde
possono cambiare a ogni release.

### Service

Due cose insieme:

- un **nome DNS stabile** (`catalog-service`) che non cambia mai, anche se i
  pod dietro nascono e muoiono con IP sempre diversi;
- un **bilanciatore** che distribuisce le richieste sui pod pronti.

"Pronti" è la parola chiave: un pod la cui readiness fallisce viene tolto
dagli endpoint e smette di ricevere traffico, pur restando vivo.

È il motivo per cui `quote-service` può chiamare `http://catalog-service:8080`
senza sapere nulla di indirizzi IP.

### ConfigMap

La configurazione, separata dall'immagine.

```yaml
data:
  PORT: "8080"
  CATALOG_URL: "http://catalog-service:8080"
```

**Modificare una ConfigMap non riavvia i pod da sola.** È la trappola più
comune: cambi un valore, fai `helm upgrade`, e i pod continuano con i valori
vecchi. La soluzione è nel `deployment.yaml`:

```yaml
annotations:
  checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") $ | sha256sum }}
```

Se il contenuto della ConfigMap cambia, cambia il checksum, cambia il template
del pod, e Kubernetes fa ripartire i pod. È il trucco più utile del chart.

---

## 3. Le tre probe

Questa è la parte che va capita bene, perché configurarla male è una delle
cause più comuni di guai in produzione.

| Probe | Domanda | Se fallisce |
|---|---|---|
| **startup** | È ancora in fase di avvio? | Le altre due restano sospese |
| **readiness** | Posso mandargli traffico? | Il pod è tolto dal Service, **non riavviato** |
| **liveness** | È ancora sano? | Il pod viene **ucciso** e ricreato |

### La regola d'oro

**La liveness non deve mai controllare le dipendenze esterne.**

Se controllasse il database, un database lento farebbe riavviare in massa
tutti i pod — peggiorando esattamente la situazione che si voleva salvare, e
aggiungendo una tempesta di riavvii a un sistema già in difficoltà.

Le dipendenze vanno nella **readiness**: il pod resta vivo, smette solo di
ricevere traffico, e torna disponibile da solo quando la dipendenza si
riprende.

Nel progetto è implementato esattamente così, in `quote-service/main.go`:

```go
// liveness: risponde sempre, dice solo che il processo è vivo
func handleHealth(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]string{"status": "UP"})
}

// readiness: verifica il catalogo
func (s *server) handleReadiness(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
    defer cancel()
    if _, err := s.catalog.Garanzie(ctx); err != nil {
        writeJSON(w, http.StatusServiceUnavailable, map[string]string{
            "status": "DOWN", "causa": "catalog-service non raggiungibile",
        })
        return
    }
    writeJSON(w, http.StatusOK, map[string]string{"status": "UP"})
}
```

### La startup probe

Serve alle applicazioni lente a partire. Finché fallisce, le altre due sono
sospese:

```yaml
startupProbe:
  httpGet: { path: /health/liveness, port: http }
  periodSeconds: 2
  failureThreshold: 15      # concede fino a 30 secondi
```

Senza, un'applicazione che impiega un minuto ad avviarsi verrebbe uccisa dalla
liveness prima di finire, in un ciclo infinito. È il problema classico delle
applicazioni Java con deploy di EAR: minuti di avvio.

Qui i servizi Go partono in millisecondi e la startup probe è quasi
superflua — c'è per completezza.

---

## 4. requests e limits

```yaml
resources:
  requests: { cpu: 50m, memory: 64Mi }
  limits:   { cpu: 300m, memory: 128Mi }
```

- **requests** = quanto lo scheduler *riserva* per decidere su quale nodo
  mettere il pod
- **limits** = il tetto oltre il quale interviene

I due comportamenti al superamento sono **molto diversi**:

- superare il limite di **CPU** rallenta il container (*throttling*): è
  sgradevole ma sopravvive;
- superare quello di **memoria** lo fa uccidere di colpo (**OOMKilled**).

La memoria non si comprime, e questa asimmetria va tenuta a mente quando si
dimensionano i limiti.

I valori bassi di questo progetto (50m CPU, 64Mi memoria) sono realistici per
un binario Go statico. Un servizio Java equivalente partirebbe da 256-512Mi
solo per la JVM: è uno dei motivi per cui Go è così usato nei servizi di
infrastruttura.

---

## 5. Gateway API

È il successore di Ingress, e la differenza principale è che **separa le
responsabilità** in risorse distinte:

| Risorsa | Chi la gestisce |
|---|---|
| `GatewayClass` | la piattaforma — chi implementa il gateway (Envoy, NGINX, Istio) |
| `Gateway` | chi amministra il cluster — porte, protocolli, certificati |
| `HTTPRoute` | chi sviluppa l'applicazione — le regole di instradamento |

Con Ingress tutto stava in un'unica risorsa piena di annotazioni proprietarie,
diverse per ogni controller. Qui il modello è standard e i ruoli sono netti.

```yaml
rules:
  - matches:
      - path: { type: PathPrefix, value: /api/garanzie }
    filters:
      - type: URLRewrite
        urlRewrite:
          path:
            type: ReplacePrefixMatch
            replacePrefixMatch: /api/v1/garanzie
    backendRefs:
      - name: catalog-service
        port: 8080
```

Due cose importanti.

**L'ordine delle regole conta.** Il match su prefisso è ingordo: `/` catturerebbe
anche `/api/garanzie`. Le regole specifiche vanno prima, `/` per ultimo.

**Il filtro URLRewrite** è il motivo principale per cui si mette un gateway
davanti ai servizi: il percorso pubblico (`/api/garanzie`) resta stabile per
chi lo consuma, mentre quello interno (`/api/v1/garanzie`) può cambiare
versione senza rompere nessuno.

> Il **controller** non fa parte del chart: è infrastruttura di cluster e si
> installa a parte. Per questo `values-local.yaml` disattiva il gateway —
> senza controller, le risorse verrebbero create ma nessuno le realizzerebbe:
> si vedrebbero con `kubectl` e non funzionerebbero. È il modo peggiore di
> fallire, perché sembra tutto a posto.

---

## 6. Helm

Helm è il gestore di pacchetti di Kubernetes: prende dei template, ci
sostituisce dei valori e produce i manifesti finali.

```
infrastructure/helm/vue-go/
├── Chart.yaml           metadati: nome, versione del chart, versione dell'app
├── values.yaml          i valori di default
├── values-local.yaml    override per il cluster locale
└── templates/
    ├── _helpers.tpl     frammenti riutilizzabili (non produce file)
    ├── namespace.yaml
    ├── configmap.yaml
    ├── deployment.yaml
    ├── service.yaml
    ├── gateway.yaml
    └── httproute.yaml
```

### Un template per N servizi

Il chart non ha un file per servizio. Ha un file che **cicla** sui servizi:

```yaml
{{- range $name, $svc := .Values.services }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ $name }}
...
{{- end }}
```

Aggiungere un servizio significa aggiungere una voce in `values.yaml`, non un
file in `templates/`.

Dentro un `range`, `.` cambia significato: punta all'elemento corrente. Per
accedere ai valori globali serve `$` — da cui `$.Values.namespace.name`. È
l'errore più comune scrivendo template Helm.

### Gli helper

I file che iniziano con `_` non producono manifesti: definiscono frammenti.

```
{{- define "vue-go.labels" -}}
app.kubernetes.io/name: {{ .name }}
...
{{- end }}
```

Si richiamano con `include`, e siccome un template riceve **un solo
argomento**, quando ne servono più d'uno si costruisce un dizionario:

```
{{- include "vue-go.labels" (dict "name" $name "release" $.Release.Name) | nindent 4 }}
```

`nindent 4` indenta di 4 spazi e va a capo prima: in YAML l'indentazione è
sintassi, e sbagliarla rompe il manifesto.

### Vedere cosa produce, prima di applicarlo

```bash
helm lint infrastructure/helm/vue-go
helm template vue-go infrastructure/helm/vue-go
helm template vue-go infrastructure/helm/vue-go -f infrastructure/helm/vue-go/values-local.yaml
```

`helm template` renderizza senza toccare il cluster. È il modo giusto di
lavorare su un chart: si guarda l'output, non si installa alla cieca.

Con i valori di default questo chart produce 13 risorse; con `values-local`
ne produce 10, perché il gateway è disattivato.

---

## 7. Provarlo

```bash
# 1. costruisci le immagini
docker compose build

# 2. rendile visibili al cluster (con kind)
kind load docker-image vue-go/catalog-service:0.1.0
kind load docker-image vue-go/quote-service:0.1.0
kind load docker-image vue-go/quote-web:0.1.0

# 3. installa
helm upgrade --install vue-go infrastructure/helm/vue-go \
  -f infrastructure/helm/vue-go/values-local.yaml

# 4. guarda
kubectl -n vue-go get pods -w

# 5. raggiungi
kubectl -n vue-go port-forward svc/quote-web 8080:8080
```

> `imagePullPolicy: IfNotPresent` è obbligatorio quando le immagini sono
> caricate sul nodo. Con `Always` Kubernetes proverebbe a scaricarle da un
> registry e fallirebbe con `ImagePullBackOff`.

### Quando qualcosa non va

```bash
kubectl -n vue-go get pods                  # stato generale
kubectl -n vue-go describe pod <nome>       # eventi: quasi sempre la risposta è qui
kubectl -n vue-go logs <nome>               # log dell'applicazione
kubectl -n vue-go logs <nome> --previous    # log del container morto prima
kubectl -n vue-go get events --sort-by=.lastTimestamp
```

`describe` è il primo comando da usare, sempre: la sezione `Events` in fondo
dice perché un pod non parte molto più chiaramente dei log.

Gli stati che si incontrano più spesso:

| Stato | Significato |
|---|---|
| `ImagePullBackOff` | immagine non trovata — controlla nome, tag e `pullPolicy` |
| `CrashLoopBackOff` | il container parte e muore in ciclo — guarda i log |
| `Pending` | nessun nodo ha risorse sufficienti — controlla le `requests` |
| `0/1 Running` | in esecuzione ma **non pronto**: la readiness fallisce |

L'ultimo è quello interessante di questo progetto: se `catalog-service` non
è raggiungibile, `quote-service` resta `Running` ma `0/1`, e non riceve
traffico. È esattamente il comportamento voluto.

---

## 8. Comandi utili

```bash
helm lint <chart>                     # controlla il chart
helm template <nome> <chart>          # renderizza senza installare
helm upgrade --install <nome> <chart> # installa o aggiorna
helm list -A                          # release installate
helm history <nome>                   # storico delle revisioni
helm rollback <nome> <revisione>      # torna indietro
helm uninstall <nome>                 # rimuove
```

`helm rollback` è la funzione che si apprezza di più il giorno che serve:
Helm conserva le revisioni precedenti e ci torna con un comando.

---

[← Docker](03-docker.md) · [Indice](../README.md#documentazione)
