# Uno step Hypernova non è un sito statico: è un SERVER Node che renderizza
# componenti su richiesta. Quindi, a differenza di quote-web, l'immagine
# finale non è nginx ma Node.
#
# Node 16 e non 18+: webpack 4 usa internamente MD4, che le versioni recenti
# di OpenSSL rifiutano. È lo stesso vincolo che si incontra su qualunque
# progetto rimasto su webpack 4 — come i micro-frontend Poste.

FROM node:16-alpine AS build

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci --no-audit --no-fund 2>/dev/null || npm install --no-audit --no-fund

COPY . .
RUN npm run build


FROM node:16-alpine

WORKDIR /app

# Solo le dipendenze di produzione: niente webpack, niente babel, niente
# vue-loader. Sono servite a costruire, non a servire.
COPY package.json package-lock.json* ./
RUN npm ci --omit=dev --no-audit --no-fund 2>/dev/null \
    || npm install --production --no-audit --no-fund

COPY --from=build /app/dist ./dist

USER node

EXPOSE 3032
CMD ["node", "dist/server.js"]
