FROM golang:1.17.1-buster as builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o payctl -ldflags="-s -w" ./cmd/cli

FROM scratch

COPY --from=builder /app/payctl /bin/payctl

CMD ["/bin/payctl"]
