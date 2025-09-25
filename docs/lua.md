# Lua API Documentation

This document provides comprehensive documentation for the Lua API bindings exposed by the Go backend. These functions allow Lua scripts to interact with the database, wallets, authentication, and other system components.

---

## Table of Contents

1. [HTTP Handlers](#1-http-handlers)
2. [Config Globals](#2-config-globals)
3. [BigMath](#3-bigmath)
4. [Database Operations](#4-database-operations)
5. [Disputes](#5-disputes)
6. [Escrow](#6-escrow)
7. [Reviews](#7-reviews)
8. [JWT Authentication](#8-jwt-authentication)
9. [Electrum (Bitcoin)](#9-electrum-bitcoin)
10. [Monero](#10-monero)
11. [WASM](#11-wasm)
12. [Chat](#12-chat)
13. [Tasks](#13-tasks)
14. [User & Admin Helpers](#14-user--admin-helpers)
15. [Profiles](#15-profiles)
16. [Wallet & Balance](#16-wallet--balance)
17. [Misc Helpers](#17-misc-helpers)

---

## 1. HTTP Handlers

### register_handler(path, handlerFn)

Registers an HTTP handler for the given path.

**Parameters:**
- `path` (string): The URL path to handle.
- `handlerFn` (function): Lua function that takes a `req` table and returns a JSON string.

**Request Table (`req`):**
- `req.method` (string): HTTP method (GET, POST, etc.).
- `req.params` (table): Query parameters or form data.

**Example:**
```lua
register_handler("/ping", function(req)
    return '{"pong": true, "method": "' .. req.method .. '"}'
end)
```

---

## 2. Config Globals

The `config` table exposes backend configuration values.

**Available Fields:**
- **Postgres:** `PostgresHost`, `PostgresPort`, `PostgresUser`, `PostgresPassword`, `PostgresDB`
- **Redis:** `RedisHost`, `RedisPort`, `RedisPassword`
- **Server:** `Port`, `JWTToken`, `ListenAddr`
- **Electrum:** `ElectrumHost`, `ElectrumPort`, `ElectrumUser`, `ElectrumPassword`
- **Monero:** `MoneroHost`, `MoneroPort`, `MoneroUser`, `MoneroPassword`, `MoneroAddress`, `MoneroCommission`
- **Bitcoin:** `BitcoinAddress`, `BitcoinCommission`
- **Constants:** `MaxProfiles`, `MaxAvatarSize`, `MaxAddrPerBlock`

**Example:**
```lua
print(config.Port)  -- e.g., "8080"
```

---

## 3. BigMath

Provides arbitrary-precision arithmetic using Go's `math/big` package.

### BigMath.New(val)

Creates a new big.Float from a string value.

**Parameters:**
- `val` (string): Numeric string.

**Returns:**
- Userdata (big.Float) or nil if invalid.

**Example:**
```lua
local num = BigMath.New("123.456")
```

### BigMath.Add(a, b)

Adds two big.Float values.

**Parameters:**
- `a`, `b` (userdata): big.Float instances.

**Returns:**
- Userdata (result).

### BigMath.Sub(a, b)

Subtracts b from a.

### BigMath.Mul(a, b)

Multiplies a and b.

### BigMath.Quo(a, b)

Divides a by b.

### BigMath.String(a)

Converts big.Float to string.

**Returns:**
- String representation.

**Example:**
```lua
local a = BigMath.New("10.5")
local b = BigMath.New("2.0")
local sum = BigMath.Add(a, b)
print(BigMath.String(sum))  -- "12.5"
```

---

## 4. Database Operations

### Redis

#### get_captcha(id)

Retrieves a captcha value from Redis.

**Parameters:**
- `id` (string): Captcha ID.

**Returns:**
- String value or nil if not found.

#### set_captcha(id, val, exp)

Sets a captcha value with expiration.

**Parameters:**
- `id` (string): Captcha ID.
- `val` (string): Value.
- `exp` (number): Expiration in seconds.

**Returns:**
- true or false.

### Postgres

#### pg_query(query, params)

Executes a SQL query.

**Parameters:**
- `query` (string): SQL query.
- `params` (table, optional): Parameters as array.

**Returns:**
- Table of rows (SELECT) or affected rows (INSERT/UPDATE/DELETE), or false, error.

**Supported Queries:** SELECT, INSERT, UPDATE, DELETE.

**Example:**
```lua
local rows, err = pg_query("SELECT id, name FROM users WHERE id = $1", {42})
if rows then
    for _, row in ipairs(rows) do
        print(row.id, row.name)
    end
end
```

#### create_task(tbl)

Creates a new task.

**Parameters:**
- `tbl` (table): {client_id, title, description, category, budget, currency, status}

**Returns:**
- true or false, error.

#### get_task(id)

Retrieves a task by ID.

**Returns:**
- Table with task fields or nil, error.

#### create_chat_request(requesterID, requestedID)

Creates a chat request.

**Returns:**
- true or false, error.

#### update_chat_request(requesterID, requestedID, status)

Updates chat request status.

#### get_chat_messages(chatID)

Gets all messages for a chat.

**Returns:**
- Table of messages or nil, error.

#### get_chat_messages_paged(chatID, limit, offset)

Gets messages for a chat with pagination.

**Parameters:**
- `limit` (number): Maximum messages to return (default 50)
- `offset` (number): Number of messages to skip (default 0)

**Returns:**
- Table of messages or nil, error.

#### send_chat_message(chatID, senderID, message)

Sends a message.

**Returns:**
- true or false, error.

#### get_chat_rooms_for_user(userID)

Gets chat rooms for a user.

**Returns:**
- Table of rooms or nil, error.

---

## 5. Disputes

#### create_dispute(taskID, openedBy)

Creates a dispute.

**Returns:**
- Dispute ID or nil, error.

#### get_dispute(id)

Gets dispute by ID.

**Returns:**
- Table with dispute fields or nil, error.

#### update_dispute_status(id, status, resolution)

Updates dispute status.

**Parameters:**
- `resolution` (string, optional): Resolution text.

**Returns:**
- true or false, error.

#### get_dispute_messages(disputeID)

Gets all dispute messages.

**Returns:**
- Table of messages or nil, error.

#### get_dispute_messages_paged(disputeID, limit, offset)

Gets dispute messages with pagination.

**Parameters:**
- `limit` (number): Maximum messages to return (default 50)
- `offset` (number): Number of messages to skip (default 0)

**Returns:**
- Table of messages or nil, error.

---

## 6. Escrow

#### create_escrow(taskID, clientID, freelancerID, amount, currency)

Creates an escrow.

**Parameters:**
- `amount` (string): Amount as string.

**Returns:**
- Escrow ID or nil, error.

#### get_escrow_by_task(taskID)

Gets escrow by task ID.

**Returns:**
- Table with escrow fields or nil, error.

---

## 7. Reviews

#### create_review(taskID, reviewerID, reviewedID, rating, comment)

Creates a review.

**Parameters:**
- `rating` (number): 1-5.

**Returns:**
- Review ID or nil, error.

#### get_reviews_by_user(userID)

Gets reviews for a user.

**Returns:**
- Table of reviews or nil, error.

---

## 8. JWT Authentication

#### get_user_from_jwt(token)

Parses JWT and returns user info.

**Returns:**
- Table {user_id, username} or nil, error.

#### generate_jwt(userID, username)

Generates a JWT token.

**Returns:**
- Token string or nil.

---

## 9. Electrum (Bitcoin)

#### electrum_create_address()

Creates a new Bitcoin address.

**Returns:**
- Address string or nil, error.

#### electrum_set_withdraw_blocked(blocked)

Sets withdrawal block status.

**Parameters:**
- `blocked` (boolean)

#### electrum_is_withdraw_blocked()

Checks if withdrawals are blocked.

**Returns:**
- boolean

#### electrum_pay_to_many(outputs)

Pays to multiple addresses.

**Parameters:**
- `outputs` (table): Array of {address, amount}

**Returns:**
- TxID or nil, error.

#### electrum_get_balance(addr)

Gets balance for an address.

**Returns:**
- Balance (number) or nil, error.

#### electrum_pay_to(destination, amount)

Pays to a single address.

**Returns:**
- TxID or nil, error.

#### electrum_list_addresses()

Lists all addresses.

**Returns:**
- Table of addresses or nil, error.

---

## 10. Monero

#### monero_get_balance()

Gets total and unlocked balance.

**Returns:**
- total (number), unlocked (number)

#### monero_create_address(label)

Creates a new Monero address.

**Returns:**
- Address string

#### monero_transfer(dest, amount)

Transfers XMR.

**Parameters:**
- `amount` (number): In XMR.

**Returns:**
- TxHash string

#### monero_get_subaddress_info(account, sub)

Gets subaddress info.

**Returns:**
- total (number), unlocked (number), address (string)

#### monero_get_subaddress_balance(account, sub)

Gets subaddress balance.

**Returns:**
- total (number), unlocked (number)

---

## 11. WASM

#### LoadWasmModule(path)

Loads a WASM module (not a global function, call directly).

#### wasm_call_bytes(fnName, ...)

Calls a WASM function with byte arguments.

**Returns:**
- Result string or error.

#### wasm_call(fnName, input)

Calls a WASM function with int input.

**Returns:**
- Result number or error.

---

## 12. Chat

#### create_chat_room()

Creates a new chat room.

**Returns:**
- Room ID or nil, error.

#### add_user_to_chat(userID, chatRoomID)

Adds user to chat.

**Returns:**
- true or nil, error.

#### get_chat_participants(chatRoomID)

Gets participants.

**Returns:**
- Table of user IDs or nil, error.

#### create_chat_message(chatRoomID, senderID, msg)

Creates a message.

**Returns:**
- true or nil, error.

#### get_chat_messages(chatRoomID)

Gets messages.

**Returns:**
- Table of messages or nil, error.

#### create_chat_request(requesterID, requestedID)

Creates chat request.

**Returns:**
- true or nil, error.

#### accept_chat_request(requesterID, requestedID)

Accepts request.

**Returns:**
- true or nil, error.

#### delete_chat_request(requesterID, requestedID)

Deletes request.

**Returns:**
- true or nil, error.

#### delete_chat_participant(chatRoomID, userID)

Removes participant.

**Returns:**
- true or nil, error.

---

## 13. Tasks

#### count_open_tasks()

Counts open tasks.

**Returns:**
- Count (number) or nil, error.

#### count_tasks_by_client_and_status(clientID, status)

Counts tasks by client and status.

**Returns:**
- Count or nil, error.

#### get_tasks_by_client_paged(clientID, limit, offset)

Gets tasks for client with pagination.

**Returns:**
- Table of tasks or nil, error.

#### get_open_tasks_paged(limit, offset)

Gets open tasks with pagination.

**Returns:**
- Table of tasks or nil, error.

#### get_tasks_by_client_and_status_paged(clientID, status, limit, offset)

Gets tasks by client and status with pagination.

**Returns:**
- Table of tasks or nil, error.

---

## 14. User & Admin Helpers

#### get_user(username)

Gets user by username.

**Returns:**
- Table {id, password_hash} or nil.

#### get_user_by_id(id)

Gets user by ID.

**Returns:**
- Table with user fields or nil, error.

#### block_user(userID)

Blocks a user.

**Returns:**
- true or error string.

#### unblock_user(userID)

Unblocks a user.

**Returns:**
- true or error string.

#### is_user_blocked(username)

Checks if user is blocked.

**Returns:**
- boolean or nil.

#### verify_password(password, hashed)

Verifies password.

**Returns:**
- boolean

#### change_password(username, newPassword)

Changes password.

**Returns:**
- Table or nil, error.

#### restore_user(username, mnemonic)

Restores user from mnemonic.

**Returns:**
- Table {id, username} or nil.

#### generate_jwt(userID, username)

Generates JWT.

**Returns:**
- Token or nil.

#### is_admin(userID)

Checks if user is admin.

**Returns:**
- boolean or nil, error.

#### make_admin(userID)

Makes user admin.

**Returns:**
- true or false, error.

#### remove_admin(userID)

Removes admin status.

**Returns:**
- true or false, error.

#### add_permission(userID, perm)

Adds a specific permission to user.

**Parameters:**
- `perm` (number): Permission constant (1, 2, 4, etc.)

**Returns:**
- true or false, error.

#### remove_permission(userID, perm)

Removes a specific permission from user.

**Parameters:**
- `perm` (number): Permission constant to remove.

**Returns:**
- true or false, error.

#### set_permissions(userID, permissions)

Sets exact permissions mask for user.

**Parameters:**
- `permissions` (number): Complete permissions bitmask.

**Returns:**
- true or false, error.

**Example:**
```lua
-- Add permission to change balances
add_permission(123, 1)

-- Remove permission to manage disputes
remove_permission(123, 4)

-- Set user to have all permissions
set_permissions(123, 7)

-- Set user to have only balance change permission
set_permissions(123, 1)
```

#### has_permission(userID, perm)

Checks if user has specific permission.

**Parameters:**
- `perm` (number): Permission constant (1=CanChangeBalance, 2=CanBlockUsers, 4=CanManageDisputes)

**Returns:**
- boolean or nil, error.

**Example:**
```lua
local can_change = has_permission(123, 1)  -- CanChangeBalance
if can_change then
    -- Allow balance change
end
```

#### require_permission(perm)

Middleware function that checks permission (used in Go handlers, not directly in Lua).

#### Permission Constants

```lua
CanChangeBalance = 1    -- Изменение балансов пользователей
CanBlockUsers = 2       -- Блокировка/разблокировка пользователей
CanManageDisputes = 4   -- Управление спорами
```

**Example Usage:**
```lua
-- Check if user can change balances
if has_permission(user_id, 1) then
    -- Allow balance modification
end

-- Check if user can manage disputes
if has_permission(user_id, 4) then
    -- Allow dispute management
end

-- Check multiple permissions
if has_permission(user_id, 1) and has_permission(user_id, 2) then
    -- User has both balance and user management permissions
end
```

---

## 15. Profiles

#### get_profile(userID)

Gets user profile.

**Returns:**
- Table {user_id, full_name, bio, skills, avatar, rating, completed_tasks, is_admin, admin_title, permissions} or nil, error.

#### upsert_profile(userID, fullName, bio, skills, avatar)

Updates or inserts profile.

**Parameters:**
- `skills` (table): Array of strings.

**Returns:**
- true or error.

#### get_profiles(limit, offset)

Gets multiple profiles.

**Returns:**
- Table of profiles or nil, error.

---

## 16. Wallet & Balance

#### get_wallet(userID, currency)

Gets wallet info.

**Returns:**
- Table {id, balance, address} or nil, error.

#### set_balance(userID, currency, newBalance)

Sets balance.

**Returns:**
- true or false, error.

#### get_balance(userID, currency)

Gets balance as string.

**Returns:**
- Balance string or nil, error.

#### add_balance(userID, currency, amount)

Adds to balance.

**Parameters:**
- `amount` (string)

**Returns:**
- true or nil, error.

#### sub_balance(userID, currency, amount)

Subtracts from balance.

**Returns:**
- true or nil, error (e.g., "insufficient balance")

#### get_transactions(walletID, limit, offset)

Gets transactions.

**Returns:**
- Table of transactions or nil, error.

---

## 17. Misc Helpers

#### generate_mnemonic()

Generates a new mnemonic.

**Returns:**
- Mnemonic string.

#### helloGo(name)

Test function.

**Returns:**
- "Hi, " + name

---

# Summary

This Lua API provides extensive access to backend functionality, including database operations, wallet management, authentication, and more. Use these functions in your Lua scripts to build custom handlers and logic.
