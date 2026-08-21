@echo off
cd /d "%~dp0release"
if not exist vikunja.exe (
  echo release\vikunja.exe not found.
  echo First run BUILD-WINDOWS.ps1 from PowerShell.
  pause
  exit /b 1
)
vikunja.exe web
pause
