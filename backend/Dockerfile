FROM golang:1.25.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.22

RUN apk add --no-cache ca-certificates wget
RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /out/server /app/server
COPY --from=builder /out/migrate /app/migrate

USER app
EXPOSE 8080
ENTRYPOINT ["/app/server"]
