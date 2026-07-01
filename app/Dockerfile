# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /notiflex-api main.go

# Runtime Stage
FROM scratch

COPY --from=builder /notiflex-api /notiflex-api

EXPOSE 8080

ENTRYPOINT ["/notiflex-api"]
