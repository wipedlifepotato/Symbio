<?php
session_start();
require_once __DIR__ . '/../src/mfrelance.php';

function isBase64Image($data) {
    $decoded = base64_decode($data, true);
    if (!$decoded) return false;

    if (substr($decoded, 0, 8) === "\x89PNG\x0D\x0A\x1A\x0A") return 'png';
    if (substr($decoded, 0, 3) === "\xFF\xD8\xFF") return 'jpeg';
    if (substr($decoded, 0, 6) === "GIF87a" || substr($decoded, 0, 6) === "GIF89a") return 'gif';

    return false;
}

$message = '';
$selectedTicket = null;
$ticketMessages = [];
$myTickets = [];

$mf = new MFrelance('localhost', 9999);
$jwt = $_SESSION['jwt'] ?? '';
$usernameCache = [];

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

if (!$jwt) {
    $message = "JWT отсутствует. Пожалуйста, войдите в систему.";
} else {
    if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['create_ticket'])) {
        $subject = trim($_POST['subject']);
        $msg = trim($_POST['message']);

        if (!empty($_FILES['file']['tmp_name'])) {
            $fileData = file_get_contents($_FILES['file']['tmp_name']);
            $msg = base64_encode($fileData);
        }

        $response = $mf->doRequest("api/ticket/createTicket", $jwt, [
            'subject' => $subject,
            'message' => $msg
        ], true);

        if ($response['httpCode'] === 200) {
            $message = "Тикет создан!";
        } else {
            $message = "Ошибка создания тикета: " . $response['response'];
        }
    }

    if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['send_message'], $_POST['ticket_id'])) {
        $ticketId = intval($_POST['ticket_id']);
        $msg = trim($_POST['message']);

        if (!empty($_FILES['file']['tmp_name'])) {
            $fileData = file_get_contents($_FILES['file']['tmp_name']);
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

    if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['close_ticket'], $_POST['ticket_id'])) {
        $ticketId = intval($_POST['ticket_id']);

        $response = $mf->doRequest("api/ticket/close", $jwt, [
            'ticket_id' => $ticketId
        ], true);

        if ($response['httpCode'] === 200) {
            $message = "Тикет успешно закрыт.";
            $selectedTicket = null;
            $ticketMessages = [];
        } else {
            $message = "Ошибка закрытия тикета: " . $response['response'];
        }
    }

    $respTickets = $mf->doRequest("api/ticket/my", $jwt, [], false);
    if ($respTickets['httpCode'] === 200) {
        $myTickets = json_decode($respTickets['response'], true);
    }

    if (isset($_GET['ticket_id'])) {
        $ticketId = intval($_GET['ticket_id']);
        $resp = $mf->doRequest("api/ticket/messages?ticket_id=$ticketId", $jwt, [], false);
        if ($resp['httpCode'] === 200) {
            $ticketMessages = json_decode($resp['response'], true);
            $selectedTicket = $ticketId;
        } else {
            $message = "Ошибка загрузки сообщений: " . $resp['response'];
        }
    }
}
?>

<h2>Мои тикеты</h2>
<?php if ($message) echo "<p style='color:red;'>$message</p>"; ?>

<h3>Список тикетов</h3>
<ul>
<?php foreach ($myTickets as $t): ?>
    <li>
        <a href="?ticket_id=<?= $t['id'] ?>">
            <?= htmlspecialchars($t['subject']) ?> (<?= $t['status'] ?>)
        </a>
    </li>
<?php endforeach; ?>
</ul>

<h3>Создать новый тикет</h3>
<form method="POST" enctype="multipart/form-data">
    <input type="hidden" name="create_ticket" value="1">
    <label>Тема: <input type="text" name="subject" required></label><br>
    <label>Сообщение: <textarea name="message"></textarea></label><br>
    <label>Или файл: <input type="file" name="file"></label><br>
    <button type="submit">Создать тикет</button>
</form>

<?php if ($selectedTicket): ?>
    <h3>Чат тикета ID <?= $selectedTicket ?></h3>
    <div style="border:1px solid #ccc; padding:10px; max-height:300px; overflow-y:scroll;">
        <?php foreach ($ticketMessages as $m): ?>
            <p>
                <?php $sender = isset($m['SenderID']) ? intval($m['SenderID']) : 0; ?>
                <strong><?= htmlspecialchars(usernameByID($mf, $jwt, $sender, $usernameCache)) ?>:</strong><br>
                <?php
                $type = isBase64Image($m['Message']);
                if ($type) {
                    $img = 'data:image/' . $type . ';base64,' . $m['Message'];
                    echo "<img src='$img' style='max-width:200px;'><br>";
                } else {
                    echo nl2br(htmlspecialchars($m['Message']));
                }
                ?>
            </p>
        <?php endforeach; ?>
    </div>

    <h4>Отправить сообщение</h4>
    <form method="POST" enctype="multipart/form-data">
        <input type="hidden" name="ticket_id" value="<?= $selectedTicket ?>">
        <input type="hidden" name="send_message" value="1">
        <textarea name="message" required></textarea><br>
        <label>Или файл: <input type="file" name="file"></label><br>
        <button type="submit">Отправить</button>
    </form>

    <h4>Закрыть тикет</h4>
    <form method="POST">
        <input type="hidden" name="ticket_id" value="<?= $selectedTicket ?>">
        <input type="hidden" name="close_ticket" value="1">
        <button type="submit" style="background:red; color:white;">Закрыть тикет</button>
    </form>
<?php endif; ?>

<p><a href="dashboard.php">Dashboard</a></p>
<p><a href="profiles.php">Профили</a></p>
