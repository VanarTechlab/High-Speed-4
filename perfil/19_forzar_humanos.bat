@echo off
rem Con el juego abierto por 13_juego_debug.bat: 2 jugadores + 2 coches extra de tipo HUMANO (prueba 4)
cd /d "%~dp0..\build-windows"
"overlay_toolchain\python\python.exe" "..\perfil\forzar_ia.py" humano
pause
