#!/bin/bash

USERNAME="testuser"
PASSWORD="12345678"
DEST="tb1q9tutjha6055xy3akk55djruhcxusxvzmsrxpk3"
AMOUNT="0.00011"

curl -s -o captcha.png -D headers.txt http://localhost:9999/captcha
echo "Print captchaID: "
cat headers.txt
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

SEND_RESPONSE=$(curl -s -X POST "http://localhost:9999/api/wallet/bitcoinSend?to=$DEST&amount=$AMOUNT" \
    -H "Authorization: Bearer $JWT")
echo "RAW RESPONSE:"
echo "$SEND_RESPONSE"

if echo "$SEND_RESPONSE" | jq empty 2>/dev/null; then
    echo "$SEND_RESPONSE" | jq
fi

