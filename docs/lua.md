````markdown
# Lua HTTP Handlers with JWT Authentication

This document demonstrates how to build HTTP endpoints in Lua using `register_handler`.  
It also shows how to protect endpoints with JWT authentication and how to use other Lua helpers provided by the Go backend.

---

## Handlers Overview

Each handler is registered with the following pattern:

```lua
register_handler("/path", function(req)
    -- your logic
    return '{"json":"response"}'
end)
````

* `req.method` → HTTP method (`GET`, `POST`, etc.).
* `req.params` → request parameters (query string or form data).
* Return value must be a JSON string.

---

## 1. Ping Handler

A minimal handler that always responds with a `pong`.

```lua
register_handler("/ping", function(req)
    return '{"pong":true,"method":"'..req.method..'"}'
end)
```

**Example request:**

```
GET /ping
```

**Response:**

```json
{"pong":true,"method":"GET"}
```

---

## 2. Echo Handler

This handler reads parameters from the request and echoes them back.

```lua
register_handler("/echo", function(req)
    local test = req.params["test"] or "nil"
    local foo  = req.params["foo"] or "nil"
    return '{"method":"'..req.method..'","test":"'..test..'","foo":"'..foo..'"}'
end)
```

**Example request:**

```
GET /echo?test=hello&foo=bar
```

**Response:**

```json
{"method":"GET","test":"hello","foo":"bar"}
```

---

## 3. Server Config Example

Handler that returns the configured server port from `config.Port`.

```lua
register_handler("/test3", function(req)
    local port = config.Port or "undefined"
    return string.format('{"testpass": true, "server_port": "%s"}', port)
end)
```

**Response example:**

```json
{"testpass": true, "server_port": "8080"}
```

---

## 4. JWT Authentication Example

JWT (JSON Web Token) can be used to protect sensitive endpoints.

### Generate a JWT

```lua
-- Example: restoring a user and generating a JWT
local mnemonic = "ice kite panda monkey apple cat fish ice monkey zebra zebra panda"
local restored = restore_user("testuser", mnemonic)

if restored ~= nil then
    change_password("testuser", "12345678")
    local token = generate_jwt(restored.id, restored.username)
    print("JWT:", token)
end
```

**Example generated token:**

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

### Protected Endpoint: `/mywallet`

This handler requires a valid JWT token in the `Authorization` parameter.

```lua
register_handler("/mywallet", function(req)
    local token = req.params["Authorization"]
    if not token then
        return '{"error":"missing token"}'
    end

    local user, err = get_user_from_jwt(token)
    if not user then
        return '{"error":"invalid token: '..err..'"}'
    end

    local user_id = user.user_id
    local username = user.username

    return '{"user_id":'..user_id..', "username":"'..username..'"}'
end)
```

**Example request:**

```
GET /mywallet?Authorization=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Example response:**

```json
{"user_id":42, "username":"alice"}
```

---

## 5. Config Globals

The `config` global exposes backend configuration values.

```lua
print(config.PostgresHost)
print(config.RedisHost)
print(config.Port)
print(config.ElectrumHost)
print(config.MoneroAddress)
print(config.BitcoinAddress)
```

### Available fields

* **Postgres**

  * `config.PostgresHost`
  * `config.PostgresPort`
  * `config.PostgresUser`
  * `config.PostgresPassword`
  * `config.PostgresDB`
* **Redis**

  * `config.RedisHost`
  * `config.RedisPort`
  * `config.RedisPassword`
* **Server**

  * `config.Port`
  * `config.JWTToken`
  * `config.ListenAddr`
* **Electrum**

  * `config.ElectrumHost`
  * `config.ElectrumPort`
  * `config.ElectrumUser`
  * `config.ElectrumPassword`
* **Monero**

  * `config.MoneroHost`
  * `config.MoneroPort`
  * `config.MoneroUser`
  * `config.MoneroPassword`
  * `config.MoneroAddress`
  * `config.MoneroCommission`
* **Bitcoin**

  * `config.BitcoinAddress`
  * `config.BitcoinCommission`
* **Constants**

  * `config.MaxProfiles`

---

## 6. Electrum API

### Create a new address

```lua
local addr, err = electrum_create_address()
```

### Block/unblock withdrawals

```lua
electrum_set_withdraw_blocked(true)
local blocked = electrum_is_withdraw_blocked()
```

### Pay to one address

```lua
local txid, err = electrum_pay_to("bcrt1qaddress...", "0.01")
```

### Pay to many addresses

```lua
local outputs = {
    {"addr1", "0.01"},
    {"addr2", "0.02"}
}
local txid, err = electrum_pay_to_many(outputs)
```

### Get balance

```lua
local balance, err = electrum_get_balance("bcrt1qaddress...")
```

### List addresses

```lua
local addrs, err = electrum_list_addresses()
```

---

## 7. User & Admin Helpers

### User management

```lua
local user = get_user("alice")
block_user(user.id)
unblock_user(user.id)
print(is_user_blocked("alice"))
```

### Admin management

```lua
local ok = make_admin(42)
local ok = remove_admin(42)
local isAdm = is_admin(42)
```

### Passwords

```lua
local ok = verify_password("secret", user.password_hash)
local changed = change_password("alice", "newpass")
```

### Restore user from mnemonic

```lua
local restored = restore_user("alice", "word1 word2 word3 ...")
```

---

## 8. Profiles

### Get a profile

```lua
local p = get_profile(42)
print(p.full_name, p.bio, p.rating)
```

### Upsert a profile

```lua
upsert_profile(42, "Alice Doe", "Blockchain dev", {"Go","Lua"}, "avatar.png")
```

### Get multiple profiles

```lua
local profiles = get_profiles(10, 0)
for _, p in ipairs(profiles) do
    print(p.full_name, p.rating)
end
```

---

## 9. Misc Helpers

* `generate_mnemonic()` → returns a new wallet mnemonic.
* `helloGo("Alice")` → logs from Go and returns `"Hi, Alice"`.

---

## 10. Example Flow

1. Restore a user from mnemonic.
2. Change the password.
3. Generate a JWT.
4. Access `/mywallet` with the token.

```lua
local restored = restore_user("alice", "ice kite panda monkey ...")
change_password("alice", "12345678")
local token = generate_jwt(restored.id, restored.username)

print("JWT:", token)
```

**Request:**

```
GET /mywallet?Authorization=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{"user_id":42, "username":"alice"}
```

````markdown
---

## 11. Wallet / Balance Helpers

These functions allow Lua scripts to **read and modify user balances** in the database.  
They support any currency (BTC, XMR, etc.) and ensure safe updates.

### Get balance

```lua
local bal = get_balance(user_id, "BTC")
if bal then
    print("User BTC balance:", bal)
else
    print("Failed to fetch balance")
end
````

**Parameters**

* `user_id` → string or number, the user’s ID.
* `currency` → string, e.g. `"BTC"` or `"XMR"`.

**Returns**

* balance as a string (decimal) if success.
* `nil, error_message` on failure.

---

### Add balance

```lua
local ok, err = add_balance(user_id, "BTC", "0.005")
if ok then
    print("Balance increased")
else
    print("Failed to increase balance:", err)
end
```

**Parameters**

* `user_id` → string/number, the user ID.
* `currency` → string.
* `amount` → string decimal, e.g. `"0.005"`.

**Returns**

* `true` on success.
* `nil, error_message` on failure.

---

### Subtract balance

```lua
local ok, err = sub_balance(user_id, "BTC", "0.002")
if ok then
    print("Balance decreased")
else
    print("Failed to subtract balance:", err)
end
```

**Parameters**

* `user_id` → string/number, the user ID.
* `currency` → string.
* `amount` → string decimal.

**Returns**

* `true` on success.
* `nil, "insufficient balance"` if user does not have enough funds.
* `nil, error_message` on other errors.

---

### Example: Transfer between users

```lua
local from_id = 42
local to_id = 43
local amount = "0.01"

local ok, err = sub_balance(from_id, "BTC", amount)
if not ok then
    print("Failed to debit sender:", err)
else
    add_balance(to_id, "BTC", amount)
    print("Transfer completed")
end
```

This pattern ensures **atomic checks** and safe updates via Go/DB.

---

# Summary

* `register_handler` allows writing Lua HTTP endpoints.
* `config` exposes backend settings.
* JWT (`generate_jwt`, `get_user_from_jwt`) provides authentication.
* Electrum methods allow Bitcoin wallet operations.
* User/admin helpers provide account management.
* Profile methods manage user profiles.
* The `/mywallet` endpoint demonstrates JWT-protected API access.



