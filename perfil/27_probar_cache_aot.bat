@echo off
rem Prueba High Speed 4: el juego usando SOLO la cache compilada desde el disco con el perfil AOT (build-aot\stage\cache),
rem sin compilar en segundo plano (lo que falte va al interprete). Registro: perfil\aot_juego.log (fps cada segundo).
rem Aparta tu cache normal en build-windows\cache_capturas y la devuelve al terminar.
cd /d "%~dp0..\build-windows"
if not exist "..\build-aot\stage\cache\SLUS-00826" (echo Falta build-aot\stage\cache: primero bash scripts/hs4_aot.sh release & pause & exit /b 1)
if exist cache_capturas (echo Ya existe build-windows\cache_capturas: revisa antes de seguir. & pause & exit /b 1)
if exist cache move cache cache_capturas >nul
xcopy /e /i /q "..\build-aot\stage\cache" cache >nul
set PSX_OVERLAY_AUTOCOMPILE_OFF=1
set PSX_FRAME_REPORT_MS=1000
set PSX_RUNTIME_PERF_DIAG=1
set PSX_RUNTIME_PERF_DIAG_MS=2000
NFS_High_Stakes__Recompiled.exe > ..\perfil\aot_juego.log 2>&1
rem Devolver la cache normal
if exist cache move cache ..\build-aot\cache_usada_en_prueba >nul
if exist cache_capturas move cache_capturas cache >nul
echo Terminado. Tu cache normal esta de vuelta en build-windows\cache.
pause
