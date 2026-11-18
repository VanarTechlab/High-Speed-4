@echo off
rem Igual que 15_forzar_coches.bat, pero ademas corrige el estado de la carrera los primeros 5 s
cd /d "%~dp0..\build-windows"
"overlay_toolchain\python\python.exe" "..\perfil\forzar.py" 2 corregir
pause
