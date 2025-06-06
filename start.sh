#!/bin/bash
# filepath: start.sh

# Exportar todas las variables del archivo .env
export $(grep -v '^#' .env | xargs)

# Levantar los servicios con Docker Compose
docker-compose up --build