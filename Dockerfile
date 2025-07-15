# syntax=docker/dockerfile:1

# Etapa 1: build
FROM golang:1.23-alpine AS builder
 
WORKDIR /app

# Copia los archivos go.mod y go.sum y descarga dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copia el resto del código fuente
COPY . .

# Genera la documentación de Swagger (si usas herramientas como swagger)
# RUN go install github.com/swaggo/swag/cmd/swag@latest
# RUN swag init --parseDependency --parseInternal -g cmd/main.go

# Compila el binario
RUN go build -o component-4 ./cmd/main.go

# Etapa 2: imagen final
FROM alpine:3.18

WORKDIR /app

# Copia el binario desde la etapa anterior
COPY --from=builder /app/component-4 .

# COPIA LA CARPETA DE MIGRACIONES
COPY --from=builder /app/migrations ./migrations

# Copia los archivos de documentación de Swagger generados (si existen)
# COPY --from=builder /app/docs ./docs

# Copia el .env si lo necesitas (opcional, solo para pruebas locales)
# COPY .env .env

# Expone el puerto (ajusta si usas otro)
EXPOSE 8080

# Comando por defecto
CMD ["./component-4"]