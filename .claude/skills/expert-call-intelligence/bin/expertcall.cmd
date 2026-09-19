@echo off
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0expertcall.ps1" %*
exit /b %ERRORLEVEL%
