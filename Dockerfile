# Use official Go image
FROM golang:1.21-alpine

# Set working directory
WORKDIR /app

# Copy files
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build
RUN go build -o main .

# Run
CMD ["./main"]