
```markdown
# Tasks API

REST API для управления списком задач, реализованное на языке Go. Проект демонстрирует применение принципов SOLID, Dependency Injection, работу с HTTP‑протоколом, JSON‑сериализацией и потокобезопасным in‑memory хранилищем.

## Структура проекта

```text
TASK_API/
├── cmd/
│   └── server/
│       └── main.go                  # Точка входа, маршрутизация, подключение middleware
├── internal/
│   ├── handlers/                    # HTTP‑хендлеры
│   │   ├── health.go                # Хендлер для проверки работоспособности
│   │   └── tasks.go                 # Хендлеры CRUD задач (логика бизнес-операций)
│   ├── middleware/                  # Промежуточное ПО
│   │   └── middleware.go            # Логирование запросов, обработка ошибок, CORS
│   ├── models/                      # Модели данных
│   │   └── task.go                  # Структура Task и конструкторы
│   └── storage/                     # Логика хранения данных
│       ├── storage.go               # Интерфейс Storage
│       └── memory.go                # Реализация in‑memory хранилища
├── .gitignore                       # Игнорируемые файлы
├── go.mod                           # Зависимости модуля
└── README.md                        # Этот файл

## Требования

- Go версии 1.21 или выше.
- Операционная система: Windows / Linux / macOS.

---

## Установка и запуск

Если `go.mod` уже существует (файл есть в корне), инициализировать модуль повторно **не требуется**.

Запустите сервер из корня проекта:

```bash
go run cmd/server/main.go
```

Сервер запустится на порту `8080`.

---

## Доступные эндпоинты

| Метод | Путь | Описание | Статус успеха | Тело ответа |
|-------|------|----------|---------------|-------------|
| GET | `/health` | Проверка работоспособности сервиса | `200 OK` | `{"status":"ok"}` |
| GET | `/tasks` | Получить список всех задач | `200 OK` | `[]` или `[{...}]` |
| POST | `/tasks` | Создать новую задачу | `201 Created` | `{...}` (созданная задача) |
| GET | `/tasks/{id}` | Получить задачу по ID | `200 OK` / `404 Not Found` | `{...}` / `{"error":"task not found"}` |
| PUT | `/tasks/{id}` | Обновить задачу по ID | `200 OK` / `404 Not Found` | `{...}` / `{"error":"task not found"}` |
| DELETE | `/tasks/{id}` | Удалить задачу по ID | `204 No Content` / `404 Not Found` | — (пусто) / `{"error":"task not found"}` |

> **Важно:** Для путей с `{id}` подставляйте числовой идентификатор задачи (например, `/tasks/1`).  
> **Примечание про 204:** При успешном удалении сервер возвращает статус `204 No Content` без тела и без заголовка `Content-Type`. Это соответствует HTTP‑стандарту.  
> **Логирование:** Все запросы (включая `/health`) теперь логируются через middleware в формате `[METHOD] PATH`.

---

## Примеры использования (curl)

Для быстрой проверки статусов и заголовков во всех примерах используется флаг `-i`.

### 1. Проверка статуса (Health Check)

**Успешный запрос:**

```bash
curl -i http://localhost:8080/health
```

Ожидаемый результат:

```http
HTTP/1.1 200 OK
Content-Type: application/json
...

{"status":"ok"}
```

**Неподдерживаемый метод (ошибка 405):**

```bash
curl -i -X POST http://localhost:8080/health
```

Ожидаемо: `405 Method Not Allowed`, заголовок `Allow: GET`, JSON‑ошибка от middleware.

---

### 2. Создание задачи (POST)

**Для Windows (PowerShell):**

```powershell
curl -X POST http://localhost:8080/tasks `
  -H "Content-Type: application/json" `
  -d '{"title":"Купить хлеб","done":false}'
```

**Для Windows (CMD):**

```cmd
curl -X POST http://localhost:8080/tasks ^
  -H "Content-Type: application/json" ^
  -d "{\"title\":\"Купить хлеб\",\"done\":false}"
```

**Для Linux / macOS:**

```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Купить хлеб","done":false}'
```

Успешный ответ (`201`):

```json
{
  "id": 1,
  "title": "Купить хлеб",
  "done": false,
  "created_at": "2025-12-10T12:34:56Z"
}
```

Ошибка валидации (`400`):

```json
{"error":"invalid or missing 'title' field"}
```

---

### 3. Получение списка задач (GET)

```bash
curl -i http://localhost:8080/tasks
```

Если задач нет, вернётся пустой массив:

```json
[]
```

---

### 4. Получение задачи по ID (GET)

**Успешный запрос (задача существует):**

```bash
curl -i http://localhost:8080/tasks/1
```

Ответ (`200`):

```json
{
  "id": 1,
  "title": "Купить хлеб",
  "done": false,
  "created_at": "2025-12-10T12:34:56Z"
}
```

**Задача не найдена (404):**

```bash
curl -i http://localhost:8080/tasks/99999
```

Ответ:

```http
HTTP/1.1 404 Not Found
Content-Type: application/json

{"error":"task not found"}
```

---

### 5. Обновление задачи (PUT)

Пример для PowerShell:

```powershell
curl -X PUT http://localhost:8080/tasks/1 `
  -H "Content-Type: application/json" `
  -d '{"title":"Купить хлеб и молоко","done":true}'
```

Ответ (`200`):

```json
{
  "id": 1,
  "title": "Купить хлеб и молоко",
  "done": true,
  "created_at": "2025-12-10T12:34:56Z"
}
```

> **Важное уточнение:** При `PUT` обновляются только переданные поля (`title`, `done`). Поле `created_at` сохраняется без изменений (не перезаписывается нулевым временем).

Если задача не найдена (`404`):

```json
{"error":"task not found"}
```

---

### 6. Удаление задачи (DELETE)

```bash
curl -i -X DELETE http://localhost:8080/tasks/1
```

При успехе:

```http
HTTP/1.1 204 No Content
```

Обратите внимание: **нет** `Content-Type`, **нет** тела — это ожидаемое поведение.

Если задача не найдена, вернётся:

```json
{"error":"task not found"}
```

---

### 7. Проверка несуществующего маршрута (404 JSON)

```bash
curl -i http://localhost:8080/unknown-path
```

Ожидаемо:

```http
HTTP/1.1 404 Not Found
Content-Type: application/json

{"error":"not found"}
```

Это подтверждает, что кастомный 404‑middleware работает корректно и отдаёт JSON даже для несуществующих путей.

---

## Технические детали

### Обработка времени

Поле `created_at` хранится как `time.Time` и автоматически сериализуется в строку формата RFC 3339 (ISO 8601). Время всегда фиксируется в UTC.

### Потокобезопасность

Реализация `MemoryStorage` использует `sync.RWMutex`:

- `RLock/RUnlock` — при чтении списка задач или одной задачи.
- `Lock/Unlock` — при создании, обновлении и удалении, чтобы избежать состояния гонки (race condition) при работе с мапой.

### Единый формат ошибок

Все ошибки (валидация, «не найдено», «метод не разрешён») теперь обрабатываются в `internal/middleware/middleware.go` и возвращаются в едином формате:

```json
{"error":"текст ошибки"}
```

### Логирование

Благодаря middleware, **все** входящие запросы логируются в консоль в формате:

```text
[GET] /health
[POST] /tasks
[DELETE] /tasks/1
[GET] /unknown-path
```

Также логируются ошибки сериализации JSON, если клиент разорвал соединение во время отправки ответа.

### Расширяемость

Благодаря использованию интерфейса `storage.Storage`, вы можете легко заменить текущее in‑memory хранилище на любую другую реализацию (например, PostgreSQL), не меняя код хендлеров и middleware. Достаточно реализовать методы интерфейса для работы с базой данных и передать новый экземпляр в `handlers.New()`.

---

```

