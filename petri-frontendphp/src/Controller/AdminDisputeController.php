<?php

namespace App\Controller;

use App\Service\MFrelance;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

class AdminDisputeController extends AbstractController
{
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

        $response = $this->mfrelance->doRequest('api/admin/disputes', $jwt);
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
        $response = $this->mfrelance->doRequest("api/admin/disputes/details?id={$id}", $jwt);
        $dispute = null;
        $task = null;
        $escrow = null;
        $messages = [];

        if (200 === $response['httpCode']) {
            $data = json_decode($response['response'], true);
            $dispute = $data['dispute'] ?? null;
            $task = $data['task'] ?? null;
            $escrow = $data['escrow'] ?? null;
            $messages = $data['messages'] ?? [];
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
                        $this->addFlash('error', 'Ошибка назначения диспута');
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
                        // exit(0);
                        $this->addFlash('error', 'Ошибка разрешения диспута');
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
