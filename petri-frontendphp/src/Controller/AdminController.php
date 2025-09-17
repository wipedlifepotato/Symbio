<?php

namespace App\Controller;

use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpFoundation\Session\SessionInterface;
use Symfony\Component\Routing\Annotation\Route;
use App\Service\MFrelance;

class AdminController extends AbstractController
{
    private function isBase64Image(string $data): false|string
    {
        $decoded = base64_decode($data, true);
        if (!$decoded) return false;

        if (substr($decoded, 0, 8) === "\x89PNG\x0D\x0A\x1A\x0A") return 'png';
        if (substr($decoded, 0, 3) === "\xFF\xD8\xFF") return 'jpeg';
        if (substr($decoded, 0, 6) === "GIF87a" || substr($decoded, 0, 6) === "GIF89a") return 'gif';

        return false;
    }

    private function usernameByID(MFrelance $mf, string $jwt, int $userId, array &$cache): string
    {
        if ($userId <= 0) return 'Unknown';
        if (isset($cache[$userId])) return $cache[$userId];

        $resp = $mf->doRequest("profile/by_id?user_id=$userId", $jwt, [], false);
        if ($resp['httpCode'] === 200) {
            $data = json_decode($resp['response'], true);
            $name = $data['username'] ?? '';
            if ($name === '') $name = 'User ' . $userId;
            $cache[$userId] = $name;
            return $name;
        }

        return 'User ' . $userId;
    }
    #[Route('/admin', name: 'app_admin')]
    public function admin(Request $request, MFrelance $mf, SessionInterface $session): Response
    {
        $message = '';
        $wallets = [];
        $transactions = [];
        $randomTicket = null;

        $jwt = $session->get('jwt', '');

        if (!$jwt) {
            $message = "JWT отсутствует. Пожалуйста, войдите в систему.";
        } else {
            // Block / Unblock users
            if ($blockUserId = $request->request->getInt('block_user_id')) {
                $res = $mf->doRequest('api/admin/block', $jwt, ['user_id' => $blockUserId], true);
                $message = $res['httpCode'] === 200 ? "Пользователь заблокирован" : "Ошибка: ".$res['response'];
            }
            if ($unblockUserId = $request->request->getInt('unblock_user_id')) {
                $res = $mf->doRequest('api/admin/unblock', $jwt, ['user_id' => $unblockUserId], true);
                $message = $res['httpCode'] === 200 ? "Пользователь разблокирован" : "Ошибка: ".$res['response'];
            }

            // Get wallets
            if ($walletsUserId = $request->query->getInt('wallets_user_id')) {
                $res = $mf->doRequest('api/admin/wallets?user_id=' . $walletsUserId, $jwt);
                if ($res['httpCode'] === 200) {
                    $wallets = json_decode($res['response'], true);
                } else {
                    $message = "Ошибка при получении кошельков: " . $res['response'];
                }
            }

            // Update wallet balance
            if ($updateWalletId = $request->request->getInt('update_wallet_id')) {
                $newBalance = $request->request->get('new_balance');
                $res = $mf->doRequest('api/admin/update_balance', $jwt, [
                    'user_id' => $updateWalletId,
                    'balance' => $newBalance
                ], true);
                $message = $res['httpCode'] === 200 ? "Баланс обновлён" : "Ошибка: ".$res['response'];
            }

            // Transactions
            $txWalletId = $request->query->getInt('tx_wallet_id', 0);
            $txLimit = $request->query->getInt('tx_limit', 20);
            $txOffset = $request->query->getInt('tx_offset', 0);
            $res = $mf->doRequest('api/admin/transactions', $jwt, [
                'wallet_id' => $txWalletId,
                'limit' => $txLimit,
                'offset' => $txOffset
            ], true);
            if ($res['httpCode'] === 200) {
                $transactions = json_decode($res['response'], true);
            } else {
                $message = "Ошибка при получении транзакций: " . $res['response'];
            }

            // Random ticket
            if ($request->query->get('get_ticket')) {
                $res = $mf->doRequest('api/admin/getRandomTicket', $jwt);
                if ($res['httpCode'] === 200) {
                    $randomTicket = json_decode($res['response'], true);
                } else {
                    $message = "Ошибка получения тикета: " . $res['response'];
                }
            }

            // Add user to chat
            if ($request->request->get('add_to_chat')) {
                $chatID = $request->request->getInt('chat_id');
                $userID = $request->request->getInt('user_id');
                if ($chatID && $userID) {
                    $res = $mf->doRequest('api/admin/addUserToChatRoom?chat_id=' . $chatID . '&user_id=' . $userID, $jwt, [], true);
                    $message = $res['httpCode'] === 200 ? "Пользователь добавлен в чат" : "Ошибка: " . $res['response'];
                } else {
                    $message = "Неверные ID";
                }
            }
        }

        return $this->render('admin/admin.html.twig', [
            'message' => $message,
            'wallets' => $wallets,
            'transactions' => $transactions,
            'randomTicket' => $randomTicket,
            'txWalletId' => $txWalletId,
            'txLimit' => $txLimit,
            'txOffset' => $txOffset
        ]);
    }
    #[Route('/admin/tickets', name: 'app_admin_tickets')]
    public function tickets(Request $request, MFrelance $mf, SessionInterface $session): Response
    {
        $jwt = $session->get('jwt', '');
        $message = '';
        $allTickets = [];
        $ticketMessages = [];
        $selectedTicket = null;
        $randomTicket = null;
        $usernameCache = [];

        if (!$jwt) {
            $message = "JWT отсутствует. Пожалуйста, войдите в систему.";
        } else {
            // Получаем все тикеты
            $resTickets = $mf->doRequest("api/ticket/my", $jwt, [], false);
            if ($resTickets['httpCode'] === 200) {
                $allTickets = json_decode($resTickets['response'], true);
		foreach ($allTickets as &$t) {
		    $t['username'] = $this->usernameByID($mf, $jwt, $t['user_id'] ?? 0, $usernameCache);
		}
		unset($t);
            }

            // Случайный тикет
            if ($request->query->get('get_ticket')) {
                $res = $mf->doRequest('api/admin/getRandomTicket', $jwt);
                if ($res['httpCode'] === 200) {
                    $randomTicket = json_decode($res['response'], true);
                    if ($randomTicket) {
			    $randomTicket['username'] = $this->usernameByID($mf, $jwt, $randomTicket['UserID'] ?? 0, $usernameCache);
		    }
                } else {
                    $message = "Ошибка получения тикета: " . $res['response'];
                }
            }

	// Просмотр выбранного тикета
	if ($ticketId = $request->query->getInt('ticket_id')) {
	    $resp = $mf->doRequest("api/ticket/messages?ticket_id=$ticketId", $jwt, [], false);
	    if ($resp['httpCode'] === 200) {
		$ticketMessages = json_decode($resp['response'], true);

		foreach ($ticketMessages as &$m) {
		    $senderId = intval($m['SenderID'] ?? 0);

		    // Получаем имя отправителя из кеша или через функцию
		    $m['SenderName'] = $usernameCache[$senderId] ?? $this->usernameByID($mf, $jwt, $senderId ?? 0, $usernameCache);
		    $usernameCache[$senderId] = $m['SenderName']; // сохраняем в кеш

		    // Проверяем, является ли сообщение изображением
		    $type = $this->isBase64Image($m['Message'] ?? '');
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

		$selectedTicket = $ticketId; // выбираем тикет
	    } else {
		$message = "Ошибка загрузки сообщений: " . $resp['response'];
	    }
	}


            // Отправка сообщения
            if ($request->isMethod('POST') && $request->request->get('send_message') && $ticketId = $request->request->getInt('ticket_id')) {
                $msg = trim($request->request->get('message', ''));

                // Файл
                $file = $request->files->get('file');
                if ($file) {
                    $fileData = file_get_contents($file->getPathname());
                    $msg = base64_encode($fileData);
                }

                $response = $mf->doRequest("api/ticket/write", $jwt, [
                    'ticket_id' => $ticketId,
                    'message' => $msg
                ], true);

                if ($response['httpCode'] === 200) {
                    $resp = $mf->doRequest("api/ticket/messages?ticket_id=$ticketId", $jwt, [], false);
                    if ($resp['httpCode'] === 200) {
                        $ticketMessages = json_decode($resp['response'], true);
                        $selectedTicket = $ticketId;
                    }
                } else {
                    $message = "Ошибка отправки сообщения: " . $response['response'];
                }
            }

            // Закрытие тикета
            if ($request->isMethod('POST') && $request->request->get('close_ticket') && $ticketId = $request->request->getInt('ticket_id')) {
                $response = $mf->doRequest("api/ticket/close", $jwt, ['ticket_id' => $ticketId], true);
                if ($response['httpCode'] === 200) {
                    $message = "Тикет успешно закрыт.";
                    $selectedTicket = null;
                    $ticketMessages = [];
                } else {
                    $message = "Ошибка закрытия тикета: " . $response['response'];
                }
            }
        }

        return $this->render('admin/tickets.html.twig', [
            'message' => $message,
            'allTickets' => $allTickets,
            'ticketMessages' => $ticketMessages,
            'selectedTicket' => $selectedTicket,
            'randomTicket' => $randomTicket,
            'usernameCache' => $usernameCache,
            'jwt' => $jwt,
        ]);
    }
}

