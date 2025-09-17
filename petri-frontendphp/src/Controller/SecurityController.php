<?php

namespace App\Controller;

use App\Service\MFrelance;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpFoundation\Session\SessionInterface;
use Symfony\Component\Routing\Annotation\Route;
use Symfony\Contracts\Translation\TranslatorInterface;

class SecurityController extends AbstractController
{
    #[Route('/auth', name: 'app_auth', methods: ['POST'])]
    public function auth(Request $request, MFrelance $mfrelance, SessionInterface $session, TranslatorInterface $translator): Response
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

            if (200 === $result['httpCode']) {
                $json = json_decode($result['response'], true);
                $jwt = $json['token'] ?? '';
                if ($jwt) {
                    $session->set('jwt', $jwt);

                    return $this->redirectToRoute('app_dashboard');
                }
                $message = $translator->trans('auth.error_jwt');
            } else {

                $json = json_decode($result['response'], true);
                if ($json['error'] == 'invalid captcha')
                	$json['error'] = $translator->trans('bad_captcha', []);
                else if($json['error'] == 'invalid username or password')
                	$json['error'] = $translator->trans('invalidusernameopassword', []);
                if (isset($json['error'])) {
                     $message = $translator->trans('auth.login_error', ['%error%' => $json['error'] ?? $result['response']]);
                } else {
                     $message = $translator->trans('auth.login_error', ['%error%' => $json['error'] ?? $result['response']]);
                }
            }
        } catch (\Exception $e) {
	    $message = $translator->trans('auth.request_error', ['%error%' => $e->getMessage()]);
        }


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
    public function restore(MFrelance $mfrelance, Request $request, SessionInterface $session, TranslatorInterface $translator): Response
    {
        $message = '';
        $jwt = $session->get('jwt', '');


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
                'username' => $username,
                'mnemonic' => $mnemonic,
                'new_password' => $newPassword,
                'captcha_id' => $captchaIdPost,
                'captcha_answer' => $captchaAnswer,
            ];

            try {
                $response = $mfrelance->doRequest('restoreuser', false, $data, true);
                if (200 === $response['httpCode']) {
                    $json = json_decode($response['response'], true);
                    $message = $json['message'] ?? 'Аккаунт восстановлен';
                    $encoded = $json['encrypted'] ?? '';
                    if ($encoded) {
                        $session->set('jwt', $encoded);

                        return $this->redirectToRoute('app_dashboard');
                    }
                } else {
                    $json = json_decode($response['response'], true);
                    if ($json['error']=='password too small')
                    {
                    	$json['error'] = $translator->trans('passwordtoosmall', []);;
                    } else if($json['error'] == 'invalid captcha')
                    {
                    	$json['error'] = $translator->trans('bad_captcha', []);;
                    } else if($json['error'] == 'failed to found user') 
                    {
                    	$json['error'] = $translator->trans('notFoundUser', []);;
                    }
                    $message = $translator->trans('restore.error', ['%error%' => $json['error']]);
                }
            } catch (\Exception $e) {
                $message = 'Request error: '.$e->getMessage();
            }
        }

        $projectName = $this->getParameter('project_name');

        return $this->render('security/restore.html.twig', [
            'projectName' => $projectName,
            'captchaId' => $captchaId,
            'captchaImage' => $captchaImage,
            'message' => $message,
        ]);
    }

#[Route('/register', name: 'app_register')]
public function register(MFrelance $mfrelance, Request $request, TranslatorInterface $translator): Response
{
    $message = '';
    $mnemonic = '';

    // Получаем капчу
    if (!$request->isMethod('POST')) {
        $captcha = $mfrelance->getCaptcha();
        $captchaId = $captcha['captchaID'] ?? '';
        $captchaImage = $captcha['captchaImg'] ?? '';
    } else {
        $captchaId = $request->request->get('captcha_id', '');
        $captchaImage = '';
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

            if (200 === $result['httpCode']) {
                $message = $translator->trans('register.success', [], null);
                $mnemonic = $json['encrypted'] ?? '';
            } elseif (400 === $result['httpCode']) {
                $errorMsg = $json['error'] ?? $result['response'];
                if ($errorMsg === 'password too small') {
                	$errorMsg = $translator->trans('passwordtoosmall', []);
                } else if ($errorMsg == 'invalid captcha' ){
                	$errorMsg = $translator->trans('bad_captcha', []);
                }
                $message = $translator->trans('register.error', ['%error%' => $errorMsg], null);
            } else {
                $message = $translator->trans('register.error', ['%error%' => $result['response']], null);
            }
        } catch (\Exception $e) {
            $message = $translator->trans('register.error', ['%error%' => $e->getMessage()], null);
        }

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
    public function index(MFrelance $mfrelance, SessionInterface $session, TranslatorInterface $translator): Response
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
