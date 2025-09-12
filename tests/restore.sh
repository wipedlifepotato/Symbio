#!/bin/bash

curl -s -o captcha.png -D headers.txt http://localhost:9999/captcha

echo "Print captchaID: "
read CaptchaID
echo "Print answer: "
read CaptchaAnswer

echo "Print mnemonic: "
read Mnemonic
echo "Print new password: "
read NewPassword

JSON=$(jq -n \
    --arg username "testuser" \
    --arg mnemonic "$Mnemonic" \
    --arg new_password "$NewPassword" \
    --arg captcha_id "$CaptchaID" \
    --arg captcha_answer "$CaptchaAnswer" \
    '{username:$username, mnemonic:$mnemonic, new_password:$new_password, captcha_id:$captcha_id, captcha_answer:$captcha_answer}')

curl -s -X POST http://localhost:9999/restoreuser \
     -H "Content-Type: application/json" \
     -d "$JSON" \
     -o restore_response.json

echo "Ответ сервера сохранён в restore_response.json"
