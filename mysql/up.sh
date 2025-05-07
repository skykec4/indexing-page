#!/bin/bash

# Start containers
docker-compose up -d

# Wait for MySQL to be ready
echo "Waiting for MySQL to be ready..."
while ! docker exec mariadb-container mysqladmin ping -h localhost -u root -pnhn2025!@ --silent; do
echo "MySQL is not ready yet..."
    sleep 1
done

# Run initialization script
echo "Initializing database..."
set -e
docker exec -i mariadb-container mysql -u root -pnhn2025!@ db_fe < ./init.sql