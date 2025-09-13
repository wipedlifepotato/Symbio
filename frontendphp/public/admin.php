<?php
session_start();
require_once __DIR__ . '/../src/mfrelance.php';

$message = '';
$wallets = [];
$transactions = [];
$mf = new MFrelance('localhost', 9999);
$jwt = $_SESSION['jwt'] ?? '';

if (!$jwt) {
    $message = "JWT отсутствует. Пожалуйста, войдите в систему.";
} else {
    if (isset($_POST['block_user_id'])) {
        $res = $mf->doRequest('api/admin/block', $jwt, ['user_id' => (int)$_POST['block_user_id']], true);
        $message = $res['httpCode'] === 200 ? "Пользователь заблокирован" : "Ошибка: ".$res['response'];
    }
    if (isset($_POST['unblock_user_id'])) {
        $res = $mf->doRequest('api/admin/unblock', $jwt, ['user_id' => (int)$_POST['unblock_user_id']], true);
        $message = $res['httpCode'] === 200 ? "Пользователь разблокирован" : "Ошибка: ".$res['response'];
    }

    if (isset($_GET['wallets_user_id'])) {
	    $user_id = (int)$_GET['wallets_user_id'];
	    $res = $mf->doRequest('api/admin/wallets?user_id=' . $user_id, $jwt);
	    if ($res['httpCode'] === 200) {
		$wallets = json_decode($res['response'], true);
	    } else {
		$message = "Ошибка при получении кошельков: " . $res['response'];
	    }
    }

    if (isset($_POST['update_wallet_id']) && isset($_POST['new_balance'])) {
        $res = $mf->doRequest('api/admin/update_balance', $jwt, [
            'user_id' => (int)$_POST['update_wallet_id'],
            'balance' => $_POST['new_balance']
        ], true);
        $message = $res['httpCode'] === 200 ? "Баланс обновлён" : "Ошибка: ".$res['response'];
    }

    $tx_wallet_id = $_GET['tx_wallet_id'] ?? 0;
    $tx_limit = $_GET['tx_limit'] ?? 20;
    $tx_offset = $_GET['tx_offset'] ?? 0;
    $res = $mf->doRequest('api/admin/transactions', $jwt, [
        'wallet_id' => (int)$tx_wallet_id,
        'limit' => (int)$tx_limit,
        'offset' => (int)$tx_offset
    ], true);
    if ($res['httpCode'] === 200) {
        $transactions = json_decode($res['response'], true);
       // var_dump($transactions);
    } else {
        $message = "Ошибка при получении транзакций: " . $res['response'];
    }
}
?>

<h2>Admin Panel</h2>
<?php if ($message) echo "<p style='color:red;'>$message</p>"; ?>

<h3>User Management</h3>
<form method="POST">
    <input type="number" name="block_user_id" placeholder="User ID to Block">
    <button type="submit">Block</button>
</form>
<form method="POST">
    <input type="number" name="unblock_user_id" placeholder="User ID to Unblock">
    <button type="submit">Unblock</button>
</form>

<h3>Get User Wallets</h3>
<form method="GET">
    <input type="number" name="wallets_user_id" placeholder="User ID">
    <button type="submit">Get Wallets</button>
</form>

<?php if ($wallets): ?>
<table border="1" cellpadding="5">
<tr><th>ID</th><th>Currency</th><th>Balance</th><th>Address</th></tr>
<?php foreach ($wallets as $w): ?>
<tr>
<td><?= htmlspecialchars($w['UserID']) ?></td>
<td><?= htmlspecialchars($w['Currency']) ?></td>
<td><?= htmlspecialchars($w['Balance']) ?></td>
<td><?= htmlspecialchars($w['Address']) ?></td>
</tr>
<?php endforeach; ?>
</table>
<?php endif; ?>

<h3>Update Wallet Balance</h3>
<form method="POST">
    <input type="number" name="update_wallet_id" placeholder="Wallet ID">
    <input type="text" name="new_balance" placeholder="New Balance">
    <button type="submit">Set Balance</button>
</form>

<h3>Transactions</h3>
<form method="GET">
    <input type="number" name="tx_wallet_id" placeholder="Wallet ID (optional)" value="<?= htmlspecialchars($tx_wallet_id) ?>">
    <input type="number" name="tx_limit" placeholder="Limit" value="<?= htmlspecialchars($tx_limit) ?>">
    <input type="number" name="tx_offset" placeholder="Offset" value="<?= htmlspecialchars($tx_offset) ?>">
    <button type="submit">Get Transactions</button>
</form>

<?php if ($transactions): ?>
<table border="1" cellpadding="5">
<tr><th>ID</th><th>From</th><th>To</th><th>Amount</th><th>Currency</th><th>Created</th></tr>
<?php foreach ($transactions as $tx): ?>
<tr>
    <td><?= htmlspecialchars($tx['ID']) ?></td>
    <td>
        <?= isset($tx['FromWalletID']['Valid']) && $tx['FromWalletID']['Valid'] ? htmlspecialchars($tx['FromWalletID']['Int64']) : '' ?>
    </td>
    <td>
        <?= isset($tx['ToWalletID']['Valid']) && $tx['ToWalletID']['Valid'] ? htmlspecialchars($tx['ToWalletID']['Int64']) : '' ?>
        (<?= isset($tx['ToAddress']['Valid']) && $tx['ToAddress']['Valid'] ? htmlspecialchars($tx['ToAddress']['String']) : '' ?>)
    </td>
    <td><?= htmlspecialchars($tx['Amount']) ?></td>
    <td><?= htmlspecialchars($tx['Currency']) ?></td>
    <td><?= htmlspecialchars($tx['CreatedAt']) ?></td>
</tr>
<?php endforeach; ?>
</table>
<?php endif; ?>

