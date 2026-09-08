$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$node = 'C:\Users\rewzh96\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe'
$pnpm = 'C:\Users\rewzh96\.cache\codex-runtimes\codex-primary-runtime\dependencies\bin\fallback\pnpm.cmd'
$go = Join-Path $root '.tools\go\bin\go.exe'
$env:GOCACHE = Join-Path $root '.cache\go-build'
if (-not (Test-Path $go)) { throw '缺少便携 Go 运行环境，请先运行 setup.ps1。' }
if (-not (Test-Path (Join-Path $root 'frontend\node_modules'))) { & $pnpm --dir (Join-Path $root 'frontend') install }
& $pnpm --dir (Join-Path $root 'frontend') build
Set-Location (Join-Path $root 'backend')
& $go run .
