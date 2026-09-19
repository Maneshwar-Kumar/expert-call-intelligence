Set-Location $PSScriptRoot
if (-not (Get-Command claude -ErrorAction SilentlyContinue)) {
    Write-Error "Claude Code was not found on PATH"
    exit 1
}
Write-Host "Starting Claude Code in $PWD"
claude
