@echo off
rem Instala el pack HD anime (hd\pack_anime): vacia replacements y copia el pack completo.
cd /d "%~dp0"
set DEST=..\build-windows\mods\texture-packs\SLUS-00826\replacements
if not exist pack_anime\*.png ( echo Falta hd\pack_anime & pause & exit /b 1 )
if exist "%DEST%" rmdir /s /q "%DEST%"
mkdir "%DEST%"
echo Copiando 21192 texturas...
xcopy /y /q pack_anime\*.png "%DEST%\" >nul
echo Instalado el pack HD anime en %DEST%
pause
