FROM golang:1.27.1-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags='-s -w' -o /out/our-sell-api ./cmd/api

FROM alpine:3.22

RUN addgroup -S app && adduser -S -G app app \
    && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/our-sell-api /app/our-sell-api
USER app
EXPOSE 8080
ENTRYPOINT ["/app/our-sell-api"]
