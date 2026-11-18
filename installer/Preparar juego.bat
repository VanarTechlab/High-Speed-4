@echo off
rem High Speed 4: traduce y compila el codigo del juego desde TU disco (unos 15-30 minutos la primera vez).
cd /d "%~dp0"
title High Speed 4 - Preparar juego
"overlay_toolchain\python\python.exe" "hs4_aot\hs4_preparar.py" %*
echo.
pause
