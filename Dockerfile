# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Dependency larni avval copy qilamiz (cache uchun)
COPY go.mod go.sum ./
RUN go mod download

# Barcha kodni copy qilamiz
COPY . .

# Binary build (reproducible, stripped — A08 Software Integrity)
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o xbank ./cmd/app

# Run stage - kichik image
FROM alpine:3.21

# CA certificates (HTTPS chiquvchi so'rovlar uchun) va timezone data
RUN apk add --no-cache ca-certificates tzdata

# Non-root user yaratish (security best practice)
RUN addgroup -S xbank && adduser -S xbank -G xbank

WORKDIR /app

# Builder dan faqat binary ni olamiz
COPY --from=builder /app/xbank .

# Non-root user bilan ishga tushirish
USER xbank

EXPOSE 3000

CMD ["./xbank"]
