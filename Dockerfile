ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .

FROM debian:bookworm

COPY --from=builder /run-app /usr/local/bin/
RUN apt update && apt install -y ca-certificates sqlite3 && rm -rf /var/lib/apt/lists/*
RUN mkdir -p /data && chmod 777 /data
VOLUME [ "/data" ] 
CMD ["run-app"]
