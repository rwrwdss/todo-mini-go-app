# Frontend — React + Vite

## Скрипты

- `npm run dev` — dev-сервер на http://localhost:5173, прокси `/api` и `/swagger` на бэкенд (localhost:8080).
- `npm run build` — сборка в `dist/`. Бэкенд раздаёт её с корня при наличии `web/dist`.
- `npm run preview` — просмотр собранного билда локально.

## Структура

- `src/api/todos.js` — запросы к API (getTodos, createTodo).
- `src/components/` — TodoForm, TodoList.
- `src/App.jsx` — корневой компонент и загрузка списка.

Запуск: сначала бэкенд (`go run cmd/server/main.go`), затем `npm run dev` в этой папке.

<img width="1470" height="833" alt="image" src="https://github.com/user-attachments/assets/a0248197-f54a-4800-ad06-69ca25bc4a07" />
