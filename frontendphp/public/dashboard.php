<?php
session_start();

$message = '';
$walletInfo = null;

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
        unset($_SESSION['JWT']);
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
<?php endif; ?>

<p><a href="index.php">Назад</a></p>
