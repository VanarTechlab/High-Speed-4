@echo off
rem Mapa de coches para el multijugador: guarda la RAM cada segundo (60 s) en perfil\coches\ desde que empieza la carrera
cd /d "%~dp0..\build-windows"
set PSX_DEBUG_SERVER=1
start "" /b "overlay_toolchain\python\python.exe" "..\perfil\coches.py"
NFS_High_Stakes__Recompiled.exe > ..\perfil\coches_juego.log 2>&1
echo Terminado.
pause
