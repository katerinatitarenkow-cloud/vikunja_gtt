FIX8 — замена левого логотипа Vikunja

Что изменено:
- logo.png из корня исходника скопирован в frontend/src/assets/company-logo.png
- frontend/src/components/home/AppHeader.vue использует новый логотип
- frontend/src/components/home/Navigation.vue использует новый логотип

Изменение затрагивает логотип приложения слева в основной навигации.
Логотип на экране входа и других служебных страницах не менялся.

Установка:
1. Распаковать архив в корень исходника Vikunja с заменой файлов.
2. Запустить .\BUILD-WINDOWS-FIXED-v2.ps1
3. Запустить .\release\vikunja.exe
4. В браузере выполнить Ctrl+F5 из-за Service Worker/PWA-кэша.
