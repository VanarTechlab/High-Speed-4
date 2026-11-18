@echo off
rem Con el juego abierto por 13_juego_debug.bat: 4 humanos + controles de prueba para los jugadores 3 y 4
cd /d "%~dp0..\build-windows"
"overlay_toolchain\python\python.exe" "..\perfil\forzar_ia.py" controles
pause
