# Multi-stage Dockerfile for Crypto Trading App

# Stage 1: Build the React frontend
FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Stage 2: Build the Go backend
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o crypto-trading-server ./cmd/trading

# Stage 3: Final image
FROM alpine:3.18
WORKDIR /app
RUN apk --no-cache add ca-certificates

# Copy the compiled backend
COPY --from=backend-builder /app/crypto-trading-server /app/

# Copy the frontend build
COPY --from=frontend-builder /app/frontend/build /app/static

# Expose the port
EXPOSE 8080

# Command to run the application
CMD ["./crypto-trading-server"]
