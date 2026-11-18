@echo off
rem Perfil para el port a Vita (mide 30 s de carrera de 1 jugador). E = normal, L = ligera (res. nativa, sin texturas HD).
rem Resultado: perfil\vita_E_*.json / vita_L_*.json y el log del juego perfil\vita_juego_E.log / _L.log
cd /d "%~dp0..\build-windows"
choice /c EL /m "E = normal, L = ligera"
if errorlevel 2 goto L
set MODO=E
goto RUN
:L
set MODO=L
set PSX_INTERNAL_RESOLUTION=native
:RUN
set PSX_DEBUG_SERVER=1
set PSX_FRAME_REPORT_MS=1000
set PSX_RUNTIME_PERF_DIAG=1
set PSX_RUNTIME_PERF_DIAG_MS=2000
start "perfil vita" "overlay_toolchain\python\python.exe" "..\perfil\vita_perfil.py" %MODO%
NFS_High_Stakes__Recompiled.exe > ..\perfil\vita_juego_%MODO%.log 2>&1
echo Terminado.
pause
