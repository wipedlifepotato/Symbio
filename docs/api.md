# API
## Version: 1.0


### Security
**BearerAuth**  

|apiKey|*API Key*|
|---|---|
|In|header|
|Name|Authorization|

### /admin/delete-user-tasks

#### POST
##### Summary:

Delete all tasks of a specific user

##### Description:

Allows an admin to delete all tasks belonging to a user by their ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| user_id | query | User ID | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | success and deleted count | object |
| 400 | Invalid user_id | object |
| 401 | Unauthorized | object |
| 403 | Admin rights required | object |
| 405 | Method not allowed | object |
| 500 | Failed to delete tasks | object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /admin/disputes/assign

#### POST
##### Summary:

Assign dispute

##### Description:

Assign a dispute to the current admin

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body | Dispute ID | Yes | [handlers.AssignDisputeRequest](#handlers.AssignDisputeRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |
| 404 | Not Found | object |
| 500 | Internal Server Error | object |

### /api/admin/IIsAdmin

#### GET
##### Summary:

Check if user is admin

##### Description:

Returns true/false if current user has admin privileges

##### Responses

| Code | Description |
| ---- | ----------- |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/admin/block

#### POST
##### Summary:

Block user

##### Description:

Blocks a user by userID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| request | body | UserID payload | Yes | [handlers.AdminRequest](#handlers.AdminRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | user blocked | string |
| 400 | Bad Request | string |
| 500 | Internal Server Error | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/admin/check

#### GET
##### Summary:

Check if user is admin

##### Description:

Returns true/false if current user has admin privileges

##### Responses

| Code | Description |
| ---- | ----------- |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/admin/getRandomTicket

#### GET
##### Summary:

Get random opened ticket (admin)

##### Description:

Set ticket to admin (random)

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [models.TicketDoc](#models.TicketDoc) |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| bearerAuth | |

### /api/admin/make

#### POST
##### Summary:

Grant admin rights

##### Description:

Makes a user admin by userID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| request | body | UserID payload | Yes | [handlers.AdminRequest](#handlers.AdminRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | user is now admin | string |
| 400 | invalid request body | string |
| 401 | unauthorized | string |
| 403 | admin rights required | string |
| 500 | internal server error | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/admin/remove

#### POST
##### Summary:

Revoke admin rights

##### Description:

Removes admin status from a user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| request | body | UserID payload | Yes | [handlers.AdminRequest](#handlers.AdminRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | user admin removed | string |
| 400 | Bad Request | string |
| 500 | Internal Server Error | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/admin/transactions

#### POST
##### Summary:

Admin: View transactions

##### Description:

Allows admin to view transactions by wallet or all transactions with pagination

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| request | body | Request payload | Yes | [handlers.AdminTransactionsRequest](#handlers.AdminTransactionsRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | id:int, from_wallet_id:int, to_wallet_id:int, to_address:string, task_id:int, amount:string, currency:string, confirmed:bool, created_at:string | [ object ] |
| 400 | Bad Request | string |
| 500 | Internal Server Error | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/admin/unblock

#### POST
##### Summary:

Unblock user

##### Description:

Unblocks a user by userID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| request | body | UserID payload | Yes | [handlers.AdminRequest](#handlers.AdminRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | user unblocked | string |
| 400 | Bad Request | string |
| 500 | Internal Server Error | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/admin/update_balance

#### POST
##### Summary:

Update wallet balance

##### Description:

Allows admin to set a new balance for a wallet

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| request | body | Wallet balance payload | Yes | [handlers.AdminUpdateBalanceRequest](#handlers.AdminUpdateBalanceRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | balance updated | string |
| 400 | Bad Request | string |
| 500 | Internal Server Error | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/admin/wallets

#### GET
##### Summary:

Get user wallets

##### Description:

Returns all wallets for a given user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| user_id | query | User ID | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [ [models.Wallet](#models.Wallet) ] |
| 400 | Bad Request | string |
| 500 | Internal Server Error | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/chats/delete

#### DELETE
##### Summary:

Delete chat room

##### Description:

Allows an admin to delete a chat room by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| chat_id | query | Chat room ID | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Result message | object |
| 400 | Bad chat_id | object |
| 401 | Unauthorized | object |
| 500 | Internal server error | object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/disputes/create

#### POST
##### Summary:

Create dispute

##### Description:

Opens a new dispute for a task (only client or accepted freelancer can open)

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body | Resolution payload | Yes | [handlers.ResolveDisputeRequest](#handlers.ResolveDisputeRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Dispute created successfully | object |
| 400 | Invalid request or dispute already exists | string |
| 401 | Unauthorized | string |
| 403 | Forbidden | string |
| 404 | Task not found | string |
| 500 | Internal server error | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/disputes/details

#### GET
##### Summary:

Get dispute details

##### Description:

Returns dispute info, related task, escrow balance and messages

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | query | Dispute ID | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | success flag and dispute details | object |
| 400 | Invalid dispute ID | string |
| 404 | Dispute, task or escrow not found | string |
| 500 | Failed to get messages | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/disputes/get

#### GET
##### Summary:

Get dispute details

##### Description:

Returns details and messages for a specific dispute

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | query | Dispute ID | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Dispute details | object |
| 400 | Invalid dispute ID | string |
| 404 | Dispute not found | string |
| 500 | Internal server error | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/disputes/message

#### POST
##### Summary:

Send dispute message

##### Description:

Allows client, freelancer or assigned admin to send a message in a dispute

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body | Dispute message | Yes | [handlers.SendDisputeMessageRequest](#handlers.SendDisputeMessageRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Message sent | object |
| 400 | Invalid JSON or request | string |
| 401 | Unauthorized | string |
| 403 | Forbidden | string |
| 404 | Dispute not found | string |
| 500 | Failed to send message | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/disputes/my

#### GET
##### Summary:

Get user disputes

##### Description:

Returns all disputes where the user is a client or accepted freelancer

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | User disputes list | object |
| 401 | Unauthorized | string |
| 500 | Internal server error | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/disputes/open

#### GET
##### Summary:

Get all open disputes

##### Description:

Returns a list of disputes with status "open"

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | success flag and disputes list | object |
| 500 | Failed to get disputes | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/disputes/resolve

#### POST
##### Summary:

Resolve dispute

##### Description:

Allows an assigned admin to resolve a dispute and release funds from escrow

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body | Resolution payload | Yes | [handlers.ResolveDisputeRequest](#handlers.ResolveDisputeRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Success flag and message | object |
| 400 | Invalid JSON or invalid resolution | object |
| 401 | Unauthorized | object |
| 403 | Dispute not assigned to you | object |
| 404 | Dispute or task not found | object |
| 500 | Failed to resolve dispute | object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/offers

#### POST
##### Summary:

Create a task offer

##### Description:

Allows a freelancer to make an offer on an open task

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body | Offer payload | Yes | [handlers.CreateTaskOfferRequest](#handlers.CreateTaskOfferRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | success flag and created offer | object |
| 400 | Invalid JSON or bad request | string |
| 401 | Unauthorized | string |
| 404 | Task not found | string |
| 500 | Failed to create offer | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/offers/delete

#### DELETE
##### Summary:

Delete user's own task offer

##### Description:

Allows a freelancer (or admin) to delete their own offer if not accepted

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | query | Offer ID | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | success | object |
| 400 | Invalid offer ID or accepted offer cannot be deleted | object |
| 401 | Unauthorized | object |
| 403 | Forbidden (not owner or admin) | object |
| 404 | Offer not found | object |
| 405 | Method not allowed | object |
| 500 | Failed to delete offer | object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/offers/update

#### PUT
##### Summary:

Update user's own task offer

##### Description:

Allows a freelancer to update their own offer (not accepted yet)

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| offer | body | Offer data | Yes | [models.TaskOffer](#models.TaskOffer) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | success and updated offer | object |
| 400 | Invalid JSON or accepted offer cannot be edited | object |
| 401 | Unauthorized | object |
| 403 | Forbidden (not owner) | object |
| 404 | Offer not found | object |
| 405 | Method not allowed | object |
| 500 | Failed to update offer | object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/reviews

#### POST
##### Summary:

Create a review

##### Description:

Create a review for a completed task. Only the client or the accepted freelancer can review.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| review | body | Review payload | Yes | [models.Review](#models.Review) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [models.Review](#models.Review) |
| 400 | Bad request | object |
| 401 | Unauthorized | object |
| 403 | Forbidden | object |
| 404 | Task not found | object |
| 500 | Internal server error | object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/reviews/by-task

#### GET
##### Summary:

Get reviews by task

##### Description:

Returns all reviews for a specific task

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| task_id | query | Task ID | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [ [models.Review](#models.Review) ] |
| 400 | Invalid task ID | object |
| 500 | Internal server error | object |

### /api/reviews/rating

#### GET
##### Summary:

Get user rating

##### Description:

Returns the average rating of a user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| user_id | query | User ID | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Rating | object |
| 400 | Invalid user ID | object |
| 500 | Internal server error | object |

### /api/tasks

#### DELETE
##### Summary:

Delete a task

##### Description:

Allows the task owner to delete a task

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | query | Task ID | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | success flag | object |
| 400 | Invalid task ID | string |
| 401 | Unauthorized | string |
| 403 | Forbidden | string |
| 404 | Task not found | string |
| 500 | Failed to delete task | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

#### GET
##### Summary:

Get tasks

##### Description:

Returns list of tasks. Use query param `status=open` to get only open tasks

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| status | query | Filter tasks by status | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | success flag and tasks list | object |
| 401 | Unauthorized | string |
| 500 | Failed to get tasks | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

#### POST
##### Summary:

Create a new task

##### Description:

Allows a user to create a new task

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body | Task payload | Yes | [handlers.CreateTaskRequest](#handlers.CreateTaskRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | success flag and created task | object |
| 400 | Invalid JSON | string |
| 401 | Unauthorized | string |
| 500 | Failed to create task | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

#### PUT
##### Summary:

Update a task

##### Description:

Allows the task owner to update a task

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body | Updated task payload | Yes | [handlers.UpdateTaskRequest](#handlers.UpdateTaskRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | success flag and updated task | object |
| 400 | Invalid JSON | string |
| 401 | Unauthorized | string |
| 403 | Forbidden | string |
| 404 | Task not found | string |
| 500 | Failed to update task | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/tasks/detail

#### GET
##### Summary:

Get task details

##### Description:

Returns details of a single task

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | query | Task ID | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | success flag and task | object |
| 400 | Invalid task ID | string |
| 404 | Task not found | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/ticket/createTicket

#### POST
##### Summary:

Create new ticket

##### Description:

Create new ticket

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| request | body | Ticket info | Yes | [handlers.TicketCreateRequest](#handlers.TicketCreateRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [handlers.TicketCreateAnswer](#handlers.TicketCreateAnswer) |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| bearerAuth | |

### /api/ticket/exit

#### POST
##### Summary:

Exit from a ticket

##### Description:

Removes the user from the ticket's participants

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| request | body | Ticket ID | Yes | [handlers.TicketIDRequest](#handlers.TicketIDRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | status: ok | object |
| 400 | Invalid payload or ticket_id | object |
| 401 | User not authenticated | object |
| 500 | Internal server error | object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| bearerAuth | |

### /api/ticket/messages

#### GET
##### Summary:

Get messages for a ticket

##### Description:

Returns messages for a given ticket if the user has access, supports limit/offset

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| ticket_id | query | Ticket ID | Yes | integer |
| limit | query | Number of messages to return (max 1000, default 100) | No | integer |
| offset | query | Offset for messages (default last messages) | No | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | List of messages | [ [models.TicketMessage](#models.TicketMessage) ] |
| 400 | Invalid parameters | object |
| 401 | User not authenticated | object |
| 403 | User does not have access | object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| bearerAuth | |

### /api/ticket/my

#### GET
##### Summary:

Get own tickets

##### Description:

Get all tickets of user

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [ [models.TicketDoc](#models.TicketDoc) ] |
| 401 | Unauthorized | object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| bearerAuth | |

### /api/ticket/write

#### POST
##### Summary:

Write to ticket

##### Description:

Add message to ticket

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| request | body | Message info | Yes | [handlers.WriteTicketRequest](#handlers.WriteTicketRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |
| 403 | Forbidden | object |
| 404 | Not Found | object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| bearerAuth | |

### /api/wallet

#### GET
##### Summary:

Get wallet balances

##### Description:

Returns user’s balances in BTC and XMR

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [db.WalletBalance](#db.WalletBalance) |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/wallet/bitcoinSend

#### POST
##### Summary:

Send Bitcoin

##### Description:

Sends Bitcoin transaction using Electrum

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| to | query | Destination address | Yes | string |
| amount | query | Amount | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /api/wallet/moneroSend

#### POST
##### Summary:

Send Monero

##### Description:

Sends Monero transaction (not implemented)

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /auth

#### POST
##### Summary:

Authenticate user

##### Description:

Logs in user and returns JWT token

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| request | body | Login credentials | Yes | [handlers.AuthRequest](#handlers.AuthRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [handlers.AuthResponse](#handlers.AuthResponse) |
| 401 | Unauthorized | object |

### /captcha

#### GET
##### Summary:

Get captcha image

##### Description:

Returns a captcha image and X-Captcha-ID header

##### Responses

| Code | Description |
| ---- | ----------- |
| 200 | image/png |

### /chat/UpdateChatRequest

#### POST
##### Summary:

Update a chat request

##### Description:

Accept or reject a chat request by the requested user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| requester_id | query | ID of the user who sent the chat request | Yes | integer |
| status | query | New status: accepted or rejected | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Returns status ok | object |
| 400 | Invalid requester_id or status | string |
| 401 | Unauthorized — user not logged in or not allowed to update | string |
| 500 | Database error | string |

### /chat/acceptChatRequest

#### POST
##### Summary:

Accept chat request

##### Description:

Accepts a chat request from another user and creates a chat room

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| requester_id | query | ID of the user who sent the request | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Returns accepted status and chatRoomID | object |
| 400 | Invalid requester_id | string |
| 401 | Unauthorized — user not logged in | string |
| 500 | Database error | string |

### /chat/cancelChatRequest

#### POST
##### Summary:

Cancel chat request

##### Description:

Cancels a previously sent chat request

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| requester_id | query | ID of the user to whom the request was sent | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Request cancelled successfully | string |
| 400 | Invalid requester_id | string |
| 401 | Unauthorized — user not logged in | string |
| 500 | Database error | string |

### /chat/createChatRequest

#### POST
##### Summary:

Create a chat request

##### Description:

Create a new chat request from the logged-in user to another user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| requested_id | query | ID of the user you want to start a chat with | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 201 | Returns the created chat request | [models.ChatRequest](#models.ChatRequest) |
| 400 | invalid requested_id | string |
| 401 | Unauthorized — user not logged in | string |
| 500 | db error | string |

### /chat/exitFromChat

#### POST
##### Summary:

Exit chat room

##### Description:

Removes the logged-in user from a chat room

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| chat_room_id | query | ID of the chat room | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | Returns status ok | object |
| 400 | Invalid chat_room_id | string |
| 401 | Unauthorized — user not logged in | string |
| 500 | Database error | string |

### /chat/getChatMessages

#### GET
##### Summary:

Get chat messages

##### Description:

Returns all messages for a given chat room if the user has access

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| chat_room_id | query | ID of the chat room | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | List of messages | [ [models.ChatMessage](#models.ChatMessage) ] |
| 400 | Invalid chat_room_id | string |
| 401 | Unauthorized — user not logged in or no access | string |
| 500 | Database error | string |

### /chat/getChatRequests

#### GET
##### Summary:

Get chat requests

##### Description:

Returns all incoming chat requests for the logged-in user

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | List of chat requests | [ [models.ChatRequest](#models.ChatRequest) ] |
| 401 | Unauthorized — user not logged in | string |
| 500 | Database error | string |

### /chat/getChatRoomsForUser

#### GET
##### Summary:

Get chat rooms

##### Description:

Returns all chat rooms the logged-in user participates in

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | List of chat rooms | [ [models.ChatRoom](#models.ChatRoom) ] |
| 401 | Unauthorized — user not logged in | string |
| 500 | Database error | string |

### /chat/sendMessage

#### POST
##### Summary:

Send message

##### Description:

Sends a message to a chat room for the logged-in user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| chat_room_id | query | ID of the chat room | Yes | integer |
| message | body | Message object | Yes | [models.ChatMessage](#models.ChatMessage) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 201 | Returns the created message | [models.ChatMessage](#models.ChatMessage) |
| 400 | Invalid chat_room_id or request body | string |
| 401 | Unauthorized — user not logged in | string |
| 500 | Database error | string |

### /hello

#### GET
##### Summary:

Health/hello

##### Description:

Simple hello endpoint

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [handlers.Response](#handlers.Response) |

### /profile

#### GET
##### Summary:

Get or update profile

##### Description:

Get current user profile (GET) or update profile (POST)

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [models.Profile](#models.Profile) |
| 400 | invalid payload | string |
| 401 | unauthorized | string |
| 500 | db error | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

#### POST
##### Summary:

Get or update profile

##### Description:

Get current user profile (GET) or update profile (POST)

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [models.Profile](#models.Profile) |
| 400 | invalid payload | string |
| 401 | unauthorized | string |
| 500 | db error | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /profile/by_id

#### GET
##### Summary:

Get public profile by user_id

##### Description:

Returns sanitized profile and username by user_id

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| user_id | query | User ID | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [models.Profile](#models.Profile) |
| 400 | invalid user_id | string |
| 404 | not found | string |

### /profiles

#### GET
##### Summary:

List profiles

##### Description:

Returns paginated list of profiles

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| limit | query | Limit | No | integer |
| offset | query | Offset | No | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [ [models.Profile](#models.Profile) ] |
| 500 | db error | string |

##### Security

| Security Schema | Scopes |
| --- | --- |
| BearerAuth | |

### /register

#### POST
##### Summary:

Register new user

##### Description:

Creates a new user with login, password and captcha

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| request | body | User credentials | Yes | [handlers.RegisterRequest](#handlers.RegisterRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [handlers.Response](#handlers.Response) |
| 400 | Bad Request | object |

### /restoreuser

#### POST
##### Summary:

Restore user account

##### Description:

Restore account by mnemonic and set new password

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| request | body | Restore payload | Yes | [handlers.RestoreRequest](#handlers.RestoreRequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | [handlers.Response](#handlers.Response) |
| 400 | Bad Request | object |

### /verify

#### GET
##### Summary:

Verify captcha

##### Description:

Verifies provided captcha answer

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | query | Captcha ID | Yes | string |
| answer | query | Captcha answer | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |

### Models


#### db.WalletBalance

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| address | string |  | No |
| balance | number |  | No |

#### handlers.AdminRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user_id | integer |  | No |

#### handlers.AdminTransactionsRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| limit | integer |  | No |
| offset | integer |  | No |
| wallet_id | integer |  | No |

#### handlers.AdminUpdateBalanceRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| balance | string |  | No |
| user_id | integer |  | No |

#### handlers.AssignDisputeRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| dispute_id | integer |  | No |

#### handlers.AuthRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| captcha_answer | string |  | No |
| captcha_id | string |  | No |
| password | string |  | No |
| username | string |  | No |

#### handlers.AuthResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string |  | No |
| token | string |  | No |

#### handlers.CreateTaskOfferRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| price | number |  | No |
| task_id | integer |  | No |

#### handlers.CreateTaskRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| currency | string |  | No |
| deadline | string | ISO8601 | No |
| description | string |  | No |
| price | number |  | No |
| title | string |  | No |

#### handlers.RegisterRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| captcha_answer | string |  | No |
| captcha_id | string |  | No |
| password | string |  | No |
| username | string |  | No |

#### handlers.ResolveDisputeRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| dispute_id | integer |  | No |
| resolution | string | "client_won" или "freelancer_won" | No |

#### handlers.Response

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| encrypted | string |  | No |
| message | string |  | No |

#### handlers.RestoreRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| captcha_answer | string |  | No |
| captcha_id | string |  | No |
| mnemonic | string |  | No |
| new_password | string |  | No |
| username | string |  | No |

#### handlers.SendDisputeMessageRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| dispute_id | integer |  | No |
| message | string |  | No |

#### handlers.TicketCreateAnswer

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| ticket_id | integer |  | No |

#### handlers.TicketCreateRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string |  | No |
| subject | string |  | No |

#### handlers.TicketIDRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| ticket_id | integer |  | No |

#### handlers.UpdateTaskRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| currency | string |  | No |
| deadline | string |  | No |
| description | string |  | No |
| id | integer |  | No |
| price | number |  | No |
| title | string |  | No |

#### handlers.WriteTicketRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string |  | No |
| ticket_id | integer |  | No |

#### models.ChatMessage

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| chat_room_id | integer |  | No |
| created_at | string |  | No |
| id | integer |  | No |
| message | string |  | No |
| sender_id | integer |  | No |

#### models.ChatRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| created_at | string |  | No |
| id | integer |  | No |
| requested_id | integer |  | No |
| requester_id | integer |  | No |
| status | string | pending, accepted, rejected | No |

#### models.ChatRoom

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| created_at | string |  | No |
| id | integer |  | No |

#### models.Profile

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| avatar | string |  | No |
| bio | string |  | No |
| completed_tasks | integer |  | No |
| full_name | string |  | No |
| rating | number |  | No |
| skills | [ string ] |  | No |
| user_id | integer |  | No |

#### models.Review

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| comment | string |  | No |
| created_at | string |  | No |
| id | integer |  | No |
| rating | integer |  | No |
| reviewed_id | integer |  | No |
| reviewer_id | integer |  | No |
| task_id | integer |  | No |

#### models.TaskOffer

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| accepted | boolean |  | No |
| created_at | string |  | No |
| freelancer_id | integer |  | No |
| id | integer |  | No |
| message | string |  | No |
| price | number |  | No |
| task_id | integer |  | No |

#### models.TicketDoc

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| additional_users_have_access | [ integer ] |  | No |
| admin_id | integer |  | No |
| created_at | string |  | No |
| id | integer |  | No |
| status | string |  | No |
| subject | string |  | No |
| updated_at | string |  | No |
| user_id | integer |  | No |

#### models.TicketMessage

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| createdAt | string |  | No |
| id | integer |  | No |
| message | string |  | No |
| read | boolean |  | No |
| senderID | integer |  | No |
| ticketID | integer |  | No |

#### models.Wallet

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| address | string |  | No |
| balance | string |  | No |
| currency | string |  | No |
| id | integer |  | No |
| userID | integer |  | No |
