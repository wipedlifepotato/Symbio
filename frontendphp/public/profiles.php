<?php
session_start();
require_once __DIR__ . '/../src/mfrelance.php';

$message = '';
$profiles = [];

$mf = new MFrelance('localhost', 9999);
$jwt = $_SESSION['jwt'] ?? '';

if (!$jwt) {
    $message = "JWT отсутствует. Пожалуйста, войдите в систему.";
} else {
    $offset = 0;
    $limit = 50;
    if(isset($_GET['offset']))
    {
    	$offset = $_GET['offset'];
    }
    if(isset($_GET['limit']))
    {
    	$limit = $_GET['limit'];
    }
    $response = $mf->doRequest("profiles?limit=".$limit."&"."offset=".$offset, $jwt);
    if ($response['httpCode'] === 200) {
        $profiles = json_decode($response['response'], true);
    } else {
        $message = "Ошибка при получении профилей: " . $response['response'];
        unset($_SESSION['jwt']);
    }
}
?>

<h2>Все профили пользователей</h2>

<?php if ($message) echo "<p style='color:red;'>$message</p>"; ?>

<?php if ($profiles): ?>
    <table border="1" cellpadding="5">
        <tr>
            <th>UserID</th>
            <th>Full Name</th>
            <th>Bio</th>
            <th>Skills</th>
            <th>Avatar</th>
            <th>Rating</th>
            <th>Completed Tasks</th>
            <th>Actions</th>
        </tr>
        <?php foreach ($profiles as $p): ?>
            <tr>
                <td><?= htmlspecialchars($p['user_id']) ?></td>
                <td><?= htmlspecialchars($p['full_name']) ?></td>
                <td><?= htmlspecialchars($p['bio']) ?></td>
                <td><?= htmlspecialchars(implode(',', $p['skills'] ?? [])) ?></td>
                <td>
                    <?php if (!empty($p['avatar'])): ?>
                        <img src="<?= $p['avatar'] ?>" alt="Avatar" style="max-width:50px;">
                    <?php endif; ?>
                </td>
                <td><?= htmlspecialchars($p['rating']) ?></td>
                <td><?= htmlspecialchars($p['completed_tasks']) ?></td>
                <td>
                    <?php if ($p['user_id'] != ($_SESSION['user_id'] ?? 0)): ?>
                        <form method="POST" style="margin:0;">
                            <input type="hidden" name="requested_id" value="<?= htmlspecialchars($p['user_id']) ?>">
                            <button type="submit" name="create_chat_request">Запросить чат</button>
                        </form>
                    <?php endif; ?>
                </td>
            </tr>
        <?php endforeach; ?>
    </table>
<?php else: ?>
    <p>Профили отсутствуют.</p>
<?php endif; ?>

<p><a href="index.php">Назад</a></p>

<?php
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['create_chat_request'], $_POST['requested_id'])) {
    $requestedID = (int)$_POST['requested_id'];
    $res = $mf->doRequest("api/chat/createChatRequest?requested_id=$requestedID", $jwt, [], true);
    if ($res['httpCode'] === 201) {
        echo "<p style='color:green;'>Запрос на чат отправлен пользователю #$requestedID</p>";
    } else {
        echo "<p style='color:red;'>Ошибка: " . htmlspecialchars($res['response']) . "</p>";
    }
}
?>
