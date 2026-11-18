@echo off
rem Jukebox fase 1: juega una carrera; pausa varias veces; se vigila quien cambia el tema
cd /d "%~dp0..\build-windows"
set PSX_DEBUG_SERVER=1
start "" /b "overlay_toolchain\python\python.exe" "..\perfil\musica2.py"
NFS_High_Stakes__Recompiled.exe > ..\perfil\musica_juego.log 2>&1
echo Terminado.
pause
