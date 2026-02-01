#!/bin/bash

# Переходим в корень проекта
cd /home/glushkova/Desktop/diplom_PP/golang-edugame

echo "🔧 Подготовка окружения разработки..."

# Проверяем наличие mkcert
if ! command -v mkcert &> /dev/null; then
    echo "❌ mkcert не установлен. Установите его:"
    echo "   Ubuntu: sudo apt install libnss3-tools && curl -JLO 'https://dl.filippo.io/mkcert/latest?for=linux/amd64'"
    echo "   Затем переименуйте файл и переместите в /usr/local/bin/mkcert"
    exit 1
fi

# Создаем сертификаты если их нет
if [ ! -f "certs/localhost.pem" ]; then
    echo "🔐 Создание SSL сертификатов..."
    mkdir -p certs
    mkcert -install
    mkcert -key-file certs/localhost-key.pem -cert-file certs/localhost.pem localhost 127.0.0.1 ::1
    echo "✅ Сертификаты созданы"
fi

# Экспортируем переменные окружения
export PORT=3000
export DATABASE_URL="postgres://user:pass@localhost:5432/math_trainer?sslmode=disable"

echo "🚀 Запуск Go сервера на https://localhost:3000"
echo "   При первом запуске нажмите 'Принять риск' в браузере"

# Запускаем сервер
go run cmd/server/main.go