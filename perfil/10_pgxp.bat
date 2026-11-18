@echo off
rem Prueba PGXP (activa "PGXP Precision" en el launcher antes). Log de cobertura en perfil\pgxp.log
cd /d "%~dp0..\build-windows"
set PSX_DEBUG_SERVER=1
set PSX_FRAME_REPORT_MS=1000
set PSX_RUNTIME_PERF_DIAG=1
set PSX_RUNTIME_PERF_DIAG_MS=2000
start "" /b "overlay_toolchain\python\python.exe" "..\perfil\pgxp.py"
NFS_High_Stakes__Recompiled.exe > ..\perfil\pgxp_juego.log 2>&1
echo Terminado.
pause
