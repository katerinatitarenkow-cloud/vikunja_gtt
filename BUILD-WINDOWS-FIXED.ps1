$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

function Require-Command($name, $hint) {
    if (-not (Get-Command $name -ErrorAction SilentlyContinue)) {
        throw "$name is not installed. $hint"
    }
}

Require-Command 'go' 'Install Go 1.26.4 or newer.'
Require-Command 'pnpm' 'Run .\SETUP-DEV.ps1 first.'

if (-not (Test-Path 'frontend\node_modules')) {
    throw 'Frontend dependencies are missing. Run .\SETUP-DEV.ps1 first.'
}

$version = 'v2.5.0-custom'
$tags = 'osusergo netgo'

# Vikunja uses github.com/mattn/go-sqlite3. For the supplied SQLite config,
# a real CGO-enabled Windows build needs a C compiler.
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

Then open "MSYS2 UCRT64" from the Start menu and run:
  pacman -Syu
  pacman -S mingw-w64-ucrt-x86_64-gcc

After that, close/reopen PowerShell and run this script again.
The script will automatically use C:\msys64\ucrt64\bin\gcc.exe.
'@
}

$env:CGO_ENABLED = '1'
$env:CC = $gcc.Source

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

Write-Host '=== Building Vikunja executable with embedded frontend ===' -ForegroundColor Cyan
Write-Host "Using C compiler: $($gcc.Source)" -ForegroundColor DarkGray

$ldflags = '-s -w -X "code.vikunja.io/api/pkg/version.Version=' + $version + '" -X "main.Tags=' + $tags + '"'
& go build -tags $tags -ldflags $ldflags -o vikunja.exe .
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

@'
@echo off
cd /d "%~dp0"
vikunja.exe web
pause
'@ | Set-Content -Encoding ASCII .\release\RUN.cmd

Write-Host ''
Write-Host 'READY: release\vikunja.exe' -ForegroundColor Green
Write-Host 'Run release\RUN.cmd and open http://localhost:3456'
