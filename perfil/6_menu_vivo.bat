@echo off
rem Abre el juego y una ventana con los valores del menu de pausa en directo.
cd /d "%~dp0..\build-windows"
set PSX_DEBUG_SERVER=1
start "Valores menu de pausa" cmd /k ""overlay_toolchain\python\python.exe" "..\perfil\vivo.py""
NFS_High_Stakes__Recompiled.exe > ..\perfil\menu_vivo.log 2>&1
