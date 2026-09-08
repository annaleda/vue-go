{{/*
  I template in _helpers.tpl non producono file: definiscono frammenti
  riutilizzabili. Il trattino iniziale nel nome del file dice a Helm di
  non trattarlo come un manifesto.

  Si richiamano con `include "nome" argomento`. Siccome un template riceve
  un solo argomento, quando servono più valori si passa un dizionario
  costruito con `dict`.
*/}}

{{/*
  Etichette complete, da mettere su ogni risorsa.
  Le app.kubernetes.io/* sono le etichette standard raccomandate: sono
  quelle che gli strumenti (kubectl, dashboard, service mesh) si aspettano.
*/}}
{{- define "vue-go.labels" -}}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/instance: {{ .release }}
app.kubernetes.io/part-of: vue-go
app.kubernetes.io/managed-by: Helm
helm.sh/chart: {{ .chart }}
{{- end }}

{{/*
  Etichette di selezione: il sottoinsieme che lega un Deployment ai suoi Pod
  e un Service ai Pod da servire.

  Devono restare STABILI: il selector di un Deployment è immutabile dopo la
  creazione, quindi cambiarle qui rompe l'upgrade. Per questo sono separate
  dalle etichette complete, che invece possono cambiare liberamente (la
  versione del chart, per esempio, cambia a ogni release).
*/}}
{{- define "vue-go.selectorLabels" -}}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/part-of: vue-go
{{- end }}

{{/*
  Nome completo dell'immagine.
  Se repository è vuoto si usa il solo nome del servizio: è il caso in cui
  le immagini sono già sul nodo e non c'è nessun registry di mezzo.
*/}}
{{- define "vue-go.image" -}}
{{- if .repository -}}
{{ .repository }}/{{ .name }}:{{ .tag }}
{{- else -}}
{{ .name }}:{{ .tag }}
{{- end -}}
{{- end }}
