# builder stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/backend ./cmd

# runtime stage
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/bin/backend /app/bin/backend
WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["/app/bin/backend"]