<?php
session_start();
require_once __DIR__ . '/../src/mfrelance.php';

$message = '';
$walletInfo = null;
$sendResult = null;

$mf = new MFrelance('localhost', 9999);

$jwt = $_SESSION['jwt'] ?? '';

if (!$jwt) {
    $message = "JWT отсутствует. Пожалуйста, войдите в систему.";
} else {
    $walletResponse = $mf->doRequest("api/wallet?currency=BTC", $jwt);
    if ($walletResponse['httpCode'] === 200) {
    	//echo "httpCode 200".$walletResponse['response'];
        $walletInfo = json_decode($walletResponse['response'], true);
    } else {
        $message = "Ошибка при получении кошелька: " . $walletResponse['response'];
        unset($_SESSION['jwt']);
    }

    if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['to'], $_POST['amount'])) {
        $to = $_POST['to'];
        $amount = $_POST['amount'];
        $sendResponse = $mf->doRequest("api/wallet/bitcoinSend?to=$to&amount=$amount", $jwt);
        $sendResult = $sendResponse['response'];

        if ($sendResponse['httpCode'] !== 200) {
            $message = "Ошибка отправки BTC: " . $sendResult;
        }
    }
}
?>

<h2>Dashboard пользователя</h2>

<?php if ($message) echo "<p style='color:red;'>$message</p>"; ?>

<?php if ($walletInfo): ?>
    <h3>Кошелек BTC:</h3>
    <p>Адрес: <?= htmlspecialchars($walletInfo['address'] ?? '-') ?></p>
    <?php if (isset($walletInfo['balance'])): ?>
        <p>Баланс: <?= htmlspecialchars($walletInfo['balance']) ?></p>
    <?php endif; ?>

    <h3>Отправить BTC</h3>
    <form method="POST">
        <label>Кому (адрес): <input type="text" name="to" required></label><br>
        <label>Сумма: <input type="text" name="amount" required></label><br>
        <button type="submit">Отправить</button>
    </form>

    <?php if ($sendResult): ?>
        <h4>Результат отправки:</h4>
        <pre><?= htmlspecialchars($sendResult) ?></pre>
    <?php endif; ?>
<?php endif; ?>
<p><a href="profiles.php">Профили</a></p>
<p><a href="profile.php">Профиль</a></p>

<p><a href="admin.php">Админка</a></p>
<p><a href="admin_ticket.php">Админ тикет</a></p>
<p><a href="chat.php">Чаты</a></p>
<p><a href="index.php">Назад</a></p>
