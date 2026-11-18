@echo off
rem Con el juego abierto por 13_juego_debug.bat: 2 jugadores + 2 rivales IA (prueba 3)
cd /d "%~dp0..\build-windows"
"overlay_toolchain\python\python.exe" "..\perfil\forzar_ia.py"
pause
