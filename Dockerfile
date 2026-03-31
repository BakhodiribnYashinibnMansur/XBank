# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Dependency larni avval copy qilamiz (cache uchun)
COPY go.mod go.sum ./
RUN go mod download

# Barcha kodni copy qilamiz
COPY . .

# Binary build
RUN CGO_ENABLED=0 GOOS=linux go build -o xbank ./cmd/api

# Run stage - kichik image
FROM alpine:3.21

WORKDIR /app

# Builder dan faqat binary ni olamiz
COPY --from=builder /app/xbank .

EXPOSE 3000

CMD ["./xbank"]
