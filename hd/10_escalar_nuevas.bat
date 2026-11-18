@echo off
rem Escala con el modelo anime x4 las texturas de hd\entrada_nuevas (las prepara Claude con preparar_nuevas.py) en hd\salida_anime_todo
cd /d "%~dp0"
if not exist entrada_nuevas\*.png ( echo No hay texturas en hd\entrada_nuevas. & pause & exit /b 1 )
echo Escalando texturas nuevas... no cierres esta ventana. Empezo: %time%
realesrgan\realesrgan-ncnn-vulkan.exe -i entrada_nuevas -o salida_anime_todo -n realesrgan-x4plus-anime -s 4 -f png
echo Termino: %time%
pause
