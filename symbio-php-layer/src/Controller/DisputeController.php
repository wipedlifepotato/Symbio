<?php

namespace App\Controller;

use App\Service\MFrelance;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

class DisputeController extends AbstractController
{
    private MFrelance $mfrelance;

    public function __construct(MFrelance $mfrelance)
    {
        $this->mfrelance = $mfrelance;
    }

    private function getUsernameById(int $userId, string $jwt): string
    {
        $userResponse = $this->mfrelance->doRequest("/profile/by_id?user_id={$userId}", $jwt);
        if (200 === $userResponse['httpCode']) {
            $userData = json_decode($userResponse['response'], true);
            return $userData['username'] ?? 'Пользователь #' . $userId;
        }
        return 'Пользователь #' . $userId;
    }

    #[Route('/disputes', name: 'disputes')]
    public function index(Request $request): Response
    {
        $jwt = $request->getSession()->get('jwt');
        if (!$jwt) {
            return $this->redirectToRoute('app_login');
        }

        $response = $this->mfrelance->doRequest('api/disputes/my', $jwt);
        $disputes = [];

        if (200 === $response['httpCode']) {
            $data = json_decode($response['response'], true);
            $disputes = $data['disputes'] ?? [];
        }

        return $this->render('dispute/index.html.twig', [
            'disputes' => $disputes,
        ]);
    }

    #[Route('/disputes/{id}', name: 'dispute_show')]
    public function show(int $id, Request $request): Response
    {
        $jwt = $request->getSession()->get('jwt');
        if (!$jwt) {
            return $this->redirectToRoute('app_login');
        }

        // Получаем диспут
        $response = $this->mfrelance->doRequest("api/disputes/get?id={$id}", $jwt);
        $dispute = null;

        if (200 === $response['httpCode']) {
            $data = json_decode($response['response'], true);
            $dispute = $data['dispute'] ?? null;
        }

        if (!$dispute) {
            throw $this->createNotFoundException('Диспут не найден');
        }

        // Получаем детали диспута (задачу, escrow и сообщения)
        $detailsResponse = $this->mfrelance->doRequest("api/disputes/get?id={$id}", $jwt);
       // var_dump($detailsResponse);
       // exit(0);
        $task = null;
        $escrow = null;
        $messages = [];
        $admin = null;

        if (200 === $detailsResponse['httpCode']) {
            $detailsData = json_decode($detailsResponse['response'], true);
            $task = $detailsData['task'] ?? null;
            $escrow = $detailsData['escrow'] ?? null;
            $messages = $detailsData['messages'] ?? [];
            $admin = $detailsData['admin'] ?? null;

            // Добавляем usernames для всех ID
            if ($dispute && isset($dispute['opened_by'])) {
                $dispute['opened_by_username'] = $this->getUsernameById($dispute['opened_by'], $jwt);
            }
            if ($task) {
                if (isset($task['client_id'])) {
                    $task['client_username'] = $this->getUsernameById($task['client_id'], $jwt);
                }
            }
            if ($escrow) {
                if (isset($escrow['client_id'])) {
                    $escrow['client_username'] = $this->getUsernameById($escrow['client_id'], $jwt);
                }
                if (isset($escrow['freelancer_id'])) {
                    $escrow['freelancer_username'] = $this->getUsernameById($escrow['freelancer_id'], $jwt);
                }
            }
        }

        // Добавляем usernames к сообщениям
        foreach ($messages as &$message) {
            if (isset($message['sender_id'])) {
                $message['sender_username'] = $this->getUsernameById($message['sender_id'], $jwt);
            }
        }

        // Получаем профиль пользователя
        $userResponse = $this->mfrelance->doRequest('api/profile', $jwt);
        $user = null;
        $userId = null;
        if (200 === $userResponse['httpCode']) {
            $user = json_decode($userResponse['response'], true);
            $userId = $user['user_id'] ?? null;
            $user['username'] = $this->getUsernameById($user['user_id'], $jwt);
        }

        // Проверяем, является ли пользователь админом
        $adminResponse = $this->mfrelance->doRequest('api/admin/IIsAdmin', $jwt);
        $isAdmin = false;
        if (200 === $adminResponse['httpCode']) {
            $adminData = json_decode($adminResponse['response'], true);
            $isAdmin = $adminData['is_admin'] ?? false;
        }

        // Проверяем, является ли пользователь участником диспута
        $isDisputeParticipant = false;
        if ($userId && $task) {
            $isDisputeParticipant = ($task['client_id'] == $userId || (isset($escrow['freelancer_id']) && $escrow['freelancer_id'] == $userId));
        }

        if ($user) {
            $user['is_admin'] = $isAdmin;
            $user['is_dispute_participant'] = $isDisputeParticipant;
        }

        // Обработка отправки сообщения
        if ($request->isMethod('POST')) {
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

                return $this->redirectToRoute('dispute_show', ['id' => $id]);
            }
        }

        return $this->render('dispute/show.html.twig', [
            'dispute' => $dispute,
            'messages' => $messages,
            'admin' => $admin,
            'user' => $user,
            'task' => $task,
            'escrow' => $escrow,
        ]);
    }

    #[Route('/disputes/create/{taskId}', name: 'dispute_create')]
    public function create(int $taskId, Request $request): Response
    {
        $jwt = $request->getSession()->get('jwt');
        if (!$jwt) {
            return $this->redirectToRoute('app_login');
        }

        $disputeData = ['task_id' => $taskId];
        $response = $this->mfrelance->doRequest('api/disputes/create', $jwt, $disputeData, true);
        // var_dump($response);
        // exit(0);
        if (200 === $response['httpCode']) {
            $data = json_decode($response['response'], true);
            $disputeId = $data['dispute']['id'] ?? null;
            if ($disputeId) {
                $this->addFlash('success', 'Диспут создан!');

                return $this->redirectToRoute('dispute_show', ['id' => $disputeId]);
            }
        }

        $this->addFlash('error', 'Ошибка создания диспута');

        return $this->redirectToRoute('task_show', ['id' => $taskId]);
    }
}
