@echo off
rem Escala TODAS las texturas con el modelo anime x4 (no inventa trama en superficies lisas) en hd\salida_anime_todo
cd /d "%~dp0"
if not exist salida_anime_todo mkdir salida_anime_todo
echo Escalando 22276 texturas con modelo anime... no cierres esta ventana. Empezo: %time%
realesrgan\realesrgan-ncnn-vulkan.exe -i ..\build-windows\mods\texture-packs\SLUS-00826\dumps -o salida_anime_todo -n realesrgan-x4plus-anime -s 4 -f png
echo Termino: %time%
pause
