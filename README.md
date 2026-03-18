# Todo Mini App

Минималистичный проект с API, базой данных и Swagger-документацией. Поддержка личных и корпоративных пространств, роли участников, назначение задач.

---

## Что умеет

### Задачи и дерево

- **Задачи:** создание, просмотр, редактирование и удаление (с учётом прав роли)
- **Редактирование:** форма с полями заголовок, описание, приоритет, тег, срок, назначенный; отмена по Cancel или Escape
- **Удаление:** для корневых задач — модальное предупреждение о каскадном удалении подзадач
- **Дерево задач:** корневые сгруппированы по тегу, внутри тега — по приоритету (high → med → low → none); подзадачи под родителем
- **Теги:** поле тега в форме с подсказками существующих тегов

### Пространства и роли

- **Личное пространство (My tasks):** только задачи текущего пользователя; дерево + опция «Show tasks from other spaces»
- **Корпоративное пространство:** задачи пространства; доска (Board) и список (List); приглашение участников (Members), роли admin / member
- **Роль Admin:** полный CRUD по задачам пространства, назначение исполнителей, управление участниками
- **Роль Member:** видит только назначенные ему задачи; может только отмечать выполнение (toggle done), без редактирования и удаления. В панели задачи и на карточках кнопки Edit/Delete для Member скрыты

### Назначение и видимость

- **Назначение задач:** в форме задачи можно указать Assignee (участника пространства). На доске и в списке отображается подпись «Assigned to: {имя}» или «You», плюс тултип с полным именем
- **Задачи из других пространств:** в личном пространстве при включённом чекбоксе «Show tasks from other spaces» блок показывает задачи из корпоративных пространств, назначенные текущему пользователю; обновление по API каждые 15 с; на карточках подпись «Assigned to you»; при ошибке загрузки выводится сообщение без очистки списка

### Авторизация и прочее

- **Авторизация:** регистрация и вход по email/паролю; JWT в заголовке
- **Dashboard:** в корпоративном пространстве — статистика по задачам; у Member — только по своим выполненным
- **Хранение:** PostgreSQL (users, todos, spaces, space_members и др.)
- **Swagger UI** для тестирования API (тёмная тема)

---

## Как запустить

1. Клонируешь репозиторий:

```bash
git clone <твоя ссылка>
cd todo-mini-app
```

2. Поднимаешь PostgreSQL (бэкенд использует **только** PostgreSQL):

- **macOS (Homebrew):** `brew install postgresql@16` (или другую версию), затем `brew services start postgresql@16`. Создать БД: `createdb todo` (или через psql: `CREATE DATABASE todo;`).
- Либо скопируй настройки: `cp .env.example .env` и отредактируй `DATABASE_URL`, `JWT_SECRET` и при необходимости `PORT`, `WEB_DIST`. Все переменные описаны в `.env.example`.

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

### Запуск через Docker (вся платформа: PostgreSQL + бэкенд + фронтенд)

```bash
docker compose up --build
```

После сборки приложение доступно на http://localhost:8080, Swagger — http://localhost:8080/swagger/index.html. БД PostgreSQL поднимается в контейнере (порт 5432). Остановка: `docker compose down`.

---

## Как это устроено

- Go backend с REST API
- PostgreSQL: таблицы `users` (email, password_hash, name), `todos` (user_id, title, done, …); пароли хешируются bcrypt, выдача JWT
- Swagger для документации API
- Фронтенд: React + Vite (папка `web/`); страница входа/регистрации в стиле `web/examples/auth.html`, токен в localStorage, запросы с `Authorization: Bearer <token>`

---

## Как проверить API через curl

Эндпоинты задач требуют авторизации. Сначала получи токен:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"your@email.com","password":"yourpassword"}'
```

В ответе будет `token`. Дальше передавай его в заголовке:

```bash
export TOKEN="<токен из ответа login>"
curl http://localhost:8080/api/todos -H "Authorization: Bearer $TOKEN"
curl -X POST http://localhost:8080/api/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Учить Go"}'
```

Удобнее тестировать через Swagger UI: там можно залогиниться и нажать «Authorize».

---

## Планы для улучшения

- Добавить логирование
- ~~Подключить Docker~~ (реализовано)
- Подключить брокер сообщений (rabbitmq, kafka и т.п.)

![Screenshot](https://github.com/user-attachments/assets/a0248197-f54a-4800-ad06-69ca25bc4a07)
