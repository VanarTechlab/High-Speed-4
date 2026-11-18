@echo off
rem Traza quien escribe el estado de carrera con 4 coches forzados (usa build-perfil, mas lento). Resultado: perfil\traza_carrera.json
cd /d "%~dp0..\build-perfil"
set PSX_DEBUG_SERVER=1
start "" /b "..\build-windows\overlay_toolchain\python\python.exe" "..\perfil\traza.py"
NFS_High_Stakes__Recompiled.exe > ..\perfil\traza_juego.log 2>&1
echo Terminado.
pause
