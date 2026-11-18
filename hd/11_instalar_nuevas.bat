@echo off
rem Copia a replacements solo las texturas del pack anime que faltan o son mas nuevas (no borra nada).
cd /d "%~dp0"
set DEST=..\build-windows\mods\texture-packs\SLUS-00826\replacements
xcopy /d /y /q pack_anime\*.png "%DEST%\"
echo Listo: texturas nuevas instaladas.
pause
