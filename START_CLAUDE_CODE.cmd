@echo off
setlocal
cd /d "%~dp0"
where claude >nul 2>nul
if errorlevel 1 (
  echo Claude Code was not found on PATH
  echo Install or repair Claude Code, then run this file again
  pause
  exit /b 1
)
echo Starting Claude Code in:
echo %CD%
claude
endlocal
