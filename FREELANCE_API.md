# API системы фриланса mFrelance

## Обзор
Система фриланса позволяет пользователям создавать задачи, подавать предложения, работать с диспутами и оставлять отзывы.

## Аутентификация
Все API эндпоинты требуют JWT токен в заголовке `Authorization: Bearer <token>`

## Управление задачами

### Создание задачи
```
POST /api/tasks/create
```
**Тело запроса:**
```json
{
  "title": "Название задачи",
  "description": "Описание задачи",
  "category": "web_development",
  "budget": 0.001,
  "currency": "BTC",
  "deadline": "2024-12-31T23:59:59Z"
}
```

### Получение списка задач
```
GET /api/tasks?status=open
```
**Параметры:**
- `status` (опционально): open, in_progress, completed, cancelled, disputed

### Получение конкретной задачи
```
GET /api/tasks/get?id=123
```

### Обновление задачи
```
PUT /api/tasks/update
```

### Удаление задачи
```
DELETE /api/tasks/delete?id=123
```

## Система предложений

### Создание предложения
```
POST /api/offers/create
```
**Тело запроса:**
```json
{
  "task_id": 123,
  "price": 0.0008,
  "message": "Мое предложение по выполнению задачи"
}
```

### Получение предложений по задаче
```
GET /api/offers?task_id=123
```

### Принятие предложения
```
POST /api/offers/accept
```
**Тело запроса:**
```json
{
  "offer_id": 456
}
```

### Завершение задачи
```
POST /api/tasks/complete
```
**Тело запроса:**
```json
{
  "task_id": 123
}
```

## Диспуты

### Создание диспута
```
POST /api/disputes/create
```
**Тело запроса:**
```json
{
  "task_id": 123
}
```

### Получение диспута
```
GET /api/disputes/get?id=789
```

### Отправка сообщения в диспут
```
POST /api/disputes/message
```
**Тело запроса:**
```json
{
  "dispute_id": 789,
  "message": "Текст сообщения"
}
```

### Получение диспутов пользователя
```
GET /api/disputes/my
```

## Система отзывов

### Создание отзыва
```
POST /api/reviews/create
```
**Тело запроса:**
```json
{
  "task_id": 123,
  "rating": 5,
  "comment": "Отличная работа!"
}
```

### Получение отзывов пользователя
```
GET /api/reviews/user?user_id=456
```

### Получение отзывов по задаче
```
GET /api/reviews/task?task_id=123
```

### Получение рейтинга пользователя
```
GET /api/reviews/rating?user_id=456
```

## Админские функции

### Получение открытых диспутов
```
GET /api/admin/disputes
```

### Назначение диспута админу
```
POST /api/admin/disputes/assign
```
**Тело запроса:**
```json
{
  "dispute_id": 789
}
```

### Разрешение диспута
```
POST /api/admin/disputes/resolve
```
**Тело запроса:**
```json
{
  "dispute_id": 789,
  "resolution": "client_won" // или "freelancer_won"
}
```

### Получение деталей диспута
```
GET /api/admin/disputes/details?id=789
```

## Статусы задач
- `open` - открыта для предложений
- `in_progress` - в работе
- `completed` - завершена
- `cancelled` - отменена
- `disputed` - в диспуте

## Статусы диспутов
- `open` - открыт
- `resolved` - разрешен

## Статусы escrow
- `pending` - заморожен
- `released` - разморожен в пользу исполнителя
- `refunded` - возвращен заказчику

## Категории задач
- `web_development` - Веб-разработка
- `mobile_development` - Мобильная разработка
- `design` - Дизайн
- `writing` - Копирайтинг
- `translation` - Переводы
- `marketing` - Маркетинг
- `other` - Другое

## Валюты
- `BTC` - Bitcoin
- `XMR` - Monero
