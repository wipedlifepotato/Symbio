<?php
session_start();
require_once __DIR__ . '/../src/mfrelance.php';

$message = '';
$profile = null;

$mf = new MFrelance('localhost', 9999);

$jwt = $_SESSION['jwt'] ?? '';

if (!$jwt) {
    $message = "JWT отсутствует. Пожалуйста, войдите в систему.";
} else {
    $profileResponse = $mf->doRequest("profile", $jwt);
    if ($profileResponse['httpCode'] === 200) {
        $profile = json_decode($profileResponse['response'], true);
    } else {
        $message = "Ошибка при получении профиля: " . $profileResponse['response'];
        //unset($_SESSION['jwt']);
    }

    if ($_SERVER['REQUEST_METHOD'] === 'POST') {
        $full_name = $_POST['full_name'] ?? '';
        $bio = $_POST['bio'] ?? '';
        $skills = isset($_POST['skills']) ? explode(',', $_POST['skills']) : [];
        $avatarBase64 = '';

        // Если загружен файл
        if (!empty($_FILES['avatar']['tmp_name'])) {
            $avatarData = file_get_contents($_FILES['avatar']['tmp_name']);
            $avatarBase64 = 'data:' . $_FILES['avatar']['type'] . ';base64,' . base64_encode($avatarData);
        } else {
            // Если не загружен файл, используем старый Base64
            $avatarBase64 = $_POST['avatar_base64'] ?? '';
        }

        $updateResponse = $mf->doRequest(
            "profile",
            $jwt,
            json_encode([
                "full_name" => $full_name,
                "bio" => $bio,
                "skills" => $skills,
                "avatar" => $avatarBase64
            ]),
            "POST"
        );

        if ($updateResponse['httpCode'] === 200) {
            $message = "Профиль обновлен успешно!";
            $profile['full_name'] = $full_name;
            $profile['bio'] = $bio;
            $profile['skills'] = $skills;
            $profile['avatar'] = $avatarBase64;
        } else {
            $message = "Ошибка при обновлении профиля: " . $updateResponse['response'];
        }
    }
}
?>

<h2>Профиль пользователя</h2>

<?php if ($message) echo "<p style='color:red;'>$message</p>"; ?>

<?php if ($profile): ?>
    <form method="POST" enctype="multipart/form-data">
        <label>Полное имя:<br>
            <input type="text" name="full_name" value="<?= htmlspecialchars($profile['full_name'] ?? '') ?>">
        </label><br><br>

        <label>Bio:<br>
            <textarea name="bio"><?= htmlspecialchars($profile['bio'] ?? '') ?></textarea>
        </label><br><br>

        <label>Навыки (через запятую):<br>
            <input type="text" name="skills" value="<?= htmlspecialchars(implode(',', $profile['skills'] ?? [])) ?>">
        </label><br><br>

        <label>Аватар:<br>
            <?php if (!empty($profile['avatar'])): ?>
                <img src="<?= $profile['avatar'] ?>" alt="Аватар" style="max-width:100px;"><br>
            <?php endif; ?>
            <input type="file" name="avatar">
            <input type="hidden" name="avatar_base64" value="<?= htmlspecialchars($profile['avatar'] ?? '') ?>">
        </label><br><br>

        <button type="submit">Сохранить профиль</button>
    </form>
<?php endif; ?>

<p><a href="index.php">Назад</a></p>

