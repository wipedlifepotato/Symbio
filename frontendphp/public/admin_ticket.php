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
$randomTicket = null;

$mf = new MFrelance('localhost', 9999);
$jwt = $_SESSION['jwt'] ?? '';

if (!$jwt) {
    $message = "JWT отсутствует. Пожалуйста, войдите в систему.";
} else {
    $allTickets = [];
    $resTickets = $mf->doRequest("api/ticket/my", $jwt, [], false);
    if ($resTickets['httpCode'] === 200) {
        $allTickets = json_decode($resTickets['response'], true);
    }
    if (isset($_GET['get_ticket'])) {
        $res = $mf->doRequest('api/admin/getRandomTicket', $jwt);
        if ($res['httpCode'] === 200) {
            $randomTicket = json_decode($res['response'], true);
        } else {
            $message = "Ошибка получения тикета: " . $res['response'];
        }
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
?>

<h2>Admin Panel - Tickets</h2>
<?php if ($message) echo "<p style='color:red;'>$message</p>"; ?>
<h3>Все тикеты</h3>
<ul>
<?php foreach ($allTickets as $t): ?>
    <li>
        <a href="?ticket_id=<?= $t['id'] ?>">
            <?= htmlspecialchars($t['subject']) ?> (<?= $t['status'] ?>) — User ID: <?= htmlspecialchars($t['user_id'] ?? '') ?>
        </a>
    </li>
<?php endforeach; ?>
</ul>
<h3>Случайный открытый тикет</h3>
<?php if ($randomTicket): ?>
    <p>
        <strong>ID:</strong> <?= htmlspecialchars($randomTicket['ID'] ?? '') ?><br>
        <strong>User ID:</strong> <?= htmlspecialchars($randomTicket['UserID'] ?? '') ?><br>
        <strong>Subject:</strong> <?= htmlspecialchars($randomTicket['Subject'] ?? '') ?><br>
        <strong>Status:</strong> <?= htmlspecialchars($randomTicket['Status'] ?? '') ?><br>
        <a href="?ticket_id=<?= $randomTicket['ID'] ?>">Открыть для ответа</a>
    </p>
<?php else: ?>
    <form method="GET">
        <button type="submit" name="get_ticket" value="1">Получить случайный тикет</button>
    </form>
<?php endif; ?>

<?php if ($selectedTicket): ?>
    <h3>Чат тикета ID <?= $selectedTicket ?></h3>
    <div style="border:1px solid #ccc; padding:10px; max-height:300px; overflow-y:scroll;">
        <?php foreach ($ticketMessages as $m): ?>
            <p>
                <strong>User <?= htmlspecialchars($m['SenderID'] ?? 'Unknown') ?>:</strong><br>
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
