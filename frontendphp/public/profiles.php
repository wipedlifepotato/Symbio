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
        //unset($_SESSION['jwt']);
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
            </tr>
        <?php endforeach; ?>
    </table>
<?php else: ?>
    <p>Профили отсутствуют.</p>
<?php endif; ?>

<p><a href="index.php">Назад</a></p>
