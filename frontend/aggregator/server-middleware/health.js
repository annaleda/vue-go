// Endpoint di health per le probe di Kubernetes.
//
// Nuxt non ne espone uno di suo: senza questo, la readinessProbe fallirebbe
// e il pod non riceverebbe mai traffico, pur essendo perfettamente sano.
//
// serverMiddleware monta un handler Connect PRIMA del rendering Nuxt, quindi
// la risposta non costa un render di pagina.
//
// Nota: risponde UP se il processo Nuxt e' vivo, non se gli step sono
// raggiungibili. E' voluto: se lo mettessimo in liveness, uno step giu'
// farebbe riavviare l'aggregator, che e' esattamente cio' che non si vuole.
// La pagina infatti gestisce gia' lo step mancante mostrando un avviso.
export default function (req, res) {
  res.setHeader('Content-Type', 'application/json')
  res.end(JSON.stringify({ status: 'UP' }))
}
