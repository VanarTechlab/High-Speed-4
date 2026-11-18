@echo off
rem Paso 3: abre el juego con el perfilador y mide SOLO cada 10 segundos.
rem No hace falta Alt+Tab: deja esta ventana abierta y juega.
cd /d "%~dp0..\build-perfil"
set PSX_DEBUG_SERVER=1
set PSX_GUEST_CYCLE_SCALE=2
set PSX_GCS_BATCH=1
set PSX_FRAME_REPORT_MS=1000
set PSX_RUNTIME_PERF_DIAG=1
set PSX_RUNTIME_PERF_DIAG_MS=2000
set NM=C:\Users\vanar\.local\share\retcomm\toolchains\cmake-clang-v1\latest\bin\llvm-nm.exe
set PY=..\build-windows\overlay_toolchain\python\python.exe
set OUT=..\perfil\resultado.txt
echo Perfil NFS4 > "%OUT%"
start "" /b cmd /c "NFS_High_Stakes__Recompiled.exe > perfil_juego.log 2>&1"
echo Juego abierto. Midiendo cada 10 segundos en perfil\resultado.txt
echo Esta ventana se cierra sola cuando cierres el juego.
timeout /t 20 /nobreak >nul
set N=0
:BUCLE
tasklist /fi "imagename eq NFS_High_Stakes__Recompiled.exe" | find /i "NFS_High_Stakes" >nul || goto FIN
set /a N+=10
echo.>> "%OUT%"
echo ===== medida %N% s (ultimos 10 s) =====>> "%OUT%"
"%PY%" "..\psxrecomp\tools\host_profile.py" --port 4370 --exe "NFS_High_Stakes__Recompiled.exe" --frames 600 --top 30 --nm "%NM%" >> "%OUT%" 2>&1
"%PY%" "..\perfil\gcp.py" >> "%OUT%" 2>&1
timeout /t 10 /nobreak >nul
goto BUCLE
:FIN
echo Terminado. Resultado en perfil\resultado.txt
pause
