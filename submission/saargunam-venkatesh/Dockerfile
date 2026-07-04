# Stage 1: Build — compile a fully static binary
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Download deps first — this layer is cached unless go.mod/go.sum change
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 → static binary, no libc dependency
# -ldflags "-w -s" → strip debug info, shrink binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o config-service ./cmd

# Stage 2: Run — distroless has no shell, no package manager, minimal attack surface
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /build/config-service /config-service

EXPOSE 8080

ENTRYPOINT ["/config-service"]
