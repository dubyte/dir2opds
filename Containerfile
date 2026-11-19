# Stage 1: Build
FROM docker.io/golang:1.27.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/dir2opds

# Stage 2: Final Image
FROM scratch
COPY --from=builder /app/dir2opds /dir2opds
ENTRYPOINT ["/dir2opds"]
