# FIX16 — SQLite в отдельной папке Vikunja

## Что меняется

После сборки и запуска SQLite хранится в отдельной папке рядом с Vikunja:

```text
release\
├─ vikunja.exe
├─ config.yml
├─ RUN.cmd
├─ database\
│  └─ vikunja.db
└─ data\
   ├─ files\
   └─ logs\
```

Путь по умолчанию:

```text
./database/vikunja.db
```

Относительный `database.path` теперь разрешается относительно `service.rootpath`, а не `%LOCALAPPDATA%`.

## Автоматический перенос старой базы

При первом запуске FIX16, если `database/vikunja.db` ещё не существует, Vikunja ищет старую базу в прежних местах:

- `<корень Vikunja>/data/vikunja.db`;
- `<корень Vikunja>/vikunja.db`;
- `%LOCALAPPDATA%/Vikunja/data/vikunja.db`;
- `%LOCALAPPDATA%/Vikunja/vikunja.db`.

Если найдено несколько старых файлов, выбирается наиболее свежий по времени изменения.

Старая база **не удаляется** — она остаётся резервной копией. Если существует SQLite WAL (`vikunja.db-wal`) или rollback journal, они также переносятся, чтобы не потерять зафиксированные, но ещё не checkpoint-нутые изменения.

Если новая `database/vikunja.db` уже существует, FIX16 её не перезаписывает.

## Установка

1. Остановить Vikunja перед первым запуском новой сборки.
2. Распаковать FIX16 в корень исходников с заменой файлов.
3. Собрать:

```powershell
.\BUILD-WINDOWS-FIXED-v2.ps1
```

4. Запустить:

```powershell
.\release\RUN.cmd
```

После первого запуска проверить:

```text
release\database\vikunja.db
```

В логе также должна появиться строка вида:

```text
Using SQLite database at: ...\release\database\vikunja.db
```

Если была найдена старая база, перед ней будет сообщение о её копировании в portable location.

## Важно

FIX16 меняет только расположение SQLite. Схема БД и данные не изменяются; новая миграция таблиц не требуется.

Папка `data` не удаляется: она по-прежнему используется для файлов и логов.
