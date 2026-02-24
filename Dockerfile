# Stage 1: Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ ./
ARG VITE_PUBLIC_URL=""
ENV VITE_PUBLIC_URL=${VITE_PUBLIC_URL}
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24-alpine AS backend
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -o otscan ./cmd/otscan

# Stage 3: Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /app/otscan .
COPY --from=backend /app/web/dist ./web/dist
COPY --from=backend /app/migrations ./migrations
EXPOSE 3000
ENTRYPOINT ["./otscan", "--config", "/app/config.yaml"]
