# L'aggregator è un'applicazione Nuxt in modalità server: fa il rendering
# lato server a ogni richiesta, quindi in produzione serve Node, non nginx.

FROM node:16-alpine AS build

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci --no-audit --no-fund 2>/dev/null || npm install --no-audit --no-fund

COPY . .
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

USER node

EXPOSE 8080
ENV HOST=0.0.0.0
CMD ["npm", "start"]
