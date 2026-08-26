$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$env:VIKUNJA_SERVICE_INTERFACE = "127.0.0.1:3457"
$env:CGO_ENABLED = "1"
$env:CC = "C:\msys64\ucrt64\bin\gcc.exe"
$env:PATH = "C:\msys64\ucrt64\bin;$env:PATH"

$air = Join-Path (go env GOPATH) "bin\air.exe"

if (-not (Test-Path $air)) {
    throw "Air not found: $air"
}

& $air -c ".air.toml"
