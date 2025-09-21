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
        $endpoint = 'all' === $status ? 'api/tasks' : "api/tasks?status={$status}";

        $response = $this->mfrelance->doRequest($endpoint, $jwt);
        $tasks = [];

        if (200 === $response['httpCode']) {
            $data = json_decode($response['response'], true);
            $tasks = $data['tasks'] ?? [];
        }

        return $this->render('task/index.html.twig', [
            'tasks' => $tasks,
            'currentStatus' => $status,
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
            $taskData = [
                'title' => $request->request->get('title'),
                'description' => $request->request->get('description'),
                'category' => $request->request->get('category'),
                'budget' => (float) $request->request->get('budget'),
                'currency' => $request->request->get('currency', 'BTC'),
                'deadline' => $request->request->get('deadline'),
            ];

            $response = $this->mfrelance->doRequest('api/tasks/create', $jwt, $taskData, true);

            if (200 === $response['httpCode']) {
                $this->addFlash('success', 'Задача успешно создана!');

                return $this->redirectToRoute('tasks');
            } else {
                $this->addFlash('error', 'Ошибка создания задачи');
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

        if (200 === $response['httpCode']) {
            $data = json_decode($response['response'], true);
            $responseID = $this->mfrelance->doRequest('api/ownID', $jwt);
            $dataID = json_decode($responseID['response'], true);
            $task = $data['task'] ?? null;
            $task['is_own'] = $dataID['user_id'] == $data['task']['client_id'];
            // var_dump($task);
            // exit(0);
        }

        if (!$task) {
            throw $this->createNotFoundException('Задача не найдена');
        }

        $response = $this->mfrelance->doRequest("api/offers?task_id={$id}", $jwt);
        $offers = [];
        $getUsernameByID = function ($userId) use ($jwt, &$usernameCache) {
            $uid = intval($userId);
            if ($uid <= 0) {
                return 'Unknown';
            }
            if (isset($usernameCache[$uid])) {
                return $usernameCache[$uid];
            }

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
            foreach ($data as &$m) {
                if (is_array($m)) {
                    foreach ($m as &$el) {
                        $el['freelancer'] = $getUsernameByID($el['freelancer_id']);
                    }
                }
            }
            unset($el);
            unset($m);
            $offers = $data['offers'] ?? [];
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
                    // die($jwt);
                    $response = $this->mfrelance->doRequest('api/offers/create', $jwt, $offerData, true);
                    if (200 === $response['httpCode']) {
                        $this->addFlash('success', 'Предложение отправлено!');
                    } else {
                        //	var_dump($response);
                        // exit(0);
                        $this->addFlash('error', 'Ошибка отправки предложения');
                    }
                    break;

                case 'accept_offer':
                    $offerId = $request->request->get('offer_id');
                    $offerData = ['offer_id' => (int) $offerId];

                    $response = $this->mfrelance->doRequest('api/offers/accept', $jwt, $offerData, true);
                    if (200 === $response['httpCode']) {
                        $this->addFlash('success', 'Предложение принято!');
                    } else {
                        // var_dump($response);
                        // exit(0);
                        $this->addFlash('error', 'Ошибка принятия предложения');
                    }
                    break;

                case 'complete_task':
                    $taskData = ['task_id' => $id];

                    $response = $this->mfrelance->doRequest('api/tasks/complete', $jwt, $taskData, true);
                    if (200 === $response['httpCode']) {
                        $this->addFlash('success', 'Задача отмечена как выполненная!');
                    } else {
                        // var_dump($response);
                        // exit(0);
                        $this->addFlash('error', 'Ошибка завершения задачи');
                    }
                    break;
            }

            return $this->redirectToRoute('task_show', ['id' => $id]);
        }

        return $this->render('task/show.html.twig', [
            'task' => $task,
            'offers' => $offers,
        ]);
    }
}
