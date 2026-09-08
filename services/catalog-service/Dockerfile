# Multi-stage build.
#
# Stadio 1: compila. Serve tutto il toolchain Go, ~800 MB.
# Stadio 2: esegue. Serve solo il binario, pochi MB.
#
# Il trucco è `COPY --from=build`: si prende solo il risultato e si butta
# via il resto. È lo stesso pattern usato dai micro-frontend in Poste, dove
# lo stadio di build ha Node e npm e quello finale no.

FROM golang:1.21-alpine AS build

WORKDIR /src

# Prima si copiano solo i file dei moduli e si scaricano le dipendenze.
# Questo layer viene messo in cache e non si rifà a ogni modifica del codice.
# Qui non ci sono dipendenze, ma la buona abitudine si prende dall'inizio.
COPY go.mod ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 produce un binario statico, senza dipendenze da libc.
# È ciò che permette di partire da un'immagine praticamente vuota.
# -ldflags="-s -w" toglie la tabella dei simboli: binario più piccolo.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/catalog-service .


# gcr.io/distroless/static contiene solo i certificati CA e il minimo
# indispensabile: niente shell, niente package manager, niente utenti extra.
# Meno superficie d'attacco, ma anche niente `docker exec ... sh` per
# guardarci dentro: si perde in comodità di debug, si guadagna in sicurezza.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=build /out/catalog-service /catalog-service

# nonroot è l'utente 65532 già definito nell'immagine distroless.
# Un container che gira da root è un rischio: se un attaccante evade
# l'isolamento, si ritrova root sull'host.
USER nonroot:nonroot

EXPOSE 8080
ENTRYPOINT ["/catalog-service"]
