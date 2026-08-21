# Vikunja Custom Base — карта проекта

Эта папка подготовлена как база для личного Trello-подобного форка Vikunja.
Исходная структура Vikunja сохранена; отдельный репозиторий сайта vikunja.io (`website-main`) не включён.

## Что запускать

### Быстрый development
1. Один раз запустить `SETUP-DEV.ps1` из PowerShell.
2. Затем `RUN-DEV.ps1`.
3. Backend: http://localhost:3456
4. Frontend dev-server: http://localhost:4173

### Production-сборка Windows
Запустить `BUILD-WINDOWS.ps1`.
Скрипт собирает Vue frontend в `frontend/dist`, затем встраивает его в Go executable.
Итоговый файл: `release/vikunja.exe`.

### Проверка чистого upstream без собственной сборки
`GET-OFFICIAL-BASELINE.ps1` скачивает официальный Windows x64 пакет Vikunja 2.5.0 в `baseline/`.
Это только контрольная копия upstream и НЕ содержит будущих локальных изменений.

## Архитектура

### Frontend — `frontend/src`
- `views/` — страницы приложения.
- `components/project/views/ProjectKanban.vue` — Kanban-представление.
- `components/tasks/partials/KanbanCard.vue` — карточка на Kanban.
- `views/tasks/TaskDetailView.vue` — подробная карточка задачи.
- `components/tasks/` — форма задачи и её части.
- `services/` — обращения frontend к API.
- `models/`, `modelTypes/`, `modelSchema/` — frontend-модели данных.
- `stores/` — Pinia state.
- `router/` — маршрутизация.
- `styles/` — глобальные стили.

### Backend — `pkg`
- `models/` — сущности, CRUD, permissions и значительная часть доменной логики.
- `services/` — сложная бизнес-логика.
- `routes/api/v2/` — новые API endpoints. Новые функции добавлять сюда.
- `routes/api/v1/` — legacy API; не расширять без необходимости.
- `migration/` — миграции БД.
- `config/` — настройки.
- `events/` и listeners в моделях/сервисах — события, уведомления и фоновые реакции.

## Карта типичных изменений

### Только внешний вид Kanban
Обычно смотреть:
- `frontend/src/components/project/views/ProjectKanban.vue`
- `frontend/src/components/tasks/partials/KanbanCard.vue`
- связанные SCSS/CSS

### Новое поле задачи
Обычно затрагивает:
1. `pkg/models/tasks.go`
2. новую миграцию в `pkg/migration/`
3. API v2, если поле требует отдельной операции
4. frontend model/modelType
5. frontend service при необходимости
6. UI-компонент

### Новая функция доски / карточки
Сначала определить, можно ли реализовать её существующей моделью.
Если данные должны храниться или логика должна обеспечиваться сервером — менять backend + frontend.

### Отключение ненужной функции
По умолчанию не удалять backend-код. Сначала скрыть функцию из UI/роутера и отключить доступ настройкой,
если это возможно. Полное удаление делать только если оно действительно нужно.

## База данных
Для личного использования оставлена SQLite — это штатный режим Vikunja и самый простой вариант для разработки.
`config.yml` складывает данные в `./data/`.

## Важная деталь production-сборки
`frontend/embed.go` содержит `//go:embed all:dist`, поэтому production frontend встраивается в Go binary.
После `BUILD-WINDOWS.ps1` отдельный frontend-сервер не нужен: `vikunja.exe` отдаёт UI и API сам.
