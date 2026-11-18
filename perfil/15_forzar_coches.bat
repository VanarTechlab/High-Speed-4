@echo off
rem Con el juego abierto por 13_juego_debug.bat: fuerza rivales IA en la carrera de 2 jugadores
cd /d "%~dp0..\build-windows"
set /p N=Rivales IA (Enter = 2): 
if "%N%"=="" set N=2
"overlay_toolchain\python\python.exe" "..\perfil\forzar.py" %N%
pause
