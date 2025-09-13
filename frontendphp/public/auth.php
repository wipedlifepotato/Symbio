<?php
session_start();

$username = $_POST['username'] ?? '';
$password = $_POST['password'] ?? '';
$captchaId = $_POST['captcha_id'] ?? '';
$captchaAnswer = $_POST['captcha_answer'] ?? '';

$data = json_encode([
    'username' => $username,
    'password' => $password,
    'captcha_id' => $captchaId,
    'captcha_answer' => $captchaAnswer,
]);

$ch = curl_init('http://localhost:9999/auth');
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
curl_setopt($ch, CURLOPT_HTTPHEADER, ['Content-Type: application/json']);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
$response = curl_exec($ch);
$httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);

if ($httpCode === 200) {
    $json = json_decode($response, true);
    $jwt = $json['token'] ?? '';
    if ($jwt) {
        $_SESSION['jwt'] = $jwt;
        header('Location: dashboard.php');
        exit;
    }
    $message = 'Ошибка: JWT не получен';
} else {
    $message = "Ошибка авторизации: $response";
}
?>

<p><?= htmlspecialchars($message) ?></p>
<p><a href="index.php">Назад</a></p>
