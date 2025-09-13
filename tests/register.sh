GPG_KEY=`cat gpgkey.txt`

curl -s -o captcha.png -D headers.txt http://localhost:9999/captcha
print "Print captchaID: "
read CaptchaID
printf "Print answer: "
read CaptchaAnswer

JSON=$(jq -n \
    --arg username "testuser1" \
    --arg password "123456" \
    --arg gpg_key "$GPG_KEY" \
    --arg captcha_id "$CaptchaID" \
    --arg captcha_answer "$CaptchaAnswer" \
    '{username:$username, password:$password, gpg_key:$gpg_key, captcha_id:$captcha_id, captcha_answer:$captcha_answer}')

curl -s -X POST http://localhost:9999/register \
     -H "Content-Type: application/json" \
     -d "$JSON" \
     -o register_response.json

echo "Ответ сервера сохранён в register_response.json"

