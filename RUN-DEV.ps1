$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

$goBin = (go env GOPATH) + '\bin'
if ($env:PATH -notlike "*$goBin*") {
    $env:PATH = "$goBin;$env:PATH"
}

if (-not (Get-Command mage -ErrorAction SilentlyContinue)) {
    throw 'Mage is missing. Run SETUP-DEV.ps1 first.'
}
if (-not (Test-Path 'frontend\node_modules')) {
    throw 'Frontend dependencies are missing. Run SETUP-DEV.ps1 first.'
}

$env:RELEASE_VERSION = 'v2.5.0-custom-dev'
New-Item -ItemType Directory -Force -Path data | Out-Null

Write-Host 'Building backend...' -ForegroundColor Cyan
mage build

Write-Host 'Starting backend on http://localhost:3456 ...' -ForegroundColor Cyan
$backend = Start-Process -FilePath (Join-Path $PSScriptRoot 'vikunja.exe') -ArgumentList 'web' -WorkingDirectory $PSScriptRoot -PassThru

Write-Host 'Starting frontend dev server on http://localhost:4173 ...' -ForegroundColor Cyan
try {
    Push-Location frontend
    pnpm dev
}
finally {
    Pop-Location
    if ($backend -and -not $backend.HasExited) {
        Stop-Process -Id $backend.Id -Force
    }
}
