#!/bin/bash
# build.sh - Сборка Docker-образа для проекта с Colima

echo "Сборка Docker-образа..."
docker build -t what-your-type-bot .
echo "Сборка завершена!"
