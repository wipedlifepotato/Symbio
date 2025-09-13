<?php
session_start();

$message = '';
$mnemonic = '';


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

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $username = $_POST['username'] ?? '';
    $password = $_POST['password'] ?? '';
    $captchaAnswer = $_POST['captcha_answer'] ?? '';

    $data = json_encode([
        'username' => $username,
        'password' => $password,
        'captcha_id' => $_POST['captcha_id'],
        'captcha_answer' => $captchaAnswer,
    ]);

    $ch = curl_init('http://localhost:9999/register');
    curl_setopt($ch, CURLOPT_POST, true);
    curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
    curl_setopt($ch, CURLOPT_HTTPHEADER, ['Content-Type: application/json']);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);

    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);

    if ($httpCode === 200) {
        $json = json_decode($response, true);
        $message = $json['message'] ?? 'Регистрация успешна';
        $mnemonic = $json['encrypted'] ?? '';
    } else {
        $message = "Ошибка регистрации: $response";
    }
}
?>

<h2>Регистрация</h2>
<?php if ($message) echo "<p>$message</p>"; ?>
<?php if ($mnemonic) echo "<p>Сохраните свою мнемонику: <b>$mnemonic</b></p>"; ?>

<form method="POST">
    Username: <input type="text" name="username" required><br>
    Password: <input type="password" name="password" required><br>
    <img src="<?= $captchaImage ?>" alt="captcha"><br>
    <input type="hidden" name="captcha_id" value="<?= $captchaId ?>">
    Введите капчу: <input type="text" name="captcha_answer" required><br>
    <button type="submit">Зарегистрироваться</button>
</form>
<a href="index.php">Вход</a>
