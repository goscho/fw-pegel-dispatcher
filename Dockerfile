# Build
FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/fw-pegel-dispatcher ./cmd/fw-pegel-dispatcher

# Runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates \
	&& adduser -D -u 1001 -h /app -s /sbin/nologin appuser

USER appuser
WORKDIR /app

COPY --from=build /out/fw-pegel-dispatcher /app/fw-pegel-dispatcher

ENTRYPOINT ["/app/fw-pegel-dispatcher"]
