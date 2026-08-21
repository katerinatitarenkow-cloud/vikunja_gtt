@echo off
setlocal
cd /d "%~dp0"

echo ============================================================
echo Vikunja Custom Wialon - Windows build
echo ============================================================
echo.
echo Starting PowerShell with ExecutionPolicy Bypass for THIS
echo build process only. Windows policy is not changed permanently.
echo.

powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File "%~dp0BUILD-WINDOWS-FIXED-v2.ps1"
set "ERR=%ERRORLEVEL%"

echo.
if not "%ERR%"=="0" (
    echo BUILD FAILED. Exit code: %ERR%
) else (
    echo BUILD FINISHED.
    echo Result: release\vikunja.exe
)
echo.
pause
exit /b %ERR%
