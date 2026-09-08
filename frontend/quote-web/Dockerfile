# Stadio 1: build.
#
# Serve Node e tutte le devDependencies (Vite, il plugin Vue, ~200 MB di
# node_modules) solo per produrre una manciata di file statici.
FROM node:18-alpine AS build

WORKDIR /app

# Prima package.json, poi npm ci, poi il resto del codice: così il layer
# delle dipendenze resta in cache finché non cambiano le dipendenze stesse.
# Invertire l'ordine significa reinstallare tutto a ogni modifica di un .vue.
COPY package.json package-lock.json* ./

# npm ci invece di npm install: installa esattamente le versioni del
# lockfile e fallisce se il lockfile non è coerente. In una build
# riproducibile è quello che si vuole.
# Il fallback a npm install serve solo se il lockfile non è stato committato.
RUN npm ci 2>/dev/null || npm install

COPY . .
RUN npm run build


# Stadio 2: esecuzione.
#
# Dell'intero stadio precedente si tiene solo /app/dist. Niente Node,
# niente node_modules, niente sorgenti: l'immagine finale è nginx più
# qualche centinaio di kilobyte di file statici.
FROM nginx:1.25-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

# L'immagine nginx ufficiale gira da root e ascolta sulla 80. Qui la si
# fa ascoltare sulla 8080 e girare come utente non privilegiato: sotto la
# porta 1024 servono privilegi di root, sopra no.
RUN chown -R nginx:nginx /usr/share/nginx/html /var/cache/nginx \
    && touch /var/run/nginx.pid \
    && chown nginx:nginx /var/run/nginx.pid

USER nginx

EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]
