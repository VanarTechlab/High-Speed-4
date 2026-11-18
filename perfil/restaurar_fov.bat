@echo off
rem Vuelve a poner los parches de campo de vision (pantalla ancha) en game.toml
cd /d "%~dp0.."
copy /y analysis\game.toml.con_fov game.toml
copy /y analysis\build-windows_game.toml.con_fov build-windows\game.toml
echo Restaurado. Ahora compila (rebuild) y ejecuta build-windows\compilar_overlays.bat
pause
