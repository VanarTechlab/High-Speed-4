@echo off
rem Con el juego abierto por 13_juego_debug.bat: configuracion con 4 humanos (prueba de vistas). DESACTIVA el mod 4 jugadores.
cd /d "%~dp0..\build-windows"
"overlay_toolchain\python\python.exe" "..\perfil\forzar_ia.py" hum4
pause
