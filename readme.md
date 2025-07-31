# Archive Service

Сервис для создания ZIP-архивов из публичных ссылок на `.jpeg` и `.pdf` файлы.

---

## Возможности

- Создание задач на архивирование файлов.
- Добавление до 3 ссылок в задачу.
- Получение статуса задачи и ссылки на готовый архив.
- Обработка недоступных ссылок (с указанием в ответе).
- Поддержка только `.jpeg` и `.pdf` файлов.
- До 3 одновременно активных задач.
- Раздача архивов через HTTP.

---

## Конфигурация

Файл: `config/config.go`

```go
Config{
    Port:              "8080",
    MaxFilesPerTask:   3,
    AllowedExtensions: []string{".jpeg", ".pdf"},
    MaxActiveTasks:    3,
}
```

---

## API

### Создание задачи
**POST** `/task`

**Пример запроса:**
```bash
curl -X POST http://localhost:8080/task
```

**Ответ:**
```json
{ "task_id": "abc123" }
```

---

### Добавление файла в задачу
**POST** `/task/{task_id}/add`

**Пример:**
```bash
curl -X POST http://localhost:8080/task/abc123/add \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/file.jpeg"}'
```

**Ответ:**
```json
{ "message": "File queued" }
```

---

### Получение статуса задачи
**GET** `/task/{task_id}/status`

**Пример:**
```bash
curl http://localhost:8080/task/abc123/status
```

**Пример ответа:**
```json
{
  "status": "done",
  "archive_url": "http://localhost:8080/archives/abc123.zip",
  "error_files": ["https://badlink.com/file.pdf"]
}
```

**Возможные статусы:**
- `pending`
- `progress`
- `done`
- `error`

---

## Архитектура проекта

```
.
├── main.go            # Точка входа
├── api/               # HTTP-обработчики
├── config/            # Конфигурация
├── task/              # Логика задач, структура TaskManager
├── utils/             # Утилиты: загрузка файлов, проверка расширений, ID
├── archives/          # Папка с готовыми архивами
└── readme.md
```

---

## Статическая раздача архивов

Архивы доступны по адресу:
```
http://localhost:8080/archives/{task_id}.zip
```

> Убедитесь, что `main.go` содержит обработку:
> ```go
> http.Handle("/archives/", http.StripPrefix("/archives/", http.FileServer(http.Dir("archives"))))
> ```

---

## Тесты

Запуск всех тестов:
```bash
go test ./...
```

Покрытие:
- task manager
- utils (валидация, генерация ID)
- api handler (в том числе HTTP-status проверки)

---

## Ограничения

- Максимум 3 задачи в статусе "выполняется" одновременно.
- Только `.jpeg` и `.pdf` ссылки.
- Только 3 файла на одну задачу.
- Без Docker и сторонних библиотек.

---

## Запуск

```bash
go run main.go
```

Сервер будет доступен по адресу:
```
http://localhost:8080
```

