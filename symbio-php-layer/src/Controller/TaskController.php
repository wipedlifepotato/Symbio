<?php

namespace App\Controller;

use App\Service\MFrelance;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

class TaskController extends AbstractController
{
    private MFrelance $mfrelance;

    public function __construct(MFrelance $mfrelance)
    {
        $this->mfrelance = $mfrelance;
    }

    #[Route('/tasks', name: 'tasks')]
    public function index(Request $request): Response
    {
        $jwt = $request->getSession()->get('jwt');
        if (!$jwt) {
            return $this->redirectToRoute('app_login');
        }

        $status = $request->query->get('status', 'all');
        $page = max(1, (int) $request->query->get('page', 1));
        $perPage = min(100, max(1, (int) $request->query->get('perPage', 20)));
        $offset = ($page - 1) * $perPage;

        $query = [];
        if ('all' !== $status) {
            $query['status'] = $status;
        }
        $query['limit'] = $perPage;
        $query['offset'] = $offset;
        $endpoint = 'api/tasks' . (empty($query) ? '' : ('?' . http_build_query($query)));

        $response = $this->mfrelance->doRequest($endpoint, $jwt);
        $tasks = [];

        if (200 === $response['httpCode']) {
            $data = json_decode($response['response'], true);
            $tasks = $data['tasks'] ?? [];
        }

        $hasNext = count($tasks) >= $perPage;

        return $this->render('task/index.html.twig', [
            'tasks' => $tasks,
            'currentStatus' => $status,
            'page' => $page,
            'perPage' => $perPage,
            'hasNext' => $hasNext,
        ]);
    }

    #[Route('/tasks/create', name: 'task_create')]
    public function create(Request $request): Response
    {
        $jwt = $request->getSession()->get('jwt');
        if (!$jwt) {
            return $this->redirectToRoute('app_login');
        }

        if ($request->isMethod('POST')) {
            $deadline = $request->request->get('deadline');
            if (!preg_match('/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/', (string) $deadline)) {
                $this->addFlash('error', 'Неверный формат даты. Используйте YYYY-MM-DDTHH:MM');
                return $this->redirectToRoute('task_create');
            }

            $budget = (float) $request->request->get('budget');
            if ($budget <= 0) {
                $this->addFlash('error', 'Бюджет должен быть больше нуля');
                return $this->redirectToRoute('task_create');
            }

            $taskData = [
                'title' => $request->request->get('title'),
                'description' => $request->request->get('description'),
                'category' => $request->request->get('category'),
                'budget' => $budget,
                'currency' => $request->request->get('currency', 'BTC'),
                'deadline' => $deadline,
            ];

            $response = $this->mfrelance->doRequest('api/tasks/create', $jwt, $taskData, true);

            if (200 === $response['httpCode']) {
                $this->addFlash('success', 'Задача успешно создана!');
                return $this->redirectToRoute('tasks');
            } else {
                $this->addFlash('error', 'Ошибка создания задачи: '.($response['response'] ?? ''));
            }
        }

        return $this->render('task/create.html.twig');
    }

    #[Route('/tasks/{id}', name: 'task_show')]
    public function show(int $id, Request $request): Response
    {
        $jwt = $request->getSession()->get('jwt');
        if (!$jwt) {
            return $this->redirectToRoute('app_login');
        }

        $response = $this->mfrelance->doRequest("api/tasks/get?id={$id}", $jwt);
        $task = null;
        $isAdmin = false;

        if (200 === $response['httpCode']) {
            $data = json_decode($response['response'], true);
            $responseID = $this->mfrelance->doRequest('api/ownID', $jwt);
            $dataID = json_decode($responseID['response'], true);
            $task = $data['task'] ?? null;
            $task['is_own'] = ($dataID['user_id'] ?? 0) == ($data['task']['client_id'] ?? -1);
            $adminResp = $this->mfrelance->doRequest('api/admin/IIsAdmin', $jwt);
            if (200 === $adminResp['httpCode']) {
                $isAdmin = (bool) (json_decode($adminResp['response'], true)['is_admin'] ?? false);
            }
        }
        if (!$task) {
            throw $this->createNotFoundException('Задача не найдена');
        }

        $response = $this->mfrelance->doRequest("api/offers?task_id={$id}", $jwt);
        $offers = [];
        $myOffer = null;
        $usernameCache = [];
        $getUsernameByID = function ($userId) use ($jwt, &$usernameCache) {
            $uid = intval($userId);
            if ($uid <= 0) { return 'Unknown'; }
            if (isset($usernameCache[$uid])) { return $usernameCache[$uid]; }
            $resp = $this->mfrelance->doRequest("profile/by_id?user_id=$uid", $jwt, [], false);
            if (200 === $resp['httpCode']) {
                $data = json_decode($resp['response'], true);
                $name = $data['username'] ?? "User $uid";
                $usernameCache[$uid] = $name;
                return $name;
            }
            return "$uid";
        };

        if (200 === $response['httpCode']) {
            $data = json_decode($response['response'], true);
            $offers = $data['offers'] ?? [];
            foreach ($offers as &$el) {
                $el['freelancer'] = $getUsernameByID($el['freelancer_id']);
            }
            unset($el);
            $ownResp = $this->mfrelance->doRequest('api/ownID', $jwt);
            if (200 === $ownResp['httpCode']) {
                $uid = (int) (json_decode($ownResp['response'], true)['user_id'] ?? 0);
                foreach ($offers as $el) {
                    if ((int)$el['freelancer_id'] === $uid) { $myOffer = $el; break; }
                }
            }
        }

        $response = $this->mfrelance->doRequest("api/reviews/task?task_id={$id}", $jwt);
        $reviews = [];
        if (200 === $response['httpCode']) {
            $data = json_decode($response['response'], true);
            $reviews = $data['reviews'] ?? [];
            foreach ($reviews as &$review) {
                $review['reviewer_name'] = $getUsernameByID($review['reviewer_id']);
            }
            unset($review);
        }

        $canReview = false;
        if ($task['status'] === 'completed') {
            $responseID = $this->mfrelance->doRequest('api/ownID', $jwt);
            if (200 === $responseID['httpCode']) {
                $userId = (int) (json_decode($responseID['response'], true)['user_id'] ?? 0);
                if ($task['client_id'] == $userId) {
                    $canReview = true;
                } else {
                    foreach ($offers as $offer) {
                        if ($offer['accepted'] && $offer['freelancer_id'] == $userId) {
                            $canReview = true;
                            break;
                        }
                    }
                }
            }
        }

        if ($request->isMethod('POST')) {
            $action = $request->request->get('action');
            switch ($action) {
                case 'create_offer':
                    $offerData = [
                        'task_id' => $id,
                        'price' => (float) $request->request->get('price'),
                        'message' => $request->request->get('message'),
                    ];
                    $response = $this->mfrelance->doRequest('api/offers/create', $jwt, $offerData, true);
                    $this->addFlash(200 === $response['httpCode'] ? 'success' : 'error', 200 === $response['httpCode'] ? 'Предложение отправлено!' : ('Ошибка отправки предложения: '.($response['response'] ?? '')));
                    break;

                case 'update_offer':
                    $offerData = [
                        'id' => (int) $request->request->get('offer_id'),
                        'price' => (float) $request->request->get('price'),
                        'message' => $request->request->get('message'),
                    ];
                    $resp = $this->mfrelance->doRequest('api/offers/update', $jwt, $offerData, true);
                    $this->addFlash(200 === $resp['httpCode'] ? 'success' : 'error', 200 === $resp['httpCode'] ? 'Предложение обновлено!' : ('Ошибка обновления предложения: '.($resp['response'] ?? '')));
                    break;

                case 'delete_offer':
                    $offerId = (int) $request->request->get('offer_id');
                    $resp = $this->mfrelance->doRequest('api/offers/delete?id='.$offerId, $jwt, [], true);
                    $this->addFlash(200 === $resp['httpCode'] ? 'success' : 'error', 200 === $resp['httpCode'] ? 'Предложение удалено!' : ('Ошибка удаления предложения: '.($resp['response'] ?? '')));
                    break;

                case 'accept_offer':
                    $offerId = $request->request->get('offer_id');
                    $offerData = ['offer_id' => (int) $offerId];
                    $response = $this->mfrelance->doRequest('api/offers/accept', $jwt, $offerData, true);
                    $this->addFlash(200 === $response['httpCode'] ? 'success' : 'error', 200 === $response['httpCode'] ? 'Предложение принято!' : ('Ошибка принятия предложения: '.($response['response'] ?? '')));
                    break;

                case 'complete_task':
                    $taskData = ['task_id' => $id];
                    $response = $this->mfrelance->doRequest('api/tasks/complete', $jwt, $taskData, true);
                    $this->addFlash(200 === $response['httpCode'] ? 'success' : 'error', 200 === $response['httpCode'] ? 'Задача отмечена как выполненная!' : ('Ошибка завершения задачи: '.($response['response'] ?? '')));
                    break;

                case 'delete_task':
                    $resp = $this->mfrelance->doRequest('api/tasks/delete?id='.$id, $jwt, [], true);
                    $this->addFlash(200 === $resp['httpCode'] ? 'success' : 'error', 200 === $resp['httpCode'] ? 'Задача удалена' : ('Ошибка удаления задачи: '.($resp['response'] ?? '')));
                    if (200 === $resp['httpCode']) { return $this->redirectToRoute('tasks'); }
                    break;

                case 'admin_delete_user_tasks':
                    $uid = (int) ($task['client_id'] ?? 0);
                    if ($uid > 0) {
                        $resp = $this->mfrelance->doRequest('api/admin/delete_user_tasks?user_id='.$uid, $jwt, [], true);
                        $this->addFlash(200 === $resp['httpCode'] ? 'success' : 'error', 200 === $resp['httpCode'] ? 'Все задачи пользователя удалены' : ('Ошибка удаления задач пользователя: '.($resp['response'] ?? '')));
                        if (200 === $resp['httpCode']) { return $this->redirectToRoute('tasks'); }
                    }
                    break;
            }
            return $this->redirectToRoute('task_show', ['id' => $id]);
        }

        return $this->render('task/show.html.twig', [
            'task' => $task,
            'offers' => $offers,
            'myOffer' => $myOffer,
            'isAdmin' => $isAdmin,
            'reviews' => $reviews,
            'canReview' => $canReview,
        ]);
    }
}
