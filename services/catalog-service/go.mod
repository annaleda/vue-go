// go.mod dichiara il modulo e le sue dipendenze.
//
// È l'equivalente del pom.xml, ma qui è quasi vuoto: questo servizio usa
// solo la standard library, quindi non c'è nessun `require`.
//
// Il nome del modulo è per convenzione il percorso da cui si scarica.
// Non deve esistere davvero per compilare in locale, ma se un giorno
// qualcuno importa questo pacchetto, è lì che Go andrà a cercarlo.
module github.com/annaleda/vue-go/services/catalog-service

go 1.21
