#!/bin/bash
# filepath: start.sh

# Cargar variables de entorno
set -a
source .env
set +a

# Levantar los servicios con Docker Compose
docker-compose up --build