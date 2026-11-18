@echo off
rem Paso 2: compila los overlays de la version de perfil (usa la misma lista que build-windows)
cd /d "%~dp0..\build-perfil"
if not exist NFS_High_Stakes__Recompiled.exe (echo Falta compilar build-perfil primero & pause & exit /b 1)
copy /y "..\build-windows\overlay_captures.json" "overlay_captures.json" >nul
copy /y "..\build-windows\compilar_overlays.bat" "compilar_overlays.bat" >nul
call compilar_overlays.bat
