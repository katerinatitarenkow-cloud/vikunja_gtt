$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

$goBin = (go env GOPATH) + '\bin'
if ($env:PATH -notlike "*$goBin*") {
    $env:PATH = "$goBin;$env:PATH"
}

if (-not (Get-Command mage -ErrorAction SilentlyContinue)) {
    throw 'Mage is missing. Run SETUP-DEV.ps1 first.'
}
if (-not (Get-Command pnpm -ErrorAction SilentlyContinue)) {
    throw 'pnpm is missing. Run SETUP-DEV.ps1 first.'
}
if (-not (Test-Path 'frontend\node_modules')) {
    throw 'Frontend dependencies are missing. Run SETUP-DEV.ps1 first.'
}

$env:RELEASE_VERSION = 'v2.5.0-custom'

Write-Host '=== Building frontend ===' -ForegroundColor Cyan
Push-Location frontend
'{"VERSION":"v2.5.0-custom"}' | Set-Content -Encoding UTF8 src/version.json
pnpm build
Pop-Location

Write-Host '=== Building Vikunja executable with embedded frontend ===' -ForegroundColor Cyan
mage build

New-Item -ItemType Directory -Force -Path release | Out-Null
Copy-Item -Force vikunja.exe release\vikunja.exe
Copy-Item -Force config.yml release\config.yml
New-Item -ItemType Directory -Force -Path release\data | Out-Null

@'
@echo off
cd /d "%~dp0"
vikunja.exe web
pause
'@ | Set-Content -Encoding ASCII release\RUN.cmd

Write-Host ''
Write-Host 'READY: release\vikunja.exe' -ForegroundColor Green
Write-Host 'Run release\RUN.cmd and open http://localhost:3456'
