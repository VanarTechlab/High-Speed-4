@echo off
rem Captura los comandos de dibujo de una carrera de 2 jugadores (build-perfil, con herramientas de depuracion). Resultado: perfil\vistas_frame*.json
cd /d "%~dp0..\build-perfil"
set PSX_DEBUG_SERVER=1
start "" /b "..\build-windows\overlay_toolchain\python\python.exe" "..\perfil\vistas.py"
NFS_High_Stakes__Recompiled.exe > ..\perfil\vistas_juego.log 2>&1
echo Terminado.
pause
