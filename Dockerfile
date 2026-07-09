# ============================================================
# Dockerfile — HRMS Backend (Go/Fiber)
# Multi-stage: development (Air hot reload) + production (alpine)
# ============================================================

# ---- Stage 1: Base ----
FROM golang:1.26-alpine AS base
RUN apk add --no-cache git ca-certificates wget
# Install goose migration tool (build-time, not runtime)
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

WORKDIR /app

# ---- Stage 2: Development (Air hot reload) ----
FROM base AS dev

# Install Air for hot reload
RUN go install github.com/air-verse/air@latest

# Copy go.mod and go.sum first for dependency caching
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the rest of the source (includes .air.toml)
COPY backend/ .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

# ---- Stage 3: Build for production ----
FROM base AS builder

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o /hrms-server .

# ---- Stage 3b: Build frontend ----
FROM node:22-alpine AS frontend-builder

WORKDIR /app

# Install dependencies
COPY frontend/package.json frontend/package-lock.json frontend/.npmrc ./
RUN npm ci

# Copy source & build
COPY frontend/ .
RUN npm run build

# ---- Stage 4: Production (minimal image) ----
FROM alpine:3.20 AS prod

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /hrms-server .

# Copy built frontend into public/ directory
COPY --from=frontend-builder /app/build ./public

# Create uploads directory
RUN mkdir -p /app/uploads

EXPOSE 8080

CMD ["./hrms-server"]
