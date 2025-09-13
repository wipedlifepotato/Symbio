<?php
session_start();

$message = '';
$walletInfo = null;
$sendResult = null;

$jwt = $_SESSION['jwt'] ?? '';

if (!$jwt) {
    $message = "JWT отсутствует. Пожалуйста, войдите в систему.";
} else {

    $ch = curl_init("http://localhost:9999/api/wallet?currency=BTC");
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        "Authorization: Bearer $jwt",
        "Content-Type: application/json"
    ]);
    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);

    if ($httpCode === 200) {
        $walletInfo = json_decode($response, true);
    } else {
        $message = "Ошибка при получении кошелька: $response";
        unset($_SESSION['jwt']);
    }

    if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['to'], $_POST['amount'])) {
        $to = $_POST['to'];
        $amount = $_POST['amount'];

        $ch = curl_init("http://localhost:9999/api/wallet/bitcoinSend?to=$to&amount=$amount");
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            "Authorization: Bearer $jwt",
            "Content-Type: application/json"
        ]);
        $sendResult = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);

        if ($httpCode !== 200) {
            $message = "Ошибка отправки BTC: $sendResult";
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

<p><a href="index.php">Назад</a></p>
