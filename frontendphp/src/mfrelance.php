<?php
class MFrelance {
    function __construct($addr='localhost', $port=9999)
    {
        $this->addr = 'http://'.$addr.":".$port."/";
    }

    function prepareRequest($page)
    {
        $ch = curl_init($this->addr.$page);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        return $ch;
    }

    function getCaptcha() {
        $ch = $this->prepareRequest("captcha");
        curl_setopt($ch, CURLOPT_HEADER, true);
        $response = curl_exec($ch);
        $headerSize = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
        $headers = substr($response, 0, $headerSize);
        $body = substr($response, $headerSize);

        if (preg_match('/X-Captcha-ID:\s*(\S+)/i', $headers, $matches)) {
            $captchaId = trim($matches[1]);
        }
        $captchaImage = 'data:image/png;base64,' . base64_encode($body);
        curl_close($ch);
        return ["captchaID" => $captchaId ?? '', "captchaImg" => $captchaImage];
    }


    function doRequest($page, $jwt = false, $postData = null, $isPost = false) 
    {
        $ch = $this->prepareRequest($page);

        $headers = ["Content-Type: application/json"];
        if ($jwt) {
            $headers[] = "Authorization: Bearer $jwt";
        }
        curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);

        if ($isPost && $postData !== null) {
            curl_setopt($ch, CURLOPT_POST, true);
            curl_setopt($ch, CURLOPT_POSTFIELDS, $postData);
        }

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);

        return ["response" => $response, "httpCode" => $httpCode];
    }
}

