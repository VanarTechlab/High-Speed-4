@echo off
rem Abre el juego con servidor de depuracion (para hacer fotos de RAM con 14_foto.bat)
cd /d "%~dp0..\build-windows"
set PSX_DEBUG_SERVER=1
NFS_High_Stakes__Recompiled.exe > ..\perfil\fotos_juego.log 2>&1
