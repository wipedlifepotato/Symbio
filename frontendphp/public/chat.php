<?php
session_start();
require_once __DIR__ . '/../src/mfrelance.php';

$mf = new MFrelance('localhost', 9999);
$jwt = $_SESSION['jwt'] ?? '';
if (!$jwt) die("Пожалуйста, войдите в систему.");

$error = '';
$message = '';
$userCache = []; // Кэш для usernameByID

// ======================
// Отмена входящего запроса на чат
// ======================
if (isset($_GET['cancel_request'])) {
    $requesterID = (int)$_GET['cancel_request'];
    $res = $mf->doRequest("api/chat/cancelChatRequest?requester_id=$requesterID", $jwt, [], true);
    if ($res['httpCode'] === 200) {
        header('Location: chat.php');
        exit;
    } else {
        $error = "Ошибка при отмене запроса: " . $res['response'];
    }
}

// Функция для получения имени пользователя по ID
function usernameByID($mf, $jwt, $userId, &$cache) {
    $uid = intval($userId);
    if ($uid <= 0) return 'Unknown';
    if (isset($cache[$uid])) return $cache[$uid];
    $resp = $mf->doRequest("profile/by_id?user_id=$uid", $jwt, [], false);
    if ($resp['httpCode'] === 200) {
        $data = json_decode($resp['response'], true);
        $name = $data['username'] ?? '';
        if ($name === '') $name = 'User '.$uid;
        $cache[$uid] = $name;
        return $name;
    }
    return 'User '.$uid;
}

// Обработка выхода из чата
if (isset($_GET['exit_chat_id'])) {
    $exitChatID = (int)$_GET['exit_chat_id'];
    $res = $mf->doRequest('api/chat/exitFromChat?chat_room_id=' . $exitChatID, $jwt, [], true);
    if ($res['httpCode'] === 200) {
        header('Location: chat.php');
        exit;
    } else {
        $error = "Ошибка при выходе из чата: " . $res['response'];
    }
}

// ======================
// 1. Обработка POST-запросов
// ======================
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    if (isset($_POST['requested_id'])) {
        $requestedID = (int)$_POST['requested_id'];
        $res = $mf->doRequest("api/chat/createChatRequest?requested_id=$requestedID", $jwt, [], true);
        if ($res['httpCode'] === 201) {
            $message = "Запрос отправлен пользователю #$requestedID";
        } else {
            $error = "Ошибка отправки запроса: " . $res['response'];
        }
    }
    if (isset($_POST['chat_room_id'], $_POST['message'])) {
        $postData = ['message' => $_POST['message']];
        $res = $mf->doRequest('api/chat/sendMessage?chat_room_id=' . (int)$_POST['chat_room_id'], $jwt, $postData, true);
        if ($res['httpCode'] === 201) {
            header('Location: chat.php?chat_id=' . (int)$_POST['chat_room_id']);
            exit;
        } else {
            $error = "Ошибка отправки сообщения: " . $res['response'];
        }
    }
}

// ======================
// 2. Обработка GET-запросов
// ======================
if (isset($_GET['accept_request'])) {
    $requesterID = (int)$_GET['accept_request'];
    $res = $mf->doRequest("api/chat/acceptChatRequest?requester_id=$requesterID", $jwt, [], true);
    if ($res['httpCode'] === 200) {
        header('Location: chat.php');
        exit;
    } else {
        $error = "Ошибка при принятии запроса: " . $res['response'];
    }
}

// ======================
// 3. Получаем данные для отображения
// ======================
$chatRequests = [];
$res = $mf->doRequest("api/chat/getChatRequests", $jwt);
if ($res['httpCode'] === 200) {
    $chatRequests = json_decode($res['response'], true);
}

$chatRooms = [];
$res = $mf->doRequest('api/chat/getChatRoomsForUser', $jwt);
if ($res['httpCode'] === 200) {
    $chatRooms = json_decode($res['response'], true);
}

$selectedChatID = $_GET['chat_id'] ?? null;
$messages = [];
if ($selectedChatID) {
    $res = $mf->doRequest('api/chat/getChatMessages?chat_room_id=' . (int)$selectedChatID, $jwt);
    if ($res['httpCode'] === 200) {
        $messages = json_decode($res['response'], true);
    }
}
?>

<h2>Входящие запросы на чат</h2>
<?php if (!empty($chatRequests)): ?>
    <ul>
        <?php foreach ($chatRequests as $req): ?>
            <li>
                От пользователя <?= htmlspecialchars(usernameByID($mf, $jwt, $req['requested_id'], $userCache)) ?> —
                Статус: <?= htmlspecialchars($req['status']) ?>
                
                <?php if ($req['status'] === 'pending'): ?>
                    <!-- Кнопка принять -->
                    <a href="?accept_request=<?= (int)$req['requested_id'] ?>">Принять</a>

                    <!-- Кнопка отменить -->
                    <a href="?cancel_request=<?= (int)$req['requested_id'] ?>" style="color:red; margin-left:10px;">Отменить</a>
                <?php endif; ?>
            </li>
        <?php endforeach; ?>
    </ul>
<?php else: ?>
    <p>Входящих запросов нет.</p>
<?php endif; ?>

<h2>Отправить запрос на чат</h2>
<form method="POST">
    <input type="number" name="requested_id" placeholder="ID пользователя" required>
    <button type="submit">Отправить запрос</button>
</form>

<h2>Ваши чаты</h2>
<ul>
    <?php foreach ($chatRooms as $chat): ?>
        <li>
            <a href="?chat_id=<?= htmlspecialchars($chat['id']) ?>">
                Chat #<?= htmlspecialchars($chat['id']) ?>
            </a>
        </li>
    <?php endforeach; ?>
</ul>

<?php if ($selectedChatID): ?>
<h3>Сообщения в чате #<?= htmlspecialchars($selectedChatID) ?></h3>
<div style="border:1px solid #ccc; padding:10px; max-height:400px; overflow-y:scroll;">
    <?php foreach ($messages as $msg): ?>
        <p>
            <strong><?= htmlspecialchars(usernameByID($mf, $jwt, $msg['sender_id'], $userCache)) ?>:</strong>
            <?= htmlspecialchars($msg['message']) ?>
            <em style="color:gray; font-size:0.8em;"><?= htmlspecialchars($msg['created_at']) ?></em>
        </p>
    <?php endforeach; ?>
</div>

<h4>Отправить сообщение</h4>
<form method="POST">
    <input type="hidden" name="chat_room_id" value="<?= htmlspecialchars($selectedChatID) ?>">
    <textarea name="message" rows="3" style="width:100%" required></textarea>
    <button type="submit">Отправить</button>
</form>

<form method="GET" style="margin-top:10px;">
    <input type="hidden" name="exit_chat_id" value="<?= htmlspecialchars($selectedChatID) ?>">
    <button type="submit" style="background-color:red; color:white;">Выйти из чата</button>
</form>
<?php endif; ?>
