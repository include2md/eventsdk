FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/adapter ./sdk/demo/adapter

FROM alpine:3.20
RUN adduser -D app
USER app
COPY --from=builder /out/adapter /app/adapter
ENTRYPOINT ["/app/adapter"]
