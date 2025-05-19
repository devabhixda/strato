FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Add verbose flag and build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -o server .

# Use a different distroless variant specifically for running Go binaries
FROM gcr.io/distroless/static-debian12

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]