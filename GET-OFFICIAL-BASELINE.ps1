$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

$url = 'https://dl.vikunja.io/vikunja/v2.5.0/vikunja-v2.5.0-windows-4.0-amd64.exe-full.zip'
$zip = Join-Path $PSScriptRoot 'baseline-v2.5.0.zip'
$dest = Join-Path $PSScriptRoot 'baseline'

Write-Host 'Downloading official Vikunja 2.5.0 Windows x64 baseline...' -ForegroundColor Cyan
Invoke-WebRequest -Uri $url -OutFile $zip
if (Test-Path $dest) { Remove-Item -Recurse -Force $dest }
Expand-Archive -Path $zip -DestinationPath $dest
Remove-Item $zip
Write-Host "Ready: $dest" -ForegroundColor Green
