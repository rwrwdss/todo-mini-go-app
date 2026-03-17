# Todo Mini App

Минималистичный проект, но уже с полноценным API, базой данных и swagger документацией.

---

## Что умеет

- **Задачи:** создание, просмотр, редактирование и удаление
- **Редактирование:** форма с полями заголовок, описание, приоритет, тег; отмена по кнопке Cancel или Escape
- **Удаление:** для корневых задач — модальное предупреждение о каскадном удалении подзадач; подтверждение или отмена
- **Дерево задач:** корневые задачи сгруппированы по тегу, внутри тега — по приоритету (high → med → low → none); подзадачи отображаются под родителем
- **Теги:** поле тега в форме с подсказками существующих тегов
- **Авторизация:** регистрация и вход по email/паролю; JWT в заголовке; у каждого пользователя свой список задач
- Хранение: PostgreSQL (таблицы `users`, `todos` с `user_id`)
- Swagger UI для тестирования API

---

## Как запустить

1. Клонируешь репозиторий:

```bash
git clone <твоя ссылка>
cd todo-mini-app
```

2. Поднимаешь PostgreSQL (бэкенд использует **только** PostgreSQL):

- **macOS (Homebrew):** `brew install postgresql@16` (или другую версию), затем `brew services start postgresql@16`. Создать БД: `createdb todo` (или через psql: `CREATE DATABASE todo;`).
- Либо укажи свою строку подключения: `export DATABASE_URL="postgres://user:pass@localhost/todo?sslmode=disable"`.
- Опционально: `export JWT_SECRET="your-secret-key"` (по умолчанию — dev-значение).

Без запущенного PostgreSQL сервер выведет ошибку и подсказку.

3. Запускаешь сервер:

```bash
go run cmd/server/main.go
```

Либо собрать бинарник (чтобы не засорять корень): `go build -o bin/server ./cmd/server` и запуск: `./bin/server`.

4. Открываешь браузер:

- Фронтенд: http://localhost:8080
- Swagger: http://localhost:8080/swagger/index.html

Теперь можно создавать задачи прямо из браузера или через Swagger UI.

**Режим разработки фронтенда (React + Vite):** запусти бэкенд в одном терминале (`go run cmd/server/main.go`), в другом — `cd web && npm run dev`. Фронтенд будет на http://localhost:5173 с прокси к API на :8080. Для продакшена: `cd web && npm run build` — Go затем раздаёт собранное из `web/dist`.

---

## Как это устроено

- Go backend с REST API
- PostgreSQL: таблицы `users` (email, password_hash, name), `todos` (user_id, title, done, …); пароли хешируются bcrypt, выдача JWT
- Swagger для документации API
- Фронтенд: React + Vite (папка `web/`); страница входа/регистрации в стиле `web/examples/auth.html`, токен в localStorage, запросы с `Authorization: Bearer <token>`

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

- Добавить логирование
- Подключить Docker
- Подключить брокер сообщений (rabbitmq, kafka и т.п.)

![Screenshot](https://github.com/user-attachments/assets/a0248197-f54a-4800-ad06-69ca25bc4a07)
