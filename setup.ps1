$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$pnpm = 'C:\Users\rewzh96\.cache\codex-runtimes\codex-primary-runtime\dependencies\bin\fallback\pnpm.cmd'
if (-not (Test-Path (Join-Path $root '.tools\go\bin\go.exe'))) { throw '便携 Go 尚未下载到 .tools\go。' }
& $pnpm --dir (Join-Path $root 'frontend') install
Write-Host '环境准备完成。运行 .\start.ps1 启动系统。'
