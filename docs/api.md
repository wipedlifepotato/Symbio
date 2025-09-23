# mFrelance API Documentation

## Version: 1.0

### Overview
mFrelance is a freelance platform API that provides user registration, authentication, task management, reviews, disputes, chat functionality, and administrative controls. The API supports cryptocurrency payments (Bitcoin and Monero) and includes comprehensive security features.

### Base URL
```
/api/
```

### Authentication
All protected endpoints require JWT Bearer token authentication.

**Header Format:**
```
Authorization: Bearer <jwt_token>
```

### Response Format
All responses follow a consistent JSON structure:

**Success Response:**
```json
{
  "success": true,
  "data": { ... },
  "message": "Optional success message"
}
```

**Error Response:**
```json
{
  "error": "Error message",
  "code": "Optional error code"
}
```

---

## Authentication Endpoints

### POST /register
Register a new user account.

**Request Body:**
```json
{
  "username": "string (2-128 chars)",
  "password": "string (6-128 chars)",
  "captcha_id": "string",
  "captcha_answer": "string"
}
```

**Success Response (200):**
```json
{
  "message": "Account created successfully. Save your recovery phrase!",
  "encrypted": "word1 word2 word3 word4 word5 word6 word7 word8 word9 word10 word11 word12"
}
```

**Error Responses:**
- `400`: Invalid input, CAPTCHA failure, or password requirements not met
- `500`: Internal server error during account creation

### POST /auth
Authenticate user and get JWT token.

**Request Body:**
```json
{
  "username": "string",
  "password": "string",
  "captcha_id": "string",
  "captcha_answer": "string"
}
```

**Success Response (200):**
```json
{
  "message": "Authenticated successfully",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Error Responses:**
- `400`: Invalid CAPTCHA
- `401`: Invalid username or password

### POST /restoreuser
Restore user account using recovery phrase.

**Request Body:**
```json
{
  "username": "string",
  "mnemonic": "word1 word2 word3...",
  "new_password": "string (6-128 chars)",
  "captcha_id": "string",
  "captcha_answer": "string"
}
```

**Success Response (200):**
```json
{
  "message": "Account restored successfully.",
  "encrypted": "new_jwt_token"
}
```

**Error Responses:**
- `400`: Invalid input or CAPTCHA
- `500`: Failed to restore account

### GET /captcha
Generate CAPTCHA image.

**Response:** PNG image with `X-Captcha-ID` header.

**Headers:**
```
X-Captcha-ID: captcha_unique_id
Content-Type: image/png
```

### GET /captcha/status
Check if CAPTCHA is enabled.

**Success Response (200):**
```json
{
  "enabled": true
}
```

### GET /verify
Verify CAPTCHA answer.

**Query Parameters:**
- `id`: CAPTCHA ID
- `answer`: User's answer

**Success Response (200):**
```json
{
  "ok": true
}
```

---

## Task Management

### POST /tasks
Create a new task.

**Request Body:**
```json
{
  "title": "string",
  "description": "string",
  "price": 100.50,
  "currency": "BTC",
  "deadline": "2023-12-31T23:59:59Z"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "task": {
    "id": 123,
    "title": "Website Design",
    "description": "Need a modern website",
    "price": 100.5,
    "currency": "BTC",
    "deadline": "2023-12-31T23:59:59Z",
    "client_id": 456,
    "status": "open",
    "created_at": "2023-12-01T10:00:00Z"
  }
}
```

**Error Responses:**
- `400`: Invalid JSON
- `401`: Unauthorized
- `500`: Failed to create task

### GET /tasks
Get user's tasks with optional filtering.

**Query Parameters:**
- `status`: Filter by status (`open`, `in_progress`, `completed`)
- `limit`: Number of results (default: 20, max: 100)
- `offset`: Pagination offset (default: 0)

**Success Response (200):**
```json
{
  "success": true,
  "tasks": [
    {
      "id": 123,
      "title": "Website Design",
      "description": "Need a modern website",
      "price": 100.5,
      "currency": "BTC",
      "deadline": "2023-12-31T23:59:59Z",
      "client_id": 456,
      "status": "open",
      "created_at": "2023-12-01T10:00:00Z"
    }
  ]
}
```

### PUT /tasks
Update an existing task (task owner only).

**Request Body:**
```json
{
  "id": 123,
  "title": "Updated Title",
  "description": "Updated description",
  "price": 150.0,
  "currency": "BTC",
  "deadline": "2023-12-31T23:59:59Z"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "task": { ... }
}
```

### DELETE /tasks
Delete a task (task owner or admin only).

**Query Parameters:**
- `id`: Task ID

**Success Response (200):**
```json
{
  "success": true
}
```

### GET /tasks/detail
Get detailed information about a specific task.

**Query Parameters:**
- `id`: Task ID

**Success Response (200):**
```json
{
  "success": true,
  "task": {
    "id": 123,
    "title": "Website Design",
    "description": "Need a modern website",
    "price": 100.5,
    "currency": "BTC",
    "deadline": "2023-12-31T23:59:59Z",
    "client_id": 456,
    "status": "open",
    "created_at": "2023-12-01T10:00:00Z"
  }
}
```

### POST /tasks/complete
Mark a task as completed and release escrow funds (client only).

**Request Body:**
```json
{
  "task_id": 123
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Task confirmed by client and completed successfully"
}
```

---

## Task Offers

### POST /offers
Create a task offer (freelancer only).

**Request Body:**
```json
{
  "task_id": 123,
  "price": 95.0
}
```

**Success Response (200):**
```json
{
  "success": true,
  "offer": {
    "id": 456,
    "task_id": 123,
    "freelancer_id": 789,
    "price": 95.0,
    "message": "",
    "accepted": false,
    "created_at": "2023-12-01T10:00:00Z"
  }
}
```

### GET /offers
Get offers for a specific task.

**Query Parameters:**
- `task_id`: Task ID

**Success Response (200):**
```json
{
  "success": true,
  "offers": [
    {
      "id": 456,
      "task_id": 123,
      "freelancer_id": 789,
      "price": 95.0,
      "message": "I can do this quickly",
      "accepted": false,
      "created_at": "2023-12-01T10:00:00Z"
    }
  ]
}
```

### PUT /offers/update
Update user's own offer (if not accepted).

**Request Body:**
```json
{
  "id": 456,
  "price": 90.0,
  "message": "Updated offer"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "offer": { ... }
}
```

### DELETE /offers/delete
Delete user's own offer (if not accepted).

**Query Parameters:**
- `id`: Offer ID

**Success Response (200):**
```json
{
  "success": true
}
```

### POST /offers/accept
Accept a freelancer's offer (task owner only).

**Request Body:**
```json
{
  "offer_id": 456
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Offer accepted successfully"
}
```

---

## Reviews

### POST /reviews
Submit a review for a completed task.

**Request Body:**
```json
{
  "task_id": 123,
  "rating": 5,
  "comment": "Excellent work!"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "review": {
    "id": 789,
    "task_id": 123,
    "reviewer_id": 456,
    "reviewed_id": 789,
    "rating": 5,
    "comment": "Excellent work!",
    "created_at": "2023-12-01T10:00:00Z"
  }
}
```

### GET /reviews/user
Get all reviews received by a user.

**Query Parameters:**
- `user_id`: User ID

**Success Response (200):**
```json
{
  "success": true,
  "reviews": [
    {
      "id": 789,
      "task_id": 123,
      "reviewer_id": 456,
      "rating": 5,
      "comment": "Excellent work!",
      "created_at": "2023-12-01T10:00:00Z"
    }
  ]
}
```

### GET /reviews/task
Get all reviews for a specific task.

**Query Parameters:**
- `task_id`: Task ID

**Success Response (200):**
```json
{
  "success": true,
  "reviews": [
    {
      "id": 789,
      "reviewer_id": 456,
      "rating": 5,
      "comment": "Excellent work!",
      "created_at": "2023-12-01T10:00:00Z"
    }
  ]
}
```

### GET /reviews/rating
Get user's average rating.

**Query Parameters:**
- `user_id`: User ID

**Success Response (200):**
```json
{
  "success": true,
  "rating": 4.5
}
```

---

## Disputes

### POST /disputes/create
Create a new dispute for a task.

**Request Body:**
```json
{
  "task_id": 123,
  "reason": "Work not completed as agreed"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Dispute created successfully"
}
```

### GET /disputes/my
Get user's disputes.

**Success Response (200):**
```json
{
  "success": true,
  "disputes": [
    {
      "id": 123,
      "task_id": 456,
      "client_id": 789,
      "freelancer_id": 101,
      "reason": "Work not completed",
      "status": "open",
      "admin_id": null,
      "resolution": null,
      "created_at": "2023-12-01T10:00:00Z"
    }
  ]
}
```

### GET /disputes/details
Get detailed dispute information.

**Query Parameters:**
- `id`: Dispute ID

**Success Response (200):**
```json
{
  "success": true,
  "dispute": { ... },
  "task": { ... },
  "escrow": { ... },
  "messages": [ ... ]
}
```

### POST /disputes/message
Send a message in a dispute.

**Request Body:**
```json
{
  "dispute_id": 123,
  "message": "Please provide evidence of completed work"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Message sent"
}
```

### POST /disputes/resolve
Resolve a dispute (admin only).

**Request Body:**
```json
{
  "dispute_id": 123,
  "resolution": "client_won"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Dispute resolved"
}
```

---

## Chat System

### POST /chat/createChatRequest
Create a chat request to another user.

**Query Parameters:**
- `requested_id`: User ID to chat with

**Success Response (201):**
```json
{
  "id": 123,
  "requester_id": 456,
  "requested_id": 789,
  "status": "pending",
  "created_at": "2023-12-01T10:00:00Z"
}
```

### POST /chat/UpdateChatRequest
Accept or reject a chat request.

**Query Parameters:**
- `requester_id`: User who sent the request
- `status`: "accepted" or "rejected"

**Success Response (200):**
```json
{
  "status": "ok"
}
```

### GET /chat/getChatRequests
Get incoming chat requests.

**Success Response (200):**
```json
{
  "chat_requests": [
    {
      "id": 123,
      "requester_id": 456,
      "requested_id": 789,
      "status": "pending",
      "created_at": "2023-12-01T10:00:00Z"
    }
  ]
}
```

### GET /chat/getChatRoomsForUser
Get user's chat rooms.

**Success Response (200):**
```json
{
  "chat_rooms": [
    {
      "id": 123,
      "created_at": "2023-12-01T10:00:00Z"
    }
  ]
}
```

### GET /chat/getChatMessages
Get messages from a chat room.

**Query Parameters:**
- `chat_room_id`: Chat room ID

**Success Response (200):**
```json
{
  "messages": [
    {
      "id": 123,
      "chat_room_id": 456,
      "sender_id": 789,
      "message": "Hello!",
      "created_at": "2023-12-01T10:00:00Z"
    }
  ]
}
```

### POST /chat/sendMessage
Send a message to a chat room.

**Query Parameters:**
- `chat_room_id`: Chat room ID

**Request Body:**
```json
{
  "message": "Hello, how are you?"
}
```

**Success Response (201):**
```json
{
  "id": 124,
  "chat_room_id": 456,
  "sender_id": 789,
  "message": "Hello, how are you?",
  "created_at": "2023-12-01T10:00:00Z"
}
```

---

## Profile Management

### GET /profile
Get current user's profile.

**Success Response (200):**
```json
{
  "user_id": 123,
  "full_name": "John Doe",
  "bio": "Experienced web developer",
  "skills": ["JavaScript", "React", "Node.js"],
  "avatar": "avatar_url",
  "rating": 4.5,
  "completed_tasks": 25
}
```

### POST /profile
Update current user's profile.

**Request Body:**
```json
{
  "full_name": "John Doe",
  "bio": "Experienced web developer",
  "skills": ["JavaScript", "React", "Node.js"],
  "avatar": "avatar_url"
}
```

**Success Response (200):**
```json
{
  "user_id": 123,
  "full_name": "John Doe",
  "bio": "Experienced web developer",
  "skills": ["JavaScript", "React", "Node.js"],
  "avatar": "avatar_url",
  "rating": 4.5,
  "completed_tasks": 25
}
```

### GET /profile/by_id
Get public profile by user ID.

**Query Parameters:**
- `user_id`: User ID

**Success Response (200):**
```json
{
  "user_id": 123,
  "username": "johndoe",
  "full_name": "John Doe",
  "bio": "Experienced web developer",
  "skills": ["JavaScript", "React", "Node.js"],
  "avatar": "avatar_url",
  "rating": 4.5,
  "completed_tasks": 25
}
```

### GET /profiles
List user profiles with pagination.

**Query Parameters:**
- `limit`: Number of results (default: 20)
- `offset`: Pagination offset (default: 0)

**Success Response (200):**
```json
{
  "profiles": [
    {
      "user_id": 123,
      "username": "johndoe",
      "full_name": "John Doe",
      "bio": "Experienced web developer",
      "skills": ["JavaScript", "React", "Node.js"],
      "avatar": "avatar_url",
      "rating": 4.5,
      "completed_tasks": 25
    }
  ]
}
```

---

## Wallet Operations

### GET /wallet
Get user's wallet balances.

**Success Response (200):**
```json
{
  "BTC": {
    "address": "1ABC...",
    "balance": "0.5"
  },
  "XMR": {
    "address": "4ABC...",
    "balance": "10.0"
  }
}
```

### POST /wallet/bitcoinSend
Send Bitcoin transaction.

**Query Parameters:**
- `to`: Destination address
- `amount`: Amount to send

**Success Response (200):**
```json
{
  "txid": "a1b2c3d4...",
  "status": "sent"
}
```

### POST /wallet/moneroSend
Send Monero transaction.

**Query Parameters:**
- `to`: Destination address
- `amount`: Amount to send

**Success Response (200):**
```json
{
  "txid": "x1y2z3w4...",
  "status": "sent"
}
```

---

## Support Tickets

### POST /ticket/createTicket
Create a new support ticket.

**Request Body:**
```json
{
  "subject": "Login Issue",
  "message": "I cannot log in to my account"
}
```

**Success Response (200):**
```json
{
  "ticket_id": 123
}
```

### GET /ticket/my
Get user's tickets.

**Success Response (200):**
```json
{
  "tickets": [
    {
      "id": 123,
      "user_id": 456,
      "admin_id": 789,
      "subject": "Login Issue",
      "status": "open",
      "created_at": "2023-12-01T10:00:00Z",
      "updated_at": "2023-12-01T10:00:00Z"
    }
  ]
}
```

### GET /ticket/messages
Get messages for a ticket.

**Query Parameters:**
- `ticket_id`: Ticket ID
- `limit`: Number of messages (default: 100, max: 1000)
- `offset`: Offset for pagination

**Success Response (200):**
```json
{
  "messages": [
    {
      "id": 123,
      "ticketID": 456,
      "senderID": 789,
      "message": "I cannot log in",
      "read": true,
      "createdAt": "2023-12-01T10:00:00Z"
    }
  ]
}
```

### POST /ticket/write
Add message to ticket.

**Request Body:**
```json
{
  "ticket_id": 123,
  "message": "Please help me reset my password"
}
```

**Success Response (200):**
```json
{
  "status": "ok"
}
```

### POST /ticket/exit
Remove user from ticket participants.

**Request Body:**
```json
{
  "ticket_id": 123
}
```

**Success Response (200):**
```json
{
  "status": "ok"
}
```

### POST /ticket/close
Close a ticket (admin only).

**Request Body:**
```json
{
  "ticket_id": 123
}
```

**Success Response (200):**
```json
{
  "status": "ok"
}
```

---

## Administrative Endpoints

### POST /admin/make
Grant admin privileges to a user.

**Request Body:**
```json
{
  "user_id": 123
}
```

**Success Response (200):**
```json
"message": "user is now admin"
```

### POST /admin/remove
Revoke admin privileges from a user.

**Request Body:**
```json
{
  "user_id": 123
}
```

**Success Response (200):**
```json
"message": "user admin removed"
```

### GET /admin/check
Check if current user is admin.

**Success Response (200):**
```json
{
  "user_id": 123,
  "is_admin": true
}
```

### GET /admin/IIsAdmin
Check if current user is admin (alternative endpoint).

**Success Response (200):**
```json
{
  "user_id": 123,
  "is_admin": true
}
```

### POST /admin/block
Block a user account.

**Request Body:**
```json
{
  "user_id": 123
}
```

**Success Response (200):**
```json
"message": "user blocked"
```

### POST /admin/unblock
Unblock a user account.

**Request Body:**
```json
{
  "user_id": 123
}
```

**Success Response (200):**
```json
"message": "user unblocked"
```

### POST /admin/transactions
Get transactions (admin view).

**Request Body:**
```json
{
  "wallet_id": 123,
  "limit": 50,
  "offset": 0
}
```

**Success Response (200):**
```json
[
  {
    "id": 123,
    "from_wallet_id": 456,
    "to_wallet_id": 789,
    "to_address": "1ABC...",
    "task_id": 101,
    "amount": "0.5",
    "currency": "BTC",
    "confirmed": true,
    "created_at": "2023-12-01T10:00:00Z"
  }
]
```

### GET /admin/wallets
Get user wallets (admin view).

**Query Parameters:**
- `user_id`: User ID

**Success Response (200):**
```json
[
  {
    "id": 123,
    "userID": 456,
    "currency": "BTC",
    "address": "1ABC...",
    "balance": "0.5"
  }
]
```

### POST /admin/update_balance
Update wallet balance (admin only).

**Request Body:**
```json
{
  "user_id": 123,
  "balance": "1.0"
}
```

**Success Response (200):**
```json
"message": "balance updated"
```

### GET /admin/getRandomTicket
Assign random open ticket to admin.

**Success Response (200):**
```json
{
  "id": 123,
  "user_id": 456,
  "admin_id": 789,
  "subject": "Login Issue",
  "status": "open",
  "created_at": "2023-12-01T10:00:00Z",
  "updated_at": "2023-12-01T10:00:00Z"
}
```

### POST /admin/delete_user_tasks
Delete all tasks belonging to a user.

**Query Parameters:**
- `user_id`: User ID

**Success Response (200):**
```json
{
  "success": true,
  "deleted": 5
}
```

### GET /admin/disputes
Get all open disputes (admin view).

**Success Response (200):**
```json
{
  "success": true,
  "disputes": [
    {
      "id": 123,
      "task_id": 456,
      "client_id": 789,
      "freelancer_id": 101,
      "reason": "Work not completed",
      "status": "open",
      "admin_id": null,
      "resolution": null,
      "created_at": "2023-12-01T10:00:00Z"
    }
  ]
}
```

### POST /admin/disputes/assign
Assign dispute to current admin.

**Request Body:**
```json
{
  "dispute_id": 123
}
```

**Success Response (200):**
```json
{
  "success": true
}
```

### POST /admin/disputes/resolve
Resolve assigned dispute.

**Request Body:**
```json
{
  "dispute_id": 123,
  "resolution": "client_won"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Dispute resolved"
}
```

---

## System Endpoints

### GET /hello
Health check endpoint.

**Success Response (200):**
```json
{
  "message": "Hello, REST API!"
}
```

### GET /ownID
Get current user's ID.

**Success Response (200):**
```json
{
  "user_id": 123
}
```

---

## Error Codes

### Common HTTP Status Codes
- `200`: Success
- `400`: Bad Request (invalid input)
- `401`: Unauthorized (invalid/missing JWT)
- `403`: Forbidden (insufficient permissions)
- `404`: Not Found
- `429`: Too Many Requests (rate limited)
- `500`: Internal Server Error

### Common Error Responses
```json
{
  "error": "Invalid JSON"
}
```

```json
{
  "error": "Unauthorized"
}
```

```json
{
  "error": "Rate limit exceeded. Please try again later."
}
```

---

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **Task Creation**: Maximum 1 task per hour per user
- **CAPTCHA Generation**: Maximum 10 per minute, 100 per hour per IP
- **General Requests**: Reasonable limits applied to prevent DoS attacks

When rate limited, the API returns HTTP 429 with an appropriate error message.

---

## Security Features

- **JWT Authentication**: Bearer token required for protected endpoints
- **CAPTCHA Protection**: Required for registration, login, and password recovery
- **Rate Limiting**: Prevents abuse and DoS attacks
- **Input Validation**: All inputs are validated and sanitized
- **SQL Injection Protection**: Parameterized queries used throughout
- **XSS Protection**: Proper output encoding
- **User Blocking**: Administrators can block malicious users

---

## Data Models

### User
```json
{
  "id": 123,
  "username": "johndoe",
  "created_at": "2023-12-01T10:00:00Z"
}
```

### Task
```json
{
  "id": 123,
  "title": "Website Design",
  "description": "Need a modern website",
  "price": 100.5,
  "currency": "BTC",
  "deadline": "2023-12-31T23:59:59Z",
  "client_id": 456,
  "status": "open",
  "created_at": "2023-12-01T10:00:00Z"
}
```

### TaskOffer
```json
{
  "id": 456,
  "task_id": 123,
  "freelancer_id": 789,
  "price": 95.0,
  "message": "I can do this quickly",
  "accepted": false,
  "created_at": "2023-12-01T10:00:00Z"
}
```

### Review
```json
{
  "id": 789,
  "task_id": 123,
  "reviewer_id": 456,
  "reviewed_id": 789,
  "rating": 5,
  "comment": "Excellent work!",
  "created_at": "2023-12-01T10:00:00Z"
}
```

### Dispute
```json
{
  "id": 123,
  "task_id": 456,
  "client_id": 789,
  "freelancer_id": 101,
  "reason": "Work not completed as agreed",
  "status": "open",
  "admin_id": null,
  "resolution": null,
  "created_at": "2023-12-01T10:00:00Z"
}
```

### Wallet
```json
{
  "id": 123,
  "userID": 456,
  "currency": "BTC",
  "address": "1ABC...",
  "balance": "0.5"
}
```

---

## Getting Started

1. **Register** a new account at `POST /register`
2. **Authenticate** to get JWT token at `POST /auth`
3. **Create tasks** at `POST /tasks`
4. **Browse and offer** on tasks at `GET /tasks` and `POST /offers`
5. **Complete workflow** through offers, completion, and reviews

For frontend integration, always check `GET /captcha/status` to conditionally show CAPTCHA fields.
