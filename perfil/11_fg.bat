@echo off
rem Prueba Smooth motion (frame generation) y anota por que no genera imagenes. Log en perfil\fg.log
cd /d "%~dp0..\build-windows"
set PSX_DEBUG_SERVER=1
set PSX_RENDER_THREAD=1
set PSX_FRAME_GEN=1
set PSX_FRAME_REPORT_MS=1000
start "" /b "overlay_toolchain\python\python.exe" "..\perfil\fg.py"
NFS_High_Stakes__Recompiled.exe > ..\perfil\fg_juego.log 2>&1
echo Terminado.
pause
