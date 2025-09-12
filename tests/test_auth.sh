#!/bin/bash

USERNAME="testuser"
PASSWORD="12345678"

curl -s -o captcha.png -D headers.txt http://localhost:9999/captcha
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

echo "Ответ /auth:"
echo "$AUTH_RESPONSE" | jq

JWT=$(echo "$AUTH_RESPONSE" | jq -r '.token')

if [ "$JWT" == "null" ] || [ -z "$JWT" ]; then
    echo "Ошибка: не удалось получить JWT"
    exit 1
fi

echo "JWT получен: $JWT"

TEST_RESPONSE=$(curl -s -X GET http://localhost:9999/api/test \
    -H "Authorization: Bearer $JWT")

echo "Ответ /api/test:"
echo "$TEST_RESPONSE" | jq

