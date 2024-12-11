# Stage 1: Build
FROM docker.io/golang:1.23.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/dir2opds

# Stage 2: Final Image
FROM docker.io/alpine
COPY --from=builder /app/dir2opds /dir2opds
