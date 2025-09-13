<?php
session_start();
require_once __DIR__ . '/../src/mfrelance.php';

$message = '';
$jwt = $_SESSION['jwt'] ?? '';

$mf = new MFrelance('localhost', 9999);

$captchaData = $mf->getCaptcha();
$captchaId = $captchaData['captchaID'] ?? '';
$captchaImage = $captchaData['captchaImg'] ?? '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $username = $_POST['username'] ?? '';
    $mnemonic = $_POST['mnemonic'] ?? '';
    $newPassword = $_POST['new_password'] ?? '';
    $captchaAnswer = $_POST['captcha_answer'] ?? '';
    $captchaIdPost = $_POST['captcha_id'] ?? '';

    $payload = json_encode([
        'username'       => $username,
        'mnemonic'       => $mnemonic,
        'new_password'   => $newPassword,
        'captcha_id'     => $captchaIdPost,
        'captcha_answer' => $captchaAnswer
    ]);

    $response = $mf->doRequest('restoreuser', false, $payload, true); // POST запрос

    if ($response['httpCode'] === 200) {
        $json = json_decode($response['response'], true);
        $message = $json['message'] ?? 'Аккаунт восстановлен';
        $encoded = $json['encrypted'] ?? '';
        if ($encoded) {
            $message .= "<br>Ваша JWT: " . htmlspecialchars($encoded);
            $_SESSION['jwt'] = $encoded;
            header('Location: dashboard.php');
            exit;
        }
    } else {
        $message = "Ошибка восстановления: " . $response['response'];
    }
}
?>

<h2>Восстановление аккаунта</h2>
<?php if ($message) echo "<p>$message</p>"; ?>

<form method="POST">
    Username: <input type="text" name="username" required><br>
    Мнемоника: <input type="text" name="mnemonic"><br>
    Новый пароль: <input type="password" name="new_password" required><br>
    <img src="<?= $captchaImage ?>" alt="captcha"><br>
    <input type="hidden" name="captcha_id" value="<?= $captchaId ?>">
    Введите капчу: <input type="text" name="captcha_answer" required><br>
    <button type="submit">Восстановить</button>
</form>

<p><a href="index.php">Назад</a></p>
