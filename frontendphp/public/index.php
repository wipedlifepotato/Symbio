<?php
session_start();
$message = '';

if (isset($_SESSION['jwt'])) {
    header('Location: dashboard.php'); 
    exit;
}

$captchaId = '';
$captchaImage = '';
$ch = curl_init('http://localhost:9999/captcha');
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_HEADER, true);
$response = curl_exec($ch);
$headerSize = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
$headers = substr($response, 0, $headerSize);
$body = substr($response, $headerSize);
curl_close($ch);

if (preg_match('/X-Captcha-ID:\s*(\S+)/i', $headers, $matches)) {
    $captchaId = trim($matches[1]);
}
$captchaImage = 'data:image/png;base64,' . base64_encode($body);

?>

<h2>Вход</h2>
<form method="POST" action="auth.php">
    Username: <input type="text" name="username" required><br>
    Password: <input type="password" name="password" required><br>
    <img src="<?= $captchaImage ?>" alt="captcha"><br>
    <input type="hidden" name="captcha_id" value="<?= $captchaId ?>">
    Введите капчу: <input type="text" name="captcha_answer" required><br>
    <button type="submit">Войти</button>
</form>
<p><a href="restore.php">Восстановить аккаунт</a></p>
<p><a href="register.php">Регистрация</a></p>

