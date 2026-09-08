FROM golang:1.21-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/quote-service .


FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=build /out/quote-service /quote-service

USER nonroot:nonroot

EXPOSE 8080
ENTRYPOINT ["/quote-service"]
