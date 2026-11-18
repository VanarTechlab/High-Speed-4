@echo off
rem Traza que mando lee cada jugador con 4 humanos forzados (build-perfil). Resultado: perfil\traza_mandos.json
cd /d "%~dp0..\build-perfil"
set PSX_DEBUG_SERVER=1
start "" /b "..\build-windows\overlay_toolchain\python\python.exe" "..\perfil\traza_mandos.py"
NFS_High_Stakes__Recompiled.exe > ..\perfil\traza_mandos_juego.log 2>&1
echo Terminado.
pause
