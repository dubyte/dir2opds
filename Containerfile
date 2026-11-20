# Stage 1: Build
# Keep this version in sync with the toolchain line in go.mod: the golang image
# sets GOTOOLCHAIN=local, so a go.mod requiring a newer toolchain fails here.
FROM docker.io/golang:1.27.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# CGO_ENABLED=0 is explicit on purpose: the final stage has no libc, so the
# binary must be static even if the builder image ever ships a C compiler.
RUN CGO_ENABLED=0 go build -o /app/dir2opds

# Stage 2: Final Image
FROM scratch
COPY --from=builder /app/dir2opds /dir2opds
ENTRYPOINT ["/dir2opds"]
