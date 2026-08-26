$ErrorActionPreference = "Stop"
Set-Location "$PSScriptRoot\frontend"

$env:DEV_PROXY = "http://127.0.0.1:3457"
$env:VIKUNJA_FRONTEND_PORT = "4173"
$env:VIKUNJA_FRONTEND_BASE = "/"

pnpm dev --host 0.0.0.0
