#!/bin/bash

USERNAME="testuser"
PASSWORD="12345678"

curl -s -o captcha.png -D headers.txt http://localhost:9999/captcha
cat headers.txt
echo "Print captchaID: "
read CaptchaID
echo "Print captcha answer: "
read CaptchaAnswer

JSON=$(jq -n \
    --arg username "$USERNAME" \
    --arg password "$PASSWORD" \
    --arg captcha_id "$CaptchaID" \
    --arg captcha_answer "$CaptchaAnswer" \
    '{username:$username, password:$password, captcha_id:$captcha_id, captcha_answer:$captcha_answer}')

AUTH_RESPONSE=$(curl -s -X POST http://localhost:9999/auth \
    -H "Content-Type: application/json" \
    -d "$JSON")

JWT=$(echo "$AUTH_RESPONSE" | jq -r '.token')
if [ "$JWT" == "null" ] || [ -z "$JWT" ]; then
    echo "Ошибка: не удалось получить JWT"
    exit 1
fi
echo "JWT получен: $JWT"

USER_ID=3

echo "Проверяем, админ ли пользователь"
curl -s -X POST "http://localhost:9999/api/admin/check" \
    -H "Authorization: Bearer $JWT" \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": $USER_ID}" 

echo "Делаем пользователя админом"
curl -s -X POST "http://localhost:9999/api/admin/make" \
    -H "Authorization: Bearer $JWT" \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": $USER_ID}" 

echo "Снова проверяем, админ ли пользователь"
curl -s -X POST "http://localhost:9999/api/admin/check" \
    -H "Authorization: Bearer $JWT" \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": $USER_ID}" 

echo "Удаляем админку у пользователя"
curl -s -X POST "http://localhost:9999/api/admin/remove" \
    -H "Authorization: Bearer $JWT" \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": $USER_ID}" 

echo "Финальная проверка админки"
curl -s -X POST "http://localhost:9999/api/admin/check" \
    -H "Authorization: Bearer $JWT" \
    -H "Content-Type: application/json" \
   -d "{\"user_id\": $USER_ID}" 

