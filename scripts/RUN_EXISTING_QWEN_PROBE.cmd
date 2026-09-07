@echo off
setlocal
cd /d "%~dp0"
title ECO Existing Qwen Offline Probe
echo.
echo ECO Existing Qwen Offline Probe
echo -------------------------------
echo This does NOT download the model again.
echo It finds the full verified Qwen model already under E:\ECO_RIG_AI_PREVIEW*,
echo then runs one corrected offline non-interactive generation with a hard timeout.
echo.
powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File "%~dp0probe-existing-rig-qwen.ps1" -SearchRoot "E:\" -TimeoutSeconds 120
set "ECO_EXIT=%ERRORLEVEL%"
echo.
if "%ECO_EXIT%"=="0" (
  echo Qwen probe completed successfully.
) else (
  echo Qwen probe stopped with an error. Send this window screenshot back to the ECO development chat.
)
echo.
pause
exit /b %ECO_EXIT%
