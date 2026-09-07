@echo off
setlocal
cd /d "%~dp0"
title ECO Rig AI Preview Setup
echo.
echo ECO Rig AI Preview Setup
echo ------------------------
echo This builds the exact qualified ECO source, installs the pinned local AI assets,
echo tests Qwen offline, then opens an isolated developer preview.
echo.
echo No administrator elevation or Windows security changes are made.
echo.
powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File "%~dp0prepare-rig-ai-preview-v3.ps1"
set "ECO_EXIT=%ERRORLEVEL%"
echo.
if not "%ECO_EXIT%"=="0" (
  echo Setup stopped. Open ECO_RIG_AI_PREVIEW\AI_SETUP_RESULT.txt and send it back to the ECO development chat.
) else (
  echo Setup completed. ECO should now be open with the verified local AI configuration.
)
echo.
pause
exit /b %ECO_EXIT%
