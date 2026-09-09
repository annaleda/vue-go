# L'aggregator è un'applicazione Nuxt in modalità server: fa il rendering
# lato server a ogni richiesta, quindi in produzione serve Node, non nginx.

FROM node:16-alpine AS build

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci --no-audit --no-fund 2>/dev/null || npm install --no-audit --no-fund

COPY . .

# router.base finisce DENTRO la build: Nuxt calcola gli URL degli asset in
# fase di compilazione, non a runtime. Passarlo solo come variabile
# d'ambiente al container non basterebbe — gli asset sarebbero gia' stati
# generati puntando alla radice.
ARG ROUTER_BASE=/
ENV ROUTER_BASE=$ROUTER_BASE
RUN npm run build


FROM node:16-alpine

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci --omit=dev --no-audit --no-fund 2>/dev/null \
    || npm install --production --no-audit --no-fund

# .nuxt contiene il risultato della build, il resto serve a nuxt start
COPY --from=build /app/.nuxt ./.nuxt
COPY --from=build /app/nuxt.config.js ./
COPY --from=build /app/static ./static
COPY --from=build /app/assets ./assets

# I serverMiddleware NON finiscono dentro .nuxt: Nuxt li carica da disco
# all'avvio, risolvendo il percorso "~/server-middleware/...". Senza questa
# riga il container parte e muore subito con
#   Error: Cannot resolve "~/server-middleware/health.js"
# I plugin invece sono impacchettati nella build e non vanno copiati.
COPY --from=build /app/server-middleware ./server-middleware

USER node

EXPOSE 8080
ENV HOST=0.0.0.0
CMD ["npm", "start"]
