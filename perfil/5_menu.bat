@echo off
rem Menu de pausa: guarda la RAM cada 2 s (unos 2,5 minutos) en perfil\menu\
cd /d "%~dp0..\build-windows"
set PSX_DEBUG_SERVER=1
start "" /b "overlay_toolchain\python\python.exe" "..\perfil\menu.py"
NFS_High_Stakes__Recompiled.exe > ..\perfil\menu_juego.log 2>&1
echo Terminado.
pause
