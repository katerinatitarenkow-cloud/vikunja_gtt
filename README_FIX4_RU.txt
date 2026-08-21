FIX4 — автоматическое получение токена Wialon Hosting

Причина ошибки FIX3: кнопка открывала login.html, но не задавала callback и не принимала access_token обратно в Vikunja.

Что исправлено:
- авторизация открывается во всплывающем окне;
- redirect_uri = https://hosting.wialon.com/post_token.html;
- Vikunja принимает access_token через window.postMessage только от https://hosting.wialon.com;
- токен автоматически сохраняется в backend;
- интеграция автоматически включается;
- сразу выполняется проверка подключения;
- если popup заблокирован, показывается понятная ошибка.

Замена:
Скопировать папку frontend из этого архива в корень исходников Vikunja с заменой файлов.
Затем выполнить .\BUILD-WINDOWS-FIXED-v2.ps1
После запуска новой сборки нажать Ctrl+F5.
