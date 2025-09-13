<?php
session_start();
require_once __DIR__ . '/../src/mfrelance.php';

$message = '';

if (isset($_SESSION['jwt'])) {
    header('Location: dashboard.php');
    exit;
}
$mf = new MFrelance('localhost', 9999);

$captcha = $mf->getCaptcha();
$captchaId = $captcha['captchaID'] ?? '';
$captchaImage = $captcha['captchaImg'] ?? '';

?>

<h2>Вход</h2>
<form method="POST" action="auth.php">
    Username: <input type="text" name="username" required><br>
    Password: <input type="password" name="password" required><br>
    <img src="<?= htmlspecialchars($captchaImage) ?>" alt="captcha"><br>
    <input type="hidden" name="captcha_id" value="<?= htmlspecialchars($captchaId) ?>">
    Введите капчу: <input type="text" name="captcha_answer" required><br>
    <button type="submit">Войти</button>
</form>

<p><a href="restore.php">Восстановить аккаунт</a></p>
<p><a href="register.php">Регистрация</a></p>

