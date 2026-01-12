# ---------- Builder ----------
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o tiny ./cmd/server


# ---------- Runtime ----------
FROM gcr.io/distroless/base-debian12

USER nonroot:nonroot
WORKDIR /app

COPY --from=builder /app/tiny /app/tiny

EXPOSE 8080

ENTRYPOINT ["/app/tiny"]
