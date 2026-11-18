@echo off
rem Prueba de "Preparar juego" como lo vera un jugador: monta una carpeta instalada falsa en build-prueba-instalado
rem (overlay_toolchain + game.toml + settings.toml con tu disco + hs4_aot) y ejecuta hs4_preparar.py con el python incluido.
rem Registro: perfil\preparar_prueba.log. Tarda ~15-30 min.
cd /d "%~dp0.."
set D=build-prueba-instalado
if exist %D% rmdir /s /q %D%
mkdir %D%\hs4_aot\psxrecomp\tools\aot_overlay_spike
robocopy build-windows\overlay_toolchain %D%\overlay_toolchain /E /NFL /NDL /NJH /NJS >nul
copy /y build-windows\game.toml %D%\game.toml >nul
copy /y installer\hs4_preparar.py %D%\hs4_aot\ >nul
copy /y aot\overlays.json %D%\hs4_aot\ >nul
copy /y psxrecomp\tools\aligned_lzss_banks.py %D%\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\aot_overlay_pipeline.py %D%\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\aot_overlay_spike\extract_generic.py %D%\hs4_aot\psxrecomp\tools\aot_overlay_spike\ >nul
copy /y psxrecomp\tools\audit_aot_cache.py %D%\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\bpe_container.py %D%\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\compile_overlays.py %D%\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\counted_mips_relocations.py %D%\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\extract_overlays.py %D%\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\indexed_lzss_pack.py %D%\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\mips_tagged_relocations.py %D%\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\mod_package_images.py %D%\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\packed_sector_table.py %D%\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\prepared_files.py %D%\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\release_stage.py %D%\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\sector_extent_archive.py %D%\hs4_aot\psxrecomp\tools\ >nul
echo [disc]> %D%\settings.toml
echo path = "D:/Proyectos/NFS4/disco/Need for Speed - High Stakes (USA).cue">> %D%\settings.toml
%D%\overlay_toolchain\python\python.exe %D%\hs4_aot\hs4_preparar.py > perfil\preparar_prueba.log 2>&1
echo Codigo de salida: %ERRORLEVEL%
type perfil\preparar_prueba.log | more
echo Terminado. Registro: perfil\preparar_prueba.log
pause
