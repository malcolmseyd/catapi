FROM golang:1.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o catapi


FROM scratch

WORKDIR /app

COPY impact.ttf /app/
COPY img/ /app/img/

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/catapi /app/

EXPOSE 8080

CMD ["./catapi"]
