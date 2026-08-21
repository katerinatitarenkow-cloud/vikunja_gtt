param(
    [switch]$SkipFrontend
)

$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

function Require-Command($name, $hint) {
    if (-not (Get-Command $name -ErrorAction SilentlyContinue)) {
        throw "$name is not installed. $hint"
    }
}

Require-Command 'go' 'Install Go 1.26.4 or newer.'
Require-Command 'pnpm' 'Run .\SETUP-DEV.ps1 first.'

# node_modules copied between machines (or packed into ZIP) may exist while
# pnpm's links/shims inside it are no longer valid. Check actual packages, not
# only the directory, and repair the install when necessary.
$frontendDepsHealthy =
    (Test-Path 'frontend\node_modules\vue') -and
    (Test-Path 'frontend\node_modules\vite') -and
    (Test-Path 'frontend\node_modules\.bin\vite.cmd')

if (-not $frontendDepsHealthy) {
    Write-Host '=== Installing/repairing frontend dependencies ===' -ForegroundColor Cyan
    Push-Location frontend
    try {
        pnpm install --frozen-lockfile
        if ($LASTEXITCODE -ne 0) {
            throw "pnpm install failed with exit code $LASTEXITCODE"
        }
    }
    finally {
        Pop-Location
    }
}

# Vikunja uses github.com/mattn/go-sqlite3 for SQLite.
# A CGO-enabled Windows build therefore needs a C compiler.
$gcc = Get-Command gcc -ErrorAction SilentlyContinue
if (-not $gcc) {
    $msysGcc = 'C:\msys64\ucrt64\bin\gcc.exe'
    if (Test-Path $msysGcc) {
        $env:PATH = "C:\msys64\ucrt64\bin;$env:PATH"
        $gcc = Get-Command gcc -ErrorAction SilentlyContinue
    }
}

if (-not $gcc) {
    throw @'
A C compiler (gcc) is required for the SQLite build of Vikunja.

Install MSYS2:
  winget install -e --id MSYS2.MSYS2

Then open "MSYS2 UCRT64" and run:
  pacman -Syu
  pacman -S mingw-w64-ucrt-x86_64-gcc

Then reopen PowerShell and run this script again.
'@
}

$env:CGO_ENABLED = '1'
$env:CC = $gcc.Source

if (-not $SkipFrontend) {
    Write-Host '=== Building frontend ===' -ForegroundColor Cyan
    Push-Location frontend
    try {
        '{"VERSION":"v2.5.0-custom"}' | Set-Content -Encoding UTF8 src/version.json
        pnpm build
        if ($LASTEXITCODE -ne 0) {
            throw "Frontend build failed with exit code $LASTEXITCODE"
        }
    }
    finally {
        Pop-Location
    }
}
else {
    if (-not (Test-Path 'frontend\dist\index.html')) {
        throw 'frontend\dist is missing, so -SkipFrontend cannot be used.'
    }
    Write-Host '=== Reusing existing frontend\dist ===' -ForegroundColor DarkGray
}

Write-Host '=== Building Vikunja executable with embedded frontend ===' -ForegroundColor Cyan
Write-Host "Using C compiler: $($gcc.Source)" -ForegroundColor DarkGray

# IMPORTANT:
# Use comma-separated tags and = syntax so Windows PowerShell cannot split
# "osusergo netgo" into separate native-process arguments.
& go build -trimpath "-tags=osusergo,netgo" "-o=vikunja.exe" .
if ($LASTEXITCODE -ne 0) {
    throw "Go build failed with exit code $LASTEXITCODE"
}

if (-not (Test-Path '.\vikunja.exe')) {
    throw 'Build finished without producing vikunja.exe.'
}

New-Item -ItemType Directory -Force -Path release | Out-Null
Copy-Item -Force .\vikunja.exe .\release\vikunja.exe
Copy-Item -Force .\config.yml .\release\config.yml
New-Item -ItemType Directory -Force -Path .\release\data | Out-Null
New-Item -ItemType Directory -Force -Path .\release\database | Out-Null

@'
@echo off
cd /d "%~dp0"
vikunja.exe web
pause
'@ | Set-Content -Encoding ASCII .\release\RUN.cmd

Write-Host ''
Write-Host 'READY: release\vikunja.exe' -ForegroundColor Green
Write-Host 'Run release\RUN.cmd and open Vikunja through this PC address (localhost, LAN IP, public IP or domain).'
