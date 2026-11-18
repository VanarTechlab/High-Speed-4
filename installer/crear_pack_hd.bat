@echo off
rem Crea el pack HD opcional (descarga APARTE del juego): dist\HighSpeed4_TexturasHD_vX.zip
rem con hd\pack_anime (todas las texturas HD) y LEEME_TexturasHD.txt. Se instala con "Texturas HD.exe" -> A).
rem Recomendado: subirlo FUERA del repositorio del codigo (riesgo de retirada por ser arte derivado de EA).
cd /d "%~dp0.."
for /f "usebackq delims=" %%v in ("VERSION") do set VER=%%v
if not exist hd\pack_anime\*.png (echo Falta hd\pack_anime & pause & exit /b 1)
if not exist dist mkdir dist
set ZIP=dist\HighSpeed4_TexturasHD_v%VER%.zip
if exist "%ZIP%" del "%ZIP%"
echo Comprimiendo el pack HD (unos 730 MB, puede tardar unos minutos)...
"%SystemRoot%\System32\tar.exe" -a -c -f "%ZIP%" -C installer\pack_hd LEEME_TexturasHD.txt -C ..\..\hd pack_anime
if errorlevel 1 (echo ERROR al crear el zip & pause & exit /b 1)
echo Listo: %ZIP%
pause
