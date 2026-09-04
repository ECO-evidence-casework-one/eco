@echo off
setlocal
cd /d "%~dp0\..\.."
powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File "%~dp0acquire-donors.ps1"
set RC=%ERRORLEVEL%
echo.
if %RC%==0 (
  echo ECO FOSS donor acquisition completed successfully.
) else (
  echo ECO FOSS donor acquisition ended with errors. Exit code: %RC%
)
pause
exit /b %RC%
