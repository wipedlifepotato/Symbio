<?php

namespace App\Controller;

use App\Service\MFrelance;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

class AdminDisputeController extends AbstractController
{
    private function getUsernameById($id, $jwt) {
        $userResponse = $this->mfrelance->doRequest("/profile/by_id?user_id={$id}", $jwt);
        if ($userResponse['httpCode'] == 200) {
            $data = json_decode($userResponse['response'], true);
            return $data['username'] ?? null; 
        }
        return null;
    }
    private MFrelance $mfrelance;

    public function __construct(MFrelance $mfrelance)
    {
        $this->mfrelance = $mfrelance;
    }

    #[Route('/admin/disputes', name: 'admin_disputes')]
    public function index(Request $request): Response
    {
        $jwt = $request->getSession()->get('jwt');
        if (!$jwt) {
            return $this->redirectToRoute('app_login');
        }

        // Build query string for filtering
        $queryParams = [];
        if ($request->query->get('id')) {
            $queryParams[] = 'id=' . $request->query->get('id');
        }
        if ($request->query->get('task_id')) {
            $queryParams[] = 'task_id=' . $request->query->get('task_id');
        }
        if ($request->query->get('status')) {
            $queryParams[] = 'status=' . $request->query->get('status');
        }

        $queryString = !empty($queryParams) ? '?' . implode('&', $queryParams) : '';
        $response = $this->mfrelance->doRequest('api/admin/disputes' . $queryString, $jwt);
        $disputes = [];

        if (200 === $response['httpCode']) {
            $data = json_decode($response['response'], true);
            $disputes = $data['disputes'] ?? [];
        } else {
            var_dump($response);
        }

        return $this->render('admin/disputes.html.twig', [
            'disputes' => $disputes,
        ]);
    }

    #[Route('/admin/disputes/{id}', name: 'admin_dispute_show')]
    public function show(int $id, Request $request): Response
    {
        $jwt = $request->getSession()->get('jwt');
        if (!$jwt) {
            return $this->redirectToRoute('app_login');
        }

        // Получаем детали диспута
        $response = $this->mfrelance->doRequest("/api/admin/disputes/details?id={$id}", $jwt);
        //var_dump($response);
        //exit(0);
        $dispute = null;
        $task = null;
        $escrow = null;
        $messages = [];

        if (200 === $response['httpCode']) {
            $data = json_decode($response['response'], true);
   
            $dispute = $data['dispute'] ?? null;
            $dispute['assigned_admin_username'] = $this->getUsernameById($dispute['assigned_admin'], $jwt);
            $dispute['opened_by_username'] = $this->getUsernameById($dispute['opened_by'], $jwt);

            $task = $data['task'] ?? null;
            $escrow = $data['escrow'] ?? null;
            $escrow['freelancer_username'] = $this->getUsernameById($escrow['freelancer_id'], $jwt);
            $messages = $data['messages'] ?? [];

            // Добавляем usernames к сообщениям
            foreach ($messages as &$message) {
                if (isset($message['sender_id'])) {
                    $userResponse = $this->mfrelance->doRequest("/profile/by_id?user_id={$message['sender_id']}", $jwt);
                    if (200 === $userResponse['httpCode']) {
                        $userData = json_decode($userResponse['response'], true);
                        $message['sender_username'] = $userData['username'] ?? 'Пользователь #' . $message['sender_id'];
                    } else {
                        $message['sender_username'] = 'Пользователь #' . $message['sender_id'];
                    }
                }
            }
        }

        if (!$dispute) {
            throw $this->createNotFoundException('Диспут не найден');
        }

        // Обработка действий
        if ($request->isMethod('POST')) {
            $action = $request->request->get('action');

            switch ($action) {
                case 'assign':
                    $assignData = ['dispute_id' => $id];
                    $response = $this->mfrelance->doRequest('api/admin/disputes/assign', $jwt, $assignData, true);
                    if (200 === $response['httpCode']) {
                        $this->addFlash('success', 'Диспут назначен вам!');
                    } else {
                        $this->addFlash('error', 'Ошибка назначения диспута: '.$response['response']);
                    }
                    break;

                case 'send_message':
                    $message = $request->request->get('message');
                    if ($message) {
                        $messageData = [
                            'dispute_id' => $id,
                            'message' => $message,
                        ];

                        $response = $this->mfrelance->doRequest('api/disputes/message', $jwt, $messageData, true);
                        if (200 === $response['httpCode']) {
                            $this->addFlash('success', 'Сообщение отправлено!');
                        } else {
                            $this->addFlash('error', 'Ошибка отправки сообщения');
                        }
                    }
                    break;

                case 'resolve':
                    $resolution = $request->request->get('resolution');
                    $resolveData = [
                        'dispute_id' => $id,
                        'resolution' => $resolution,
                    ];

                    $response = $this->mfrelance->doRequest('api/admin/disputes/resolve', $jwt, $resolveData, true);
                    if (200 === $response['httpCode']) {
                        $this->addFlash('success', 'Диспут разрешен!');

                        return $this->redirectToRoute('admin_disputes');
                    } else {
                        var_dump($response);
                         //exit(0);
                        $this->addFlash('error', 'Ошибка разрешения диспута:'.$response['response']);
                    }
                    break;
            }

            return $this->redirectToRoute('admin_dispute_show', ['id' => $id]);
        }

        return $this->render('admin/dispute_show.html.twig', [
            'dispute' => $dispute,
            'task' => $task,
            'escrow' => $escrow,
            'messages' => $messages,
        ]);
    }
}
