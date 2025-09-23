<?php

namespace App\Service;

use Symfony\Contracts\HttpClient\HttpClientInterface;

class MFrelance
{
    private $addr;
    private $client;

    public function __construct(HttpClientInterface $client, string $addr = 'localhost', int $port = 9999)
    {
        $this->addr = 'http://'.$addr.':'.$port.'/';
        $this->client = $client;
    }

    public function getCaptcha(): array
    {
    	try {
		$response = $this->client->request('GET', $this->addr.'captcha', [
		    'headers' => [
		        'Accept' => 'image/png',
		    ],
		]);

		$captchaId = $response->getHeaders(false)['x-captcha-id'][0] ?? '';
		
		$captchaImg = 'data:image/png;base64,'.base64_encode($response->getContent(false));
		
		return [
		    'captchaID' => $captchaId,
		    'captchaImg' => $captchaImg,
		];
        } catch(Exception $e) {
        	return [];
        }
    }

    public function doRequest(string $page, ?string $jwt = null, ?array $postData = null, bool $isPost = false): array
    {
        $options = [
            'headers' => ['Content-Type' => 'application/json'],
        ];

        if ($jwt) {
            $options['headers']['Authorization'] = "Bearer $jwt";
        }

        if ($isPost && null !== $postData) {
            $options['json'] = $postData;
        }

        $method = $isPost ? 'POST' : 'GET';

        try {
            $response = $this->client->request($method, $this->addr.$page, $options);
            $content = $response->getContent(false); // false — не бросать исключение для HTTP 4xx/5xx
            $status = $response->getStatusCode();
        } catch (TransportExceptionInterface $e) {
            $content = $e->getMessage();
            $status = 0;
        } catch (ClientExceptionInterface|ServerExceptionInterface $e) {
            $content = $e->getMessage();
            $status = $e->getCode() ?: 0;
        }

        return [
            'response' => $content,
            'httpCode' => $status,
        ];
    }
}
