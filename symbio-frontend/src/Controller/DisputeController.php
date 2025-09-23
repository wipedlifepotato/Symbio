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

        // Получаем сообщения диспута
        $messages = $data['messages'] ?? [];

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
