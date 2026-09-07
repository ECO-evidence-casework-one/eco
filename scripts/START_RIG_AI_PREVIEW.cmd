@echo off
setlocal
cd /d "%~dp0"
title ECO Rig AI Preview Setup
echo.
echo ECO Rig AI Preview Setup
echo ------------------------
echo This builds the exact controlled ECO source, installs the pinned local AI assets,
echo tests Qwen offline, then opens an isolated developer preview.
echo.
echo No administrator elevation or Windows security changes are made.
echo.
set "ECO_OUTPUT=%~dp0ECO_RIG_AI_PREVIEW"
if exist "E:\" set "ECO_OUTPUT=E:\ECO_RIG_AI_PREVIEW"
set "ECO_POINTER=%ECO_OUTPUT%.LAST_OUTPUT.txt"
set "ECO_ACTUAL=%ECO_OUTPUT%"
echo Requested preview location: %ECO_OUTPUT%
echo.
powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File "%~dp0prepare-rig-ai-preview-supervisor.ps1" -OutputRoot "%ECO_OUTPUT%"
set "ECO_EXIT=%ERRORLEVEL%"
if exist "%ECO_POINTER%" set /p ECO_ACTUAL=<"%ECO_POINTER%"
echo.
echo Actual preview location: %ECO_ACTUAL%
if not "%ECO_EXIT%"=="0" (
  echo Setup stopped. Open "%ECO_ACTUAL%\AI_SETUP_RESULT.txt" and send it back to the ECO development chat.
) else (
  echo Setup completed. ECO should now be open with the verified local AI configuration.
)
echo.
pause
exit /b %ECO_EXIT%
