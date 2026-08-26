param(
    [ValidateSet("Start","Stop")]
    [string]$Action = "Start"
)

$ErrorActionPreference = "SilentlyContinue"

$root = $PSScriptRoot
$logDir = Join-Path $root "logs\headless"
$pidFile = Join-Path $root "tmp\headless-pids.json"

New-Item -ItemType Directory -Force -Path $logDir | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $root "tmp") | Out-Null

function Get-PortPid([int]$Port) {
    return Get-NetTCPConnection `
        -LocalPort $Port `
        -State Listen `
        -ErrorAction SilentlyContinue |
        Select-Object -First 1 -ExpandProperty OwningProcess
}

function Stop-Tree([int]$ProcessId) {
    if ($ProcessId -gt 0) {
        & taskkill.exe /PID $ProcessId /T /F *> $null
    }
}

if ($Action -eq "Stop") {

    foreach ($port in @(4173, 3457, 3456)) {
        $processId = Get-PortPid $port

        if ($processId) {
            Stop-Tree $processId
        }
    }

    Remove-Item $pidFile -Force -ErrorAction SilentlyContinue
    exit
}


# ============================================================
# PRODUCTION :3456
# ============================================================

$prodPid = Get-PortPid 3456

if (-not $prodPid) {

    $prod = Start-Process `
        -FilePath (Join-Path $root "release\vikunja.exe") `
        -ArgumentList "web" `
        -WorkingDirectory (Join-Path $root "release") `
        -WindowStyle Hidden `
        -RedirectStandardOutput (Join-Path $logDir "production-out.log") `
        -RedirectStandardError (Join-Path $logDir "production-error.log") `
        -PassThru

    $prodPid = $prod.Id
}


# ============================================================
# DEV BACKEND :3457
# ============================================================

$devBackendPid = Get-PortPid 3457

if (-not $devBackendPid) {

    $backend = Start-Process `
        -FilePath "powershell.exe" `
        -ArgumentList @(
            "-NoLogo",
            "-NoProfile",
            "-ExecutionPolicy", "Bypass",
            "-File", (Join-Path $root "DEV-BACKEND.ps1")
        ) `
        -WorkingDirectory $root `
        -WindowStyle Hidden `
        -RedirectStandardOutput (Join-Path $logDir "dev-backend-out.log") `
        -RedirectStandardError (Join-Path $logDir "dev-backend-error.log") `
        -PassThru

    $devBackendPid = $backend.Id
}


# Немного времени Air на первый запуск
Start-Sleep -Seconds 2


# ============================================================
# DEV FRONTEND :4173
# ============================================================

$devFrontendPid = Get-PortPid 4173

if (-not $devFrontendPid) {

    $frontend = Start-Process `
        -FilePath "powershell.exe" `
        -ArgumentList @(
            "-NoLogo",
            "-NoProfile",
            "-ExecutionPolicy", "Bypass",
            "-File", (Join-Path $root "DEV-FRONTEND.ps1")
        ) `
        -WorkingDirectory $root `
        -WindowStyle Hidden `
        -RedirectStandardOutput (Join-Path $logDir "dev-frontend-out.log") `
        -RedirectStandardError (Join-Path $logDir "dev-frontend-error.log") `
        -PassThru

    $devFrontendPid = $frontend.Id
}


@{
    production = $prodPid
    devBackend = $devBackendPid
    devFrontend = $devFrontendPid
    started = (Get-Date).ToString("s")
} |
ConvertTo-Json |
Set-Content $pidFile -Encoding UTF8


