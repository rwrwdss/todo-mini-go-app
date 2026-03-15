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

---

## Как это устроено

- Go backend с REST API
- SQLite для хранения задач
- Swagger для документации API
- Минимальный фронтенд на HTML/CSS/JS

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
