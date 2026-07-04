FROM mirror.gcr.io/library/golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server .

FROM mirror.gcr.io/library/alpine:3.20
RUN apk add --no-cache ca-certificates tini
COPY --from=builder /app/server /app/server
EXPOSE 8080
ENTRYPOINT ["/sbin/tini", "--", "/app/server"]