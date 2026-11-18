@echo off
rem HD activado + hilo de dibujo aparte; vigila cuando el hilo se para (perfil\hilo.log) y el juego en perfil\hilo_juego.log
cd /d "%~dp0..\build-windows"
set PSX_DEBUG_SERVER=1
set PSX_RENDER_THREAD=1
set PSX_FRAME_REPORT_MS=1000
set PSX_RUNTIME_PERF_DIAG=1
set PSX_RUNTIME_PERF_DIAG_MS=2000
start "" /b "overlay_toolchain\python\python.exe" "..\perfil\hilo.py"
NFS_High_Stakes__Recompiled.exe > ..\perfil\hilo_juego.log 2>&1
echo Terminado.
pause
