# Система фриланса - Фронтенд

## Обзор
Полноценная система фриланса интегрирована в Symfony приложение с использованием Bootstrap 5 для UI.

## Установленные компоненты

### Bootstrap 5 (локально)
- **CSS**: `public/css/bootstrap.min.css`
- **JS**: `public/js/bootstrap.bundle.min.js`
- **Icons**: `public/css/bootstrap-icons.css`
- **Fonts**: `public/fonts/bootstrap-icons.woff2`, `public/fonts/bootstrap-icons.woff`

## Контроллеры

### TaskController
- `GET /tasks` - Список задач с фильтрацией
- `GET /tasks/create` - Форма создания задачи
- `POST /tasks/create` - Создание задачи
- `GET /tasks/{id}` - Детальный просмотр задачи
- `POST /tasks/{id}` - Действия с задачей (предложения, завершение)

### DisputeController
- `GET /disputes` - Список диспутов пользователя
- `GET /disputes/{id}` - Детальный просмотр диспута
- `POST /disputes/{id}` - Отправка сообщений в диспут
- `GET /disputes/create/{taskId}` - Создание диспута

### ReviewController
- `GET /reviews/create/{taskId}` - Форма создания отзыва
- `POST /reviews/create/{taskId}` - Создание отзыва
- `GET /reviews/user/{userId}` - Отзывы пользователя

### AdminDisputeController
- `GET /admin/disputes` - Список всех диспутов для админа
- `GET /admin/disputes/{id}` - Детальный просмотр диспута для админа
- `POST /admin/disputes/{id}` - Админские действия (назначение, разрешение)

## Шаблоны

### Задачи
- `templates/task/index.html.twig` - Список задач с фильтрами
- `templates/task/create.html.twig` - Форма создания задачи
- `templates/task/show.html.twig` - Детальный просмотр задачи

### Диспуты
- `templates/dispute/index.html.twig` - Список диспутов пользователя
- `templates/dispute/show.html.twig` - Детальный просмотр диспута

### Отзывы
- `templates/review/create.html.twig` - Форма создания отзыва
- `templates/review/user.html.twig` - Отзывы пользователя

### Админка
- `templates/admin/disputes.html.twig` - Управление диспутами
- `templates/admin/dispute_show.html.twig` - Детальный просмотр диспута для админа

## Функциональность

### Для пользователей
1. **Создание задач** - заказчики могут создавать задачи с описанием и бюджетом
2. **Подача предложений** - исполнители могут предлагать свои услуги
3. **Принятие предложений** - заказчики выбирают исполнителей
4. **Завершение задач** - исполнители отмечают задачи как выполненные
5. **Создание диспутов** - при спорах можно открыть диспут
6. **Оставление отзывов** - после завершения можно оставить отзыв

### Для администраторов
1. **Просмотр диспутов** - список всех открытых диспутов
2. **Назначение диспутов** - админы могут назначить диспуты себе
3. **Разрешение споров** - админы решают исход диспутов
4. **Управление средствами** - админы могут разморозить/заморозить средства

## Интеграция с API

Все контроллеры используют сервис `MFrelance` для взаимодействия с Go API:
- Аутентификация через JWT токены
- CRUD операции с задачами, предложениями, диспутами
- Обработка ошибок и уведомлений

## Стилизация

- **Bootstrap 5** для responsive дизайна
- **Bootstrap Icons** для иконок
- **Кастомные стили** в `public/css/style.css`
- **Адаптивная верстка** для мобильных устройств

## Навигация

Обновлен базовый шаблон `base.html.twig` с:
- Навигационным меню
- Поддержкой flash сообщений
- Локальными ресурсами Bootstrap
- Адаптивным дизайном

## Запуск

```bash
cd /home/user/mFrelance/petri-frontendphp
php -S localhost:8001 -t public
```

Система будет доступна по адресу `http://localhost:8001`
