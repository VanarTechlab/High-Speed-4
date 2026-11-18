@echo off
rem Hace una foto de la RAM del juego abierto con 13_juego_debug.bat. Pide un nombre.
cd /d "%~dp0..\build-windows"
set /p NOMBRE=Nombre de la foto (sin espacios): 
"overlay_toolchain\python\python.exe" "..\perfil\foto.py" %NOMBRE%
timeout /t 2 >nul
