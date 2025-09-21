<?php

namespace App\Controller;

use App\Service\MFrelance;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

class ReviewController extends AbstractController
{
    private MFrelance $mfrelance;

    public function __construct(MFrelance $mfrelance)
    {
        $this->mfrelance = $mfrelance;
    }

    #[Route('/reviews/create/{taskId}', name: 'review_create')]
    public function create(int $taskId, Request $request): Response
    {
        $jwt = $request->getSession()->get('jwt');
        if (!$jwt) {
            return $this->redirectToRoute('app_login');
        }

        if ($request->isMethod('POST')) {
            $reviewData = [
                'task_id' => $taskId,
                'rating' => (int) $request->request->get('rating'),
                'comment' => $request->request->get('comment'),
            ];

            $response = $this->mfrelance->doRequest('api/reviews/create', $jwt, $reviewData, true);

            if (200 === $response['httpCode']) {
                $this->addFlash('success', 'Отзыв оставлен!');

                return $this->redirectToRoute('task_show', ['id' => $taskId]);
            } else {
                $this->addFlash('error', 'Ошибка создания отзыва');
            }
        }

        return $this->render('review/create.html.twig', [
            'taskId' => $taskId,
        ]);
    }

    #[Route('/reviews/user/{userId}', name: 'review_user')]
    public function userReviews(int $userId, Request $request): Response
    {
        $jwt = $request->getSession()->get('jwt');
        if (!$jwt) {
            return $this->redirectToRoute('app_login');
        }

        $response = $this->mfrelance->doRequest("api/reviews/user?user_id={$userId}", $jwt);
        $reviews = [];

        if (200 === $response['httpCode']) {
            $data = json_decode($response['response'], true);
            $reviews = $data['reviews'] ?? [];
        }

        return $this->render('review/user.html.twig', [
            'reviews' => $reviews,
            'userId' => $userId,
        ]);
    }
}
