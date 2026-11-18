@echo off
rem Guarda la RAM durante la primera pausa (3 fotos en 10 s) en perfil\titulo\
cd /d "%~dp0..\build-windows"
set PSX_DEBUG_SERVER=1
start "" /b "overlay_toolchain\python\python.exe" "..\perfil\titulo.py"
NFS_High_Stakes__Recompiled.exe > ..\perfil\titulo_juego.log 2>&1
echo Terminado.
pause
