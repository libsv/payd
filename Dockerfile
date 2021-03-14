FROM golang:1.16.1-alpine as builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" ./cmd/bip270-server

FROM scratch

COPY --from=builder /app/bip270-server /usr/local/bin

EXPOSE 8442

CMD ["bip270-server"]
