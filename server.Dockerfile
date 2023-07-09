FROM golang:1.17.8 AS builder

WORKDIR /build

COPY . .

RUN go mod download

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server server.go

# multistage build to copy only binary and config
FROM scratch

COPY --from=builder /build/server /
COPY --from=builder /build/config.yaml /config.yaml

EXPOSE 3333

ENTRYPOINT ["/server"]
