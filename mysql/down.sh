#!/bin/bash
docker exec -i mariadb-container mysql -u root -pnhn2025!@ db_fe < ./down.sql
echo 'good'

docker-compose down