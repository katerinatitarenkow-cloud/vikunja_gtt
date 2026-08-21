$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

Write-Host '=== Vikunja Custom Base: setup ===' -ForegroundColor Cyan

function Require-Command($name, $hint) {
    if (-not (Get-Command $name -ErrorAction SilentlyContinue)) {
        throw "$name is not installed. $hint"
    }
}

Require-Command 'go' 'Install Go 1.26.4 or newer.'
Require-Command 'node' 'Install Node.js 24 or newer.'
Require-Command 'npm' 'npm must be available with Node.js.'

$nodeVersion = (node --version).TrimStart('v')
$nodeMajor = [int]($nodeVersion.Split('.')[0])
if ($nodeMajor -lt 24) {
    throw "Node.js 24+ is required by this Vikunja tree. Found $nodeVersion"
}

$goVersionRaw = (go version)
if ($goVersionRaw -notmatch 'go(\d+)\.(\d+)(?:\.(\d+))?') {
    throw "Could not parse Go version: $goVersionRaw"
}
$goMajor = [int]$Matches[1]
$goMinor = [int]$Matches[2]
if (($goMajor -lt 1) -or ($goMajor -eq 1 -and $goMinor -lt 26)) {
    throw "Go 1.26.4+ is required by go.mod. Found: $goVersionRaw"
}

Write-Host 'Installing/enabling pnpm 11.19.0 through Corepack...'
npm install -g corepack
corepack enable
corepack prepare pnpm@11.19.0 --activate

Write-Host 'Installing Mage...'
go install github.com/magefile/mage@latest

$goBin = (go env GOPATH) + '\bin'
if ($env:PATH -notlike "*$goBin*") {
    $env:PATH = "$goBin;$env:PATH"
}

Write-Host 'Installing frontend dependencies...'
Push-Location frontend
pnpm install --frozen-lockfile
Pop-Location

New-Item -ItemType Directory -Force -Path data | Out-Null
Write-Host 'Setup complete.' -ForegroundColor Green
Write-Host 'Next: run .\RUN-DEV.ps1'
