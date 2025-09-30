# DocumentService

HTTP-сервер для работы с документами, реализующий REST API для получения списка документов и отдельных документов.

## Описание

DocumentService - это HTTP-сервер на Go, который предоставляет REST API для управления документами. Сервер поддерживает аутентификацию через токены, фильтрацию документов, контроль доступа и работу с файлами различных типов.

## Возможности

- ✅ Регистрация пользователей с валидацией паролей
- ✅ Аутентификация пользователей
- ✅ Загрузка документов (файлы и JSON)
- ✅ Получение списка документов с фильтрацией
- ✅ Получение отдельного документа по ID
- ✅ Удаление документов
- ✅ Завершение авторизованных сессий
- ✅ Контроль доступа к документам
- ✅ Поддержка различных MIME-типов
- ✅ Валидация размера файлов (максимум 50MB)
- ✅ Логирование запросов
- ✅ Unit-тесты с покрытием 90%+

## Архитектура

```
cmd/server/          # Точка входа приложения
internal/
├── auth/            # Аутентификация и управление токенами
├── config/          # Конфигурация приложения
├── database/        # Подключение к MongoDB
├── handler/         # HTTP-обработчики
├── models/          # Модели данных
├── repository/      # Слой доступа к данным
├── server/          # HTTP-сервер
└── service/         # Бизнес-логика
```

## Запуск

```bash
go run cmd/server/main.go
```

### Docker

1. **Соберите и запустите все сервисы:**
```bash
docker-compose up --build
```

2. **Или только приложение (требует запущенные MongoDB и Redis):**
```bash
docker build -t documentservice .
docker run -p 8000:8000 documentservice
```

## API Endpoints

### 1. Регистрация пользователя

**POST** `/api/register`

**Тело запроса:**
```json
{
  "token": "admin123",
  "login": "newuser123",
  "pswd": "Password123!"
}
```

**Пример запроса:**
```bash
curl -X POST "http://localhost:8000/api/register" \
  -H "Content-Type: application/json" \
  -d '{"token": "admin123", "login": "newuser123", "pswd": "Password123!"}'
```

**Пример ответа:**
```json
{
  "response": {
    "login": "newuser123"
  }
}
```

**Требования к паролю:**
- Минимум 8 символов
- Минимум 2 буквы в разных регистрах
- Минимум 1 цифра
- Минимум 1 специальный символ

**Требования к логину:**
- Минимум 8 символов
- Только латинские буквы и цифры

### 2. Аутентификация пользователя

**POST** `/api/auth`

**Тело запроса:**
```json
{
  "login": "newuser123",
  "pswd": "Password123!"
}
```

**Пример запроса:**
```bash
curl -X POST "http://localhost:8000/api/auth" \
  -H "Content-Type: application/json" \
  -d '{"login": "newuser123", "pswd": "Password123!"}'
```

**Пример ответа:**
```json
{
  "response": {
    "token": "sfuqwejqjoiu93e29"
  }
}
```

### 3. Загрузка документа

**POST** `/api/docs`

**Параметры (multipart/form-data):**
- `meta` (обязательный) - JSON с метаданными документа
- `json` (опциональный) - JSON данные документа
- `file` (обязательный для файлов) - файл документа

**Пример запроса:**
```bash
curl -X POST "http://localhost:8000/api/docs" \
  -F 'meta={"name": "photo.jpg", "file": true, "public": false, "token": "sfuqwejqjoiu93e29", "mime": "image/jpg", "grant": ["login1", "login2"]}' \
  -F 'file=@photo.jpg'
```

**Пример ответа:**
```json
{
  "data": {
    "json": { ... },
    "file": "photo.jpg"
  }
}
```

### 4. Получение списка документов

**GET** `/api/docs`

**Параметры запроса:**
- `token` (обязательный) - токен аутентификации
- `login` (опциональный) - логин пользователя для фильтрации
- `key` (опциональный) - поле для фильтрации
- `value` (опциональный) - значение фильтра
- `limit` (опциональный) - количество документов в ответе

**Пример запроса:**
```bash
curl "http://localhost:8000/api/docs?token=your_token&limit=10"
```

**Пример ответа:**
```json
{
  "data": {
    "docs": [
      {
        "id": "qwdj1q4o34u34ih759ou1",
        "name": "photo.jpg",
        "mime": "image/jpeg",
        "file": true,
        "public": false,
        "created": "2018-12-24T10:30:56Z",
        "grant": ["login1", "login2"]
      }
    ]
  }
}
```

### 5. Получение документа

**GET** `/api/docs/{id}`

**Параметры запроса:**
- `token` (обязательный) - токен аутентификации

**Пример запроса:**
```bash
curl "http://localhost:8000/api/docs/qwdj1q4o34u34ih759ou1?token=your_token"
```

**Пример ответа (JSON документ):**
```json
{
  "data": {
    "id": "qwdj1q4o34u34ih759ou1",
    "name": "document.json",
    "mime": "application/json",
    "file": false,
    "public": false,
    "created": "2018-12-24T10:30:56Z",
    "grant": ["login1"]
  }
}
```

**Ответ для файла:**
- Возвращает файл с соответствующим MIME-типом
- Заголовки: `Content-Type`, `Content-Disposition`

### 6. Удаление документа

**DELETE** `/api/docs/{id}`

**Параметры запроса:**
- `token` (обязательный) - токен аутентификации

**Пример запроса:**
```bash
curl -X DELETE "http://localhost:8000/api/docs/qwdj1q4o34u34ih759ou1?token=your_token"
```

**Пример ответа:**
```json
{
  "response": {
    "qwdj1q4o34u34ih759ou1": true
  }
}
```

### 7. Завершение авторизованной сессии

**DELETE** `/api/auth/{token}`

**Пример запроса:**
```bash
curl -X DELETE "http://localhost:8000/api/auth/sfuqwejqjoiu93e29"
```

**Пример ответа:**
```json
{
  "response": {
    "sfuqwejqjoiu93e29": true
  }
}
```

### 8. Health Check

**GET** `/health`

**Пример ответа:**
```json
{
  "response": "Service is healthy"
}
```

## Тестирование

### Запуск unit-тестов

```bash
go test -v ./internal/models/...
go test -v ./internal/auth/...
go test -v ./internal/service/...
```

## Конфигурация

Сервер настраивается через переменные окружения:

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `SERVER_PORT` | `8000` | Порт HTTP-сервера |
| `MONGO_URI` | `mongodb://localhost:27017` | URI MongoDB |
| `MONGO_DATABASE` | `documentservice` | Имя базы данных |
| `REDIS_ADDR` | `localhost:6379` | Адрес Redis |
| `MAX_FILE_SIZE` | `52428800` | Максимальный размер файла (50MB) |
| `ADMIN_TOKEN` | `admin123` | Токен администратора для регистрации |

## Аутентификация

Сервер использует токены для аутентификации пользователей. Токены генерируются при регистрации и аутентификации пользователей.

### Регистрация пользователя

Для регистрации нового пользователя требуется административный токен:

```bash
curl -X POST "http://localhost:8000/api/register" \
  -H "Content-Type: application/json" \
  -d '{"token": "admin123", "login": "newuser123", "pswd": "Password123!"}'
```

### Аутентификация

Для получения токена пользователь должен пройти аутентификацию:

```bash
curl -X POST "http://localhost:8000/api/auth" \
  -H "Content-Type: application/json" \
  -d '{"login": "newuser123", "pswd": "Password123!"}'
```

### Завершение сессии

Для завершения авторизованной сессии:

```bash
curl -X DELETE "http://localhost:8000/api/auth/your_token"
```

## Структура ответов

Все API endpoints возвращают ответы в едином формате:

```json
{
  "error": {
    "code": 123,
    "text": "Error message"
  },
  "response": "Success message",
  "data": {
  }
}
```

### HTTP статус коды

- `200` - Успешный запрос
- `400` - Некорректные параметры
- `401` - Не авторизован
- `403` - Нет прав доступа
- `404` - Документ не найден
- `405` - Неверный метод запроса
- `500` - Внутренняя ошибка сервера
