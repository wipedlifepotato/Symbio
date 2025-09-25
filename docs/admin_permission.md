# Система разрешений администраторов mFrelance

## Обзор

Система разрешений администраторов построена на битовых флагах (bitwise permissions), что позволяет гибко назначать различные комбинации прав доступа разным администраторам.

## Структура разрешений

### База данных

В таблице `users` добавлены два новых поля:
- `admin_title` VARCHAR(100) - должность/звание администратора
- `permissions` INTEGER - битовая маска разрешений

### Константы разрешений

```go
const (
    CanChangeBalance = 1 << iota  // 1 (бит 0) - изменение балансов пользователей
    CanBlockUsers                 // 2 (бит 1) - блокировка/разблокировка пользователей
    CanManageDisputes             // 4 (бит 2) - управление спорами
    // Будущие разрешения:
    // CanDeleteTasks     = 8  (бит 3)
    // CanViewAnalytics   = 16 (бит 4)
    // CanManageAdmins    = 32 (бит 5)
)
```

## Как работают битовые разрешения

### Принцип работы

Каждое разрешение представлено степенью двойки:
- 2^0 = 1 (CanChangeBalance)
- 2^1 = 2 (CanBlockUsers)
- 2^2 = 4 (CanManageDisputes)
- 2^3 = 8 (будущие разрешения)

### Примеры значений

| Значение | Двоичный вид | Разрешения |
|----------|-------------|------------|
| 0 | 000 | Нет разрешений |
| 1 | 001 | Только изменение балансов |
| 2 | 010 | Только блокировка пользователей |
| 3 | 011 | Изменение балансов + блокировка |
| 4 | 100 | Только управление спорами |
| 7 | 111 | Все разрешения |

## API для работы с разрешениями

### Проверка разрешений

```go
// Проверка конкретного разрешения
func HasPermission(userID int64, perm int) bool {
    var permissions int
    err := db.Postgres.Get(&permissions, "SELECT permissions FROM users WHERE id=$1", userID)
    if err != nil {
        return false
    }
    return permissions & perm != 0
}

// Пример использования
if HasPermission(userID, models.CanChangeBalance) {
    // Пользователь может изменять балансы
}
```

### Middleware для защиты эндпоинтов

```go
func RequirePermission(perm int) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            claims := server.GetUserFromContext(r)
            if claims == nil {
                http.Error(w, "user not found", http.StatusUnauthorized)
                return
            }
            if !HasPermission(claims.UserID, perm) {
                http.Error(w, "insufficient permissions", http.StatusForbidden)
                return
            }
            next(w, r)
        }
    }
}

// Применение
func AdminUpdateBalanceHandler(w http.ResponseWriter, r *http.Request) {
    // Этот обработчик требует разрешения CanChangeBalance
}
```

## Назначение разрешений

### SQL запросы для управления разрешениями

```sql
-- Дать разрешение на изменение балансов
UPDATE users SET permissions = permissions | 1 WHERE id = <admin_id>;

-- Дать разрешение на блокировку пользователей
UPDATE users SET permissions = permissions | 2 WHERE id = <admin_id>;

-- Дать разрешение на управление спорами
UPDATE users SET permissions = permissions | 4 WHERE id = <admin_id>;

-- Дать все разрешения
UPDATE users SET permissions = 7 WHERE id = <admin_id>;

-- Убрать разрешение на изменение балансов
UPDATE users SET permissions = permissions & ~1 WHERE id = <admin_id>;

-- Проверить разрешения пользователя
SELECT permissions & 1 as can_change_balance,
       permissions & 2 as can_block_users,
       permissions & 4 as can_manage_disputes
FROM users WHERE id = <admin_id>;
```

### Программное назначение разрешений

```go
// Добавить разрешение
func AddPermission(userID int64, perm int) error {
    _, err := db.Postgres.Exec("UPDATE users SET permissions = permissions | $1 WHERE id = $2", perm, userID)
    return err
}

// Убрать разрешение
func RemovePermission(userID int64, perm int) error {
    _, err := db.Postgres.Exec("UPDATE users SET permissions = permissions & ~$1 WHERE id = $2", perm, userID)
    return err
}

// Установить точные разрешения
func SetPermissions(userID int64, permissions int) error {
    _, err := db.Postgres.Exec("UPDATE users SET permissions = $1 WHERE id = $2", permissions, userID)
    return err
}
```

## Применение в коде

### Защищенные эндпоинты

```go
// Изменение баланса - требует CanChangeBalance
func AdminUpdateBalanceHandler(w http.ResponseWriter, r *http.Request) {
    // Код обработчика
}

// Блокировка пользователей - требует CanBlockUsers
func BlockUserHandler(w http.ResponseWriter, r *http.Request) {
    // Код обработчика
}
```

### Проверка в бизнес-логике

```go
// В SendDisputeMessageHandler
isAdmin, _ := db.IsAdmin(db.Postgres, userID)
hasPerm := HasPermission(userID, models.CanManageDisputes)
if task.ClientID == userID || acceptedOffer.FreelancerID == userID || isAdmin || hasPerm {
    // Разрешено отправлять сообщения в спор
}
```

## API ответы

### Профиль пользователя

```json
{
  "username": "adminuser",
  "profile": { ... },
  "is_admin": true,
  "admin_title": "Senior Administrator",
  "permissions": 7
}
```

### Детали спора

```json
{
  "success": true,
  "dispute": { ... },
  "task": { ... },
  "escrow": { ... },
  "messages": [ ... ],
  "admin": {
    "id": 123,
    "username": "adminuser",
    "title": "Senior Administrator"
  }
}
```

## Миграция и обратная совместимость

### Миграция базы данных

```sql
ALTER TABLE users ADD COLUMN admin_title VARCHAR(100);
ALTER TABLE users ADD COLUMN permissions INTEGER DEFAULT 0;
```

### Обратная совместимость

- Старые админы с `is_admin = true` продолжают работать
- Новые возможности используют поле `permissions`
- Поле `is_admin` остается для совместимости

## Преимущества системы

1. **Гибкость**: Разные комбинации разрешений для разных админов
2. **Масштабируемость**: Легко добавить новые разрешения
3. **Безопасность**: Принцип наименьших привилегий
4. **Производительность**: Быстрые битовые операции
5. **Прозрачность**: Легко понять, какие разрешения есть у пользователя

## Будущие расширения

Система легко расширяема. Для добавления нового разрешения:

1. Добавить константу: `NewPermission = 1 << iota`
2. Обновить документацию
3. Использовать в коде: `RequirePermission(models.NewPermission)`

Примеры будущих разрешений:
- CanDeleteTasks (удаление задач)
- CanViewAnalytics (просмотр аналитики)
- CanManageAdmins (управление другими админами)
- CanExportData (экспорт данных)