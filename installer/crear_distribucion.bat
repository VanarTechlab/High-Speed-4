@echo off
rem Crea el zip de distribucion de High Speed 4 (estilo Belial 64):
rem   dist\HighSpeed4_vX\  Instalar High Speed 4.exe, instalar_hs4.exe, LEEME.txt, herramientas\fondo.bmp,
rem                         juego\ (el paquete de package_title_release) y LICENCIAS\ (con el codigo del instalador)
rem Antes: psxrecomp\tools\package_title_release.ps1 --channel public --platform windows  (deja el zip en dist\)
cd /d "%~dp0.."
set ZIP=
for %%f in (dist\HighSpeed4-v*-windows-x64.zip) do set ZIP=%%f
if "%ZIP%"=="" (echo No hay dist\HighSpeed4-v*-windows-x64.zip: primero empaqueta el juego. & pause & exit /b 1)
for /f "usebackq delims=" %%v in ("VERSION") do set VER=%%v
set OUT=dist\HighSpeed4_v%VER%
if exist "%OUT%" rmdir /s /q "%OUT%"
if exist dist\_tmp_hs4 rmdir /s /q dist\_tmp_hs4
mkdir "%OUT%\herramientas" "%OUT%\LICENCIAS\instalador_codigo"
powershell -NoProfile -Command "Expand-Archive -Force -LiteralPath '%ZIP%' -DestinationPath 'dist\_tmp_hs4'"
if exist dist\_tmp_hs4\HighSpeed4.exe (move dist\_tmp_hs4 "%OUT%\juego" >nul) else (for /d %%d in (dist\_tmp_hs4\*) do if exist "%%d\HighSpeed4.exe" move "%%d" "%OUT%\juego" >nul)
if exist dist\_tmp_hs4 rmdir /s /q dist\_tmp_hs4
if not exist "%OUT%\juego\HighSpeed4.exe" (echo El zip no trae HighSpeed4.exe & pause & exit /b 1)
copy /y "installer\bin\Instalar High Speed 4.exe" "%OUT%\" >nul
copy /y "installer\bin\instalar_hs4.exe" "%OUT%\" >nul
copy /y "installer\herramientas\fondo.bmp" "%OUT%\herramientas\" >nul
powershell -NoProfile -Command "(Get-Content -Raw 'installer\LEEME.txt') -replace '@VERSION@','%VER%' | Set-Content -NoNewline -Encoding UTF8 '%OUT%\LEEME.txt'"
copy /y LICENSE "%OUT%\LICENCIAS\LICENSE_HighSpeed4_MIT.txt" >nul
copy /y LICENCIAS.md "%OUT%\LICENCIAS\" >nul
copy /y installer\go\*.go "%OUT%\LICENCIAS\instalador_codigo\" >nul
copy /y installer\go\go.mod "%OUT%\LICENCIAS\instalador_codigo\" >nul
copy /y installer\fondo.py "%OUT%\LICENCIAS\instalador_codigo\" >nul
if exist "dist\HighSpeed4_v%VER%.zip" del "dist\HighSpeed4_v%VER%.zip"
powershell -NoProfile -Command "Compress-Archive -Path '%OUT%' -DestinationPath 'dist\HighSpeed4_v%VER%.zip'"
echo Listo: dist\HighSpeed4_v%VER%.zip
pause
