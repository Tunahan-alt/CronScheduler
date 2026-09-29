# ---- build stage ----
FROM golang:1.23-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ssl-monitor .

# ---- final stage ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /ssl-monitor /usr/local/bin/ssl-monitor

USER nobody
ENTRYPOINT ["/usr/local/bin/ssl-monitor"]