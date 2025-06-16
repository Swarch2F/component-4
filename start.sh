#!/bin/bash
# filepath: start.sh
# Mostrar la ubicación actual
echo "Ubicación actual: $(pwd)"

# Cargar variables de entorno
set -a
source .env
set +a

# Levantar los servicios con Docker Compose
if [ "$1" = "-d" ]; then
    export DB_HOST=localhost
    export PORT=8082
    go run cmd/main.go
else
    docker-compose up --build
fi