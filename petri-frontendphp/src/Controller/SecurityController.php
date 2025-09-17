<?php
namespace App\Controller;

use App\Service\MFrelance;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;
use Symfony\Component\HttpFoundation\Session\SessionInterface;
use Symfony\Component\HttpFoundation\Request;

class SecurityController extends AbstractController
{
#[Route('/auth', name: 'app_auth', methods: ['POST'])]
public function auth(Request $request, MFrelance $mfrelance, SessionInterface $session): Response
{
    $username = $request->request->get('username', '');
    $password = $request->request->get('password', '');
    $captchaId = $request->request->get('captcha_id', '');
    $captchaAnswer = $request->request->get('captcha_answer', '');

    $data = [
        'username' => $username,
        'password' => $password,
        'captcha_id' => $captchaId,
        'captcha_answer' => $captchaAnswer,
    ];

    try {
        $result = $mfrelance->doRequest('auth', false, $data, true);

        if ($result['httpCode'] === 200) {
            $json = json_decode($result['response'], true);
            $jwt = $json['token'] ?? '';
            if ($jwt) {
                $session->set('jwt', $jwt);
                return $this->redirectToRoute('app_dashboard');
            }
            $message = 'Ошибка: JWT не получен';
        } else {
            // Парсим текст ошибки из JSON, если сервер вернул JSON
            $json = json_decode($result['response'], true);
            if (isset($json['error'])) {
                $message = "Ошибка авторизации: " . $json['error'];
            } else {
                $message = "Ошибка авторизации: " . $result['response'];
            }
        }

    } catch (\Exception $e) {
        // Исключение сетевого запроса

        $message = 'Ошибка запроса: ' . $e->getMessage();
    }

    // Получаем новую капчу для повторного ввода
    $captcha = $mfrelance->getCaptcha();
    $captchaId = $captcha['captchaID'] ?? '';
    $captchaImage = $captcha['captchaImg'] ?? '';
    $projectName = $this->getParameter('project_name');

    return $this->render('security/index.html.twig', [
        'projectName' => $projectName,
        'captchaId' => $captchaId,
        'captchaImage' => $captchaImage,
        'message' => $message,
    ]);
}


#[Route('/restore', name: 'app_restore')]
public function restore(MFrelance $mfrelance, Request $request, SessionInterface $session): Response
{
    $message = '';
    $jwt = $session->get('jwt', '');

    // Получаем капчу
    $captcha = $mfrelance->getCaptcha();
    $captchaId = $captcha['captchaID'] ?? '';
    $captchaImage = $captcha['captchaImg'] ?? '';

    if ($request->isMethod('POST')) {
        $username = $request->request->get('username', '');
        $mnemonic = $request->request->get('mnemonic', '');
        $newPassword = $request->request->get('new_password', '');
        $captchaAnswer = $request->request->get('captcha_answer', '');
        $captchaIdPost = $request->request->get('captcha_id', '');

        $data = [
            'username'       => $username,
            'mnemonic'       => $mnemonic,
            'new_password'   => $newPassword,
            'captcha_id'     => $captchaIdPost,
            'captcha_answer' => $captchaAnswer,
        ];

        try {
            $response = $mfrelance->doRequest('restoreuser', false, $data, true);
            if ($response['httpCode'] === 200) {
                $json = json_decode($response['response'], true);
                $message = $json['message'] ?? 'Аккаунт восстановлен';
                $encoded = $json['encrypted'] ?? '';
                if ($encoded) {
                    $session->set('jwt', $encoded);
                    return $this->redirectToRoute('app_dashboard');
                }
            } else {
                $message = "Ошибка восстановления: " . $response['response'];
            }
        } catch (\Exception $e) {
            $message = "Ошибка запроса: " . $e->getMessage();
        }
    }

    $projectName = $this->getParameter('project_name');

    return $this->render('security/restore.html.twig', [
        'projectName' => $projectName,
        'captchaId'   => $captchaId,
        'captchaImage'=> $captchaImage,
        'message'     => $message,
    ]);
}

#[Route('/register', name: 'app_register')]
public function register(MFrelance $mfrelance, Request $request): Response
{
    $message = '';
    $mnemonic = '';

    // Получаем капчу только при GET
    if (!$request->isMethod('POST')) {
        $captcha = $mfrelance->getCaptcha();
        $captchaId = $captcha['captchaID'] ?? '';
        $captchaImage = $captcha['captchaImg'] ?? '';
    } else {
        // При POST используем captcha_id из формы
        $captchaId = $request->request->get('captcha_id', '');
        $captchaImage = ''; // оставляем пустым, можно перегенерить позже
    }

    if ($request->isMethod('POST')) {
        $username = $request->request->get('username', '');
        $password = $request->request->get('password', '');
        $captchaAnswer = $request->request->get('captcha_answer', '');

        $data = [
            'username' => $username,
            'password' => $password,
            'captcha_id' => $captchaId,
            'captcha_answer' => $captchaAnswer,
        ];

        try {
            $result = $mfrelance->doRequest('register', false, $data, true);
            $json = json_decode($result['response'], true);

            if ($result['httpCode'] === 200) {
                $message = $json['message'] ?? 'Регистрация успешна';
                $mnemonic = $json['encrypted'] ?? '';
            } elseif ($result['httpCode'] === 400) {
                // Специально обрабатываем капчу
                $errorMsg = $json['error'] ?? $result['response'];
                $message = "Ошибка регистрации: " . $errorMsg;
            } else {
                $message = "Ошибка регистрации: " . $result['response'];
            }
        } catch (\Exception $e) {
            $message = "Ошибка регистрации: " . $e->getMessage();
        }

        // При неудачной регистрации подгружаем новую капчу
        $captcha = $mfrelance->getCaptcha();
        $captchaId = $captcha['captchaID'] ?? '';
        $captchaImage = $captcha['captchaImg'] ?? '';
    }

    $projectName = $this->getParameter('project_name');

    return $this->render('security/register.html.twig', [
        'projectName' => $projectName,
        'captchaId' => $captchaId,
        'captchaImage' => $captchaImage,
        'message' => $message,
        'mnemonic' => $mnemonic,
    ]);
}
    	#[Route('/', name: 'app_login')]
    	public function index(MFrelance $mfrelance, SessionInterface $session): Response
    	{
		if ($session->has('jwt')) {
		    return $this->redirectToRoute('app_dashboard');
		}

		$captcha = $mfrelance->getCaptcha();
		$captchaId = $captcha['captchaID'] ?? '';
		$captchaImage = $captcha['captchaImg'] ?? '';
		$projectName = $this->getParameter('project_name');
		return $this->render('security/index.html.twig', [
		     'projectName' => $projectName,
		    'captchaId' => $captchaId,
		    'captchaImage' => $captchaImage,
		]);
    	}

}

