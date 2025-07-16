# Build the React-TypeScript frontend
FROM node:24.1.0-alpine AS builder

WORKDIR /app/frontend

COPY frontend/package*.json ./

RUN npm install

COPY frontend/ ./

RUN npm run build

# Build the Go backend
FROM golang:1.24.3-alpine AS go-builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

COPY --from=builder /app/frontend/dist ./frontend/dist

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /server ./cmd/api/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=go-builder /server /app/server

EXPOSE 8000

CMD ["/app/server"]
