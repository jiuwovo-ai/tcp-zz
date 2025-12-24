# Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Build backend
FROM golang:1.21-alpine AS backend-builder
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o port-forward-dashboard .

# Final image
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY --from=backend-builder /app/port-forward-dashboard .
COPY --from=frontend-builder /app/backend/static ./static
EXPOSE 8080
CMD ["./port-forward-dashboard"]
