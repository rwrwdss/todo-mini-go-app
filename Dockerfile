# Stage 1: build frontend
FROM node:20-alpine AS frontend
WORKDIR /build
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: build backend
FROM golang:1.22-alpine AS backend
WORKDIR /go
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

# Stage 3: runtime
FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=backend /server ./server
COPY --from=frontend /build/dist ./web/dist
EXPOSE 8080
CMD ["./server"]
