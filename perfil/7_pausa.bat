@echo off
rem Busca el indicador de pausa: guarda la RAM cada segundo (~70 s) en perfil\pausa2\ (empieza al entrar en carrera)
cd /d "%~dp0..\build-windows"
set PSX_DEBUG_SERVER=1
start "" /b "overlay_toolchain\python\python.exe" "..\perfil\pausa.py"
NFS_High_Stakes__Recompiled.exe > ..\perfil\pausa_juego.log 2>&1
echo Terminado.
pause
