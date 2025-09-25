<?php

// src/Controller/DashboardController.php

namespace App\Controller;

use App\Service\MFrelance;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpFoundation\Session\SessionInterface;
use Symfony\Component\Routing\Annotation\Route;
use Symfony\Contracts\Translation\TranslatorInterface;

function isBase64Image($data)
{
    $decoded = base64_decode($data, true);
    if (!$decoded) {
        return false;
    }

    if ("\x89PNG\x0D\x0A\x1A\x0A" === substr($decoded, 0, 8)) {
        return 'png';
    }
    if ("\xFF\xD8\xFF" === substr($decoded, 0, 3)) {
        return 'jpeg';
    }
    if ('GIF87a' === substr($decoded, 0, 6) || 'GIF89a' === substr($decoded, 0, 6)) {
        return 'gif';
    }

    return false;
}

class DashboardController extends AbstractController
{
    #[Route('/dashboard', name: 'app_dashboard')]
    public function dashboard(MFrelance $mfrelance, Request $request, SessionInterface $session, TranslatorInterface $translator): Response
    {
        $message = '';
        $walletInfo = null;
        $sendResult = null;
        $isAdmin = false;
        $jwt = $session->get('jwt', '');

        if (!$jwt) {
            $message = $translator->trans('auth.jwt_missing');
        } else {
            try {
                $isAdminResponse = $mfrelance->doRequest('api/admin/IIsAdmin', $jwt);
                // var_dump($isAdminResponse);
                if (200 === $isAdminResponse['httpCode']) {
                    $isAdminResponse = json_decode($isAdminResponse['response'], true);
                    $isAdmin = $isAdminResponse['is_admin'];
                }
                $walletResponse = $mfrelance->doRequest('api/wallet?currency=BTC', $jwt);
                if (200 === $walletResponse['httpCode']) {
                    $walletInfo = json_decode($walletResponse['response'], true);
                } else {
                    $message = 'Ошибка при получении кошелька: '.$walletResponse['response'];
                    $session->remove('jwt');
                }
            } catch (\Exception $e) {
                $message = 'Ошибка запроса кошелька: '.$e->getMessage();
                $session->remove('jwt');
            }

            // send BTC
            if ($request->isMethod('POST')) {
                $to = $request->request->get('to');
                $amount = $request->request->get('amount');

                if ($to && $amount) {
                    try {
                        $sendResponse = $mfrelance->doRequest("api/wallet/bitcoinSend?to=$to&amount=$amount", $jwt, null, false);
                        $sendResult = $sendResponse['response'];

                        if (200 !== $sendResponse['httpCode']) {
                            $message = 'Ошибка отправки BTC: '.$sendResult;
                        }
                    } catch (\Exception $e) {
                        $message = 'Ошибка отправки BTC: '.$e->getMessage();
                    }
                }
            }
        }
        $projectName = $this->getParameter('project_name');

        return $this->render('dashboard/index.html.twig', [
            'walletInfo' => $walletInfo,
            'sendResult' => $sendResult,
            'message' => $message,
            'projectName' => $projectName,
            'isAdmin' => $isAdmin,
        ]);
    }

    #[Route('/profiles', name: 'app_profiles')]
    public function profiles(Request $request, MFrelance $mfrelance, SessionInterface $session, TranslatorInterface $translator): Response
    {
        $message = '';
        $profiles = [];
        $total = 0;
        $offset = (int) $request->query->get('offset', 0);
        $limit = (int) $request->query->get('limit', 5);
        $jwt = $session->get('jwt', '');

        if (!$jwt) {
            $message = $translator->trans('auth.jwt_missing');
        } else {

            try {
                $response = $mfrelance->doRequest("profiles?limit={$limit}&offset={$offset}", $jwt);
                if (200 === $response['httpCode']) {
                    $data = json_decode($response['response'], true);
                    $profiles = $data['profiles'] ?? [];
                    $total = $data['total'] ?? 0;
                } else {
                    $message = 'Ошибка при получении профилей: '.$response['response'];
                    $session->remove('jwt');
                }
            } catch (\Exception $e) {
                $message = 'Ошибка запроса: '.$e->getMessage();
            }

            if ($request->isMethod('POST') && $request->request->has('create_chat_request')) {
                $requestedID = (int) $request->request->get('requested_id', 0);
                try {
                    $res = $mfrelance->doRequest("api/chat/createChatRequest?requested_id={$requestedID}", $jwt, [], true);
                    if (201 === $res['httpCode']) {
                        $message = "Запрос на чат отправлен пользователю #$requestedID";
                    } else {
                        $message = 'Ошибка: '.$res['response'];
                    }
                } catch (\Exception $e) {
                    $message = 'Ошибка запроса: '.$e->getMessage();
                }
            }
        }

        $projectName = $this->getParameter('project_name');

        return $this->render('dashboard/profiles.html.twig', [
            'projectName' => $projectName,
            'profiles' => $profiles,
            'message' => $message,
            'userId' => $session->get('user_id', 0),
            'total' => $total ?? 0,
            'offset' => $offset,
            'limit' => $limit,
        ]);
    }

    #[Route('/profile', name: 'app_profile')]
    public function profile(Request $request, MFrelance $mfrelance, SessionInterface $session, TranslatorInterface $translator): Response
    {
        $message = '';
        $profile = null;
        $username = '';
        $jwt = $session->get('jwt', '');

        if (!$jwt) {
            $message = $translator->trans('auth.jwt_missing');
        } else {
            // Получаем текущий профиль
            try {
                $profileResponse = $mfrelance->doRequest('profile', $jwt);
                if (200 === $profileResponse['httpCode']) {
                    $data = json_decode($profileResponse['response'], true);
                    $username = $data['username'] ?? '';
                    $profile = $data['profile'] ?? null;
                } else {
                    $message = 'Ошибка при получении профиля: '.$profileResponse['response'];
                }
            } catch (\Exception $e) {
                $message = 'Ошибка запроса: '.$e->getMessage();
            }

            // Update profile
            if ($request->isMethod('POST')) {
                $fullName = $request->request->get('full_name', '');
                $bio = $request->request->get('bio', '');
                $skills = $request->request->get('skills', '');
                $skillsArray = array_filter(array_map('trim', explode(',', $skills)));

                $avatarBase64 = $request->request->get('avatar_base64', '');

                $uploadedFile = $request->files->get('avatar');
                if ($uploadedFile) {
                    $avatarData = file_get_contents($uploadedFile->getPathname());
                    $avatarBase64 = 'data:'.$uploadedFile->getClientMimeType().';base64,'.base64_encode($avatarData);
                }

                try {
                    $updateResponse = $mfrelance->doRequest(
                        'profile',
                        $jwt,
                        [
                            'full_name' => $fullName,
                            'bio' => $bio,
                            'skills' => $skillsArray,
                            'avatar' => $avatarBase64,
                        ],
                        true
                    );

                    if (200 === $updateResponse['httpCode']) {
                        $message = 'Профиль обновлен успешно!';
                        $profile['full_name'] = $fullName;
                        $profile['bio'] = $bio;
                        $profile['skills'] = $skillsArray;
                        $profile['avatar'] = $avatarBase64;
                    } else {
                        $message = 'Ошибка при обновлении профиля: '.$updateResponse['response'];
                    }
                } catch (\Exception $e) {
                    $message = 'Ошибка запроса: '.$e->getMessage();
                }
            }
        }

        $projectName = $this->getParameter('project_name');

        return $this->render('dashboard/profile.html.twig', [
            'projectName' => $projectName,
            'profile' => $profile,
            'username' => $username,
            'message' => $message,
        ]);
    }

    #[Route('/ticket', name: 'app_ticket')]
    public function ticket(Request $request, MFrelance $mfrelance, SessionInterface $session, TranslatorInterface $translator): Response
    {
        $message = '';
        $selectedTicket = null;
        $ticketMessages = [];
        $myTickets = [];
        $usernameCache = [];

        $jwt = $session->get('jwt', '');
        if (!$jwt) {
            $message = $translator->trans('auth.jwt_missing');
        } else {
            $getUsernameByID = function ($userId) use ($mfrelance, $jwt, &$usernameCache) {
                $uid = intval($userId);
                if ($uid <= 0) {
                    return 'Unknown';
                }
                if (isset($usernameCache[$uid])) {
                    return $usernameCache[$uid];
                }
                $resp = $mfrelance->doRequest("profile/by_id?user_id=$uid", $jwt, [], false);
                if (200 === $resp['httpCode']) {
                    $data = json_decode($resp['response'], true);
                    $name = $data['username'] ?? "User $uid";
                    $usernameCache[$uid] = $name;

                    return $name;
                }

                return "User $uid";
            };

            // Create a new ticket
            if ($request->isMethod('POST') && $request->request->get('create_ticket')) {
                $subject = $request->request->get('subject', '');
                $msg = $request->request->get('message', '');
                $uploadedFile = $request->files->get('file');
                if ($uploadedFile) {
                    $msg = base64_encode(file_get_contents($uploadedFile->getPathname()));
                }
                $resp = $mfrelance->doRequest('api/ticket/createTicket', $jwt, [
                    'subject' => $subject,
                    'message' => $msg,
                ], true);
                $message = (200 === $resp['httpCode']) ? 'Тикет создан!' : 'Ошибка создания тикета: '.$resp['response'];
            }

            // Send message to the ticket
            if ($request->isMethod('POST') && $request->request->get('send_message')) {
                $ticketId = intval($request->request->get('ticket_id'));
                $msg = $request->request->get('message', '');
                $uploadedFile = $request->files->get('file');
                if ($uploadedFile) {
                    $msg = base64_encode(file_get_contents($uploadedFile->getPathname()));
                }
                $resp = $mfrelance->doRequest('api/ticket/write', $jwt, [
                    'ticket_id' => $ticketId,
                    'message' => $msg,
                ], true);

                if (200 === $resp['httpCode']) {
                    $respMsg = $mfrelance->doRequest("api/ticket/messages?ticket_id=$ticketId", $jwt, [], false);
                    if (200 === $respMsg['httpCode']) {
                        $ticketMessages = json_decode($respMsg['response'], true);
                        $selectedTicket = $ticketId;
                    }
                } else {
                    $message = 'Ошибка отправки сообщения: '.$resp['response'];
                }
            }

            // close a ticket
            if ($request->isMethod('POST') && $request->request->get('close_ticket')) {
                $ticketId = intval($request->request->get('ticket_id'));
                $resp = $mfrelance->doRequest('api/ticket/close', $jwt, ['ticket_id' => $ticketId], true);
                if (200 === $resp['httpCode']) {
                    $message = 'Тикет успешно закрыт.';
                    $selectedTicket = null;
                    $ticketMessages = [];
                } else {
                    $message = 'Ошибка закрытия тикета: '.$resp['response'];
                }
            }

            // Get all ticket by user
            $respTickets = $mfrelance->doRequest('api/ticket/my', $jwt, [], false);
            if (200 === $respTickets['httpCode']) {
                $myTickets = json_decode($respTickets['response'], true);
            }

            // See ticket
            $ticketIdGet = $request->query->getInt('ticket_id', 0);
            if ($ticketIdGet) {
                $respMsg = $mfrelance->doRequest("api/ticket/messages?ticket_id=$ticketIdGet", $jwt, [], false);
                if (200 === $respMsg['httpCode']) {
                    $ticketMessages = json_decode($respMsg['response'], true);
                    foreach ($ticketMessages as &$m) {
                        $senderId = intval($m['SenderID'] ?? 0);
                        $m['SenderName'] = $usernameCache[$senderId] ?? $getUsernameByID($senderId);
                        $type = isBase64Image($m['Message'] ?? '');
                        if ($type) {
                            $m['is_image'] = true;
                            $m['image_type'] = $type;
                            $m['image_data'] = $m['Message'];
                        } else {
                            $m['is_image'] = false;
                            $m['text'] = $m['Message'] ?? '';
                        }
                    }
                    unset($m);
                    $selectedTicket = $ticketIdGet;
                } else {
                    $message = 'Ошибка загрузки сообщений: '.$respMsg['response'];
                }
            }
        }
        $projectName = $this->getParameter('project_name');

        return $this->render('dashboard/ticket.html.twig', [
            'projectName' => $projectName,
            'message' => $message,
            'myTickets' => $myTickets,
            'selectedTicket' => $selectedTicket,
            'ticketMessages' => $ticketMessages,
            'getUsernameByID' => $getUsernameByID,
        ]);
    }

    #[Route('/chat', name: 'app_chat')]
    public function chat(Request $request, MFrelance $mfrelance, SessionInterface $session, TranslatorInterface $translator): Response
    {
        $message = '';
        $error = '';
        $chatRequests = [];
        $chatRooms = [];
        $messages = [];
        $selectedChatID = null;
        $usernameCache = [];

        $jwt = $session->get('jwt', '');
        if (!$jwt) {
            $message = $translator->trans('auth.jwt_missing');
        } else {
            $getUsernameByID = function ($userId) use ($mfrelance, $jwt, &$usernameCache) {
                $uid = intval($userId);
                if ($uid <= 0) {
                    return 'Unknown';
                }
                if (isset($usernameCache[$uid])) {
                    return $usernameCache[$uid];
                }
                $resp = $mfrelance->doRequest("profile/by_id?user_id=$uid", $jwt, [], false);
                if (200 === $resp['httpCode']) {
                    $data = json_decode($resp['response'], true);
                    $name = $data['username'] ?? "User $uid";
                    $usernameCache[$uid] = $name;

                    return $name;
                }

                return "User $uid";
            };

            // Cancel request
            if ($request->query->has('cancel_request')) {
                $reqID = $request->query->getInt('cancel_request');
                $res = $mfrelance->doRequest("api/chat/cancelChatRequest?requester_id=$reqID", $jwt, [], true);
                if (200 === $res['httpCode']) {
                    return $this->redirectToRoute('app_chat');
                } else {
                    $error = 'Ошибка при отмене запроса: '.$res['response'];
                }
            }

            // Accept request
            if ($request->query->has('accept_request')) {
                $reqID = $request->query->getInt('accept_request');
                $res = $mfrelance->doRequest("api/chat/acceptChatRequest?requester_id=$reqID", $jwt, [], true);
                if (200 === $res['httpCode']) {
                    return $this->redirectToRoute('app_chat');
                } else {
                    $error = 'Ошибка при принятии запроса: '.$res['response'];
                }
            }

            // Exit from chat
            if ($request->query->has('exit_chat_id')) {
                $chatID = $request->query->getInt('exit_chat_id');
                $res = $mfrelance->doRequest("api/chat/exitFromChat?chat_room_id=$chatID", $jwt, [], true);
                if (200 === $res['httpCode']) {
                    return $this->redirectToRoute('app_chat');
                } else {
                    $error = 'Ошибка выхода: '.$res['response'];
                }
            }

            // POST → send a new message
            if ($request->isMethod('POST')) {
                if ($request->request->has('requested_id')) {
                    $rid = $request->request->getInt('requested_id');
                    $res = $mfrelance->doRequest("api/chat/createChatRequest?requested_id=$rid", $jwt, [], true);
                    if (201 === $res['httpCode']) {
                        $message = "Запрос отправлен пользователю #$rid";
                    } else {
                        $error = 'Ошибка запроса: '.$res['response'];
                    }
                }
                if ($request->request->has('chat_room_id') && $request->request->has('message')) {
                    $chatID = $request->request->getInt('chat_room_id');
                    $msg = $request->request->get('message');
                    $uploadedFile = $request->files->get('file');
                    if ($uploadedFile) {
                        $msg = base64_encode(file_get_contents($uploadedFile->getPathname()));
                    }
                    $res = $mfrelance->doRequest("api/chat/sendMessage?chat_room_id=$chatID", $jwt, ['message' => $msg], true);
                    if (201 === $res['httpCode']) {
                        return $this->redirectToRoute('app_chat', ['chat_id' => $chatID]);
                    } else {
                        $error = 'Ошибка отправки: '.$res['response'];
                    }
                }
            }

            // enter request
            $res = $mfrelance->doRequest('api/chat/getChatRequests', $jwt);
            if (200 === $res['httpCode']) {
                $chatRequests = json_decode($res['response'], true);
            }

            // my chats
            $res = $mfrelance->doRequest('api/chat/getChatRoomsForUser', $jwt);
            //var_dump($res);
            //exit(0);
            if (200 === $res['httpCode']) {
                $chatRooms = json_decode($res['response'], true);
                if (!is_array($chatRooms)) {
                    $error = 'API вернул некорректный формат данных для чатов';
                    $chatRooms = [];
                }
            } else {
                $error = 'Ошибка загрузки чатов: HTTP ' . $res['httpCode'] . ' - ' . $res['response'];
                $chatRooms = [];
            }

            // messages of chat
            $selectedChatID = $request->query->getInt('chat_id', 0);
            $selectedChatName = null;
            if ($selectedChatID && is_array($chatRooms)) {
                // Find chat name
                foreach ($chatRooms as $chat) {
                    if ($chat['id'] == $selectedChatID) {
                        $selectedChatName = $chat['name'] ?? "Chat #$selectedChatID";
                        break;
                    }
                }

                $res = $mfrelance->doRequest("api/chat/getChatMessages?chat_room_id=$selectedChatID", $jwt);
                if (200 === $res['httpCode']) {
                    $messages = json_decode($res['response'], true);
                    if ($messages) {
                        foreach ($messages as &$m) {
                            $senderId = intval($m['sender_id'] ?? 0);
                            $m['SenderName'] = $usernameCache[$senderId] ?? $getUsernameByID($senderId);
                            $type = isBase64Image($m['message'] ?? '');
                            $m['raw_message'] = $m['message'] ?? '';

                            if ($type) {
                                $m['is_image'] = true;
                                $m['image_type'] = $type;
                                $m['image_data'] = $m['message'];
                            } else {
                                $m['is_image'] = false;
                                $m['text'] = $m['message'] ?? '';
                            }
                        }
                    }
                    unset($m);
                }
            }
        }

        return $this->render('chat/chat.html.twig', [
            'error' => $error,
            'message' => $message,
            'chatRequests' => $chatRequests,
            'chatRooms' => $chatRooms,
            'selectedChatID' => $selectedChatID,
            'selectedChatName' => $selectedChatName,
            'messages' => $messages,
        ]);
    }
}
