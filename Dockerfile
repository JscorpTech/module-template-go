# Multi-stage build — kichik va xavfsiz image.
# Builder: golang:alpine — full toolchain
# Runtime: alpine — minimal (~10MB)

# ---- Builder ----
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Manba kod + vendor/ (SDK shu yerda — build paytida network kerak emas).
COPY . .

# vendor/ mavjud → -mod=vendor offline build. CGO o'chiq → statik binary.
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -trimpath -ldflags="-s -w" -o /module .

# ---- Runtime ----
FROM alpine:3.20

# Non-root foydalanuvchi (xavfsizlik).
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /module ./module

# CA sertifikatlari (HTTPS so'rovlar uchun — ixtiyoriy, kerak bo'lsa saqlang).
RUN apk --no-cache add ca-certificates

USER appuser

ENV PORT=8100
EXPOSE 8100

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8100/health || exit 1

CMD ["/app/module"]
