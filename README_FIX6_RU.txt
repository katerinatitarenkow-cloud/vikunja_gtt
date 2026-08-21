FIX6: Wialon netconn может приходить как JSON boolean (true/false) или как число (0/1).
Заменить файл:
  pkg\integrations\wialon\client.go

Затем пересобрать:
  .\BUILD-WINDOWS-FIXED-v2.ps1

Базу данных, токен и настройки Wialon менять не нужно.
