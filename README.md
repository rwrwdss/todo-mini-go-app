# Todo Mini App

Минималистичный проект, но уже с полноценным API, базой данных и swagger документацией.

---

## Что умеет

- Создавать новые задачи
- Смотреть список всех задач
- Все данные сохраняются в SQLite базе
- Есть минимальный фронтенд, чтобы можно было добавить и увидеть задачи
- Swagger UI для тестирования API

---

## Как запустить

1. Клонируешь репозиторий:

```bash
git clone <твоя ссылка>
cd todo-mini-app
```

2. Запускаешь сервер:

```bash
go run cmd/server/main.go
```

3. Открываешь браузер:

- Фронтенд: http://localhost:8080
- Swagger: http://localhost:8080/swagger/index.html

Теперь можно создавать задачи прямо из браузера или через Swagger UI.

**Режим разработки фронтенда (React + Vite):** запусти бэкенд в одном терминале (`go run cmd/server/main.go`), в другом — `cd web && npm run dev`. Фронтенд будет на http://localhost:5173 с прокси к API на :8080. Для продакшена: `cd web && npm run build` — Go затем раздаёт собранное из `web/dist`.

---

## Как это устроено

- Go backend с REST API
- SQLite для хранения задач
- Swagger для документации API
- Фронтенд: React + Vite (папка `web/`), в продакшене раздаётся из `web/dist`

---

## Как проверить API через curl

Создать задачу:

```bash
curl -X POST http://localhost:8080/api/create \
  -H "Content-Type: application/json" \
  -d '{"title":"Учить Go"}'
```

Получить список задач:

```bash
curl http://localhost:8080/api/todos
```

---

## Планы для улучшения

- Сделать возможность редактировать и удалять задачи
- Добавить логирование
- Сделать красивый frontend с React или Tailwind
- Подключить Docker
- Подключить брокер сообщений(по типу rabbitmq, kafka)
