@echo off
rem Prueba del instalador nuevo ANTES de empaquetar: monta build-prueba-dist como si fuera el zip descomprimido
rem (juego\ = copia de build-windows sin cache ni registros, con el exe renombrado a HighSpeed4.exe) y abre el instalador.
rem En el instalador elige como destino una carpeta de prueba, por ejemplo D:\Proyectos\NFS4\prueba_instalacion.
cd /d "%~dp0.."
set D=build-prueba-dist
if exist %D% rmdir /s /q %D%
mkdir %D%\herramientas
robocopy build-windows %D%\juego /E /NFL /NDL /NJH /NJS /XD cache texture-packs nfshs.unlocks nfshs.4p overlay_captures.json.d native-cpu-spans native-idle-wait CMakeFiles _deps psx_execution recomp-net retcomm-rbengine saves /XF state.toml *.log *.json.d overlay_captures.json settings.toml *.ninja* CMakeCache.txt *.cmake *.a psx_freeze_* psx_crash.txt *.bat >nul
move /y %D%\juego\NFS_High_Stakes__Recompiled.exe %D%\juego\HighSpeed4.exe >nul
mkdir %D%\juego\hs4_aot\psxrecomp\tools\aot_overlay_spike
copy /y "installer\Preparar juego.bat" %D%\juego\ >nul
copy /y installer\hs4_preparar.py %D%\juego\hs4_aot\ >nul
copy /y aot\overlays.json %D%\juego\hs4_aot\ >nul
copy /y psxrecomp\tools\aligned_lzss_banks.py %D%\juego\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\aot_overlay_pipeline.py %D%\juego\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\aot_overlay_spike\extract_generic.py %D%\juego\hs4_aot\psxrecomp\tools\aot_overlay_spike\ >nul
copy /y psxrecomp\tools\audit_aot_cache.py %D%\juego\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\bpe_container.py %D%\juego\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\compile_overlays.py %D%\juego\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\counted_mips_relocations.py %D%\juego\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\extract_overlays.py %D%\juego\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\indexed_lzss_pack.py %D%\juego\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\mips_tagged_relocations.py %D%\juego\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\mod_package_images.py %D%\juego\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\packed_sector_table.py %D%\juego\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\prepared_files.py %D%\juego\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\release_stage.py %D%\juego\hs4_aot\psxrecomp\tools\ >nul
copy /y psxrecomp\tools\sector_extent_archive.py %D%\juego\hs4_aot\psxrecomp\tools\ >nul
rem Kit HD dentro del juego (como en el paquete)
mkdir %D%\juego\hd_kit\realesrgan\models
copy /y "installer\bin\Texturas HD.exe" %D%\juego\ >nul
copy /y installer\bin\texturas_hd.exe %D%\juego\ >nul
copy /y installer\hd_kit\fondo.bmp %D%\juego\hd_kit\ >nul
copy /y installer\hd_kit\decisiones.txt %D%\juego\hd_kit\ >nul
copy /y installer\hd_kit\excluidas.txt %D%\juego\hd_kit\ >nul
copy /y installer\hd_kit\LICENCIAS_kit_HD.txt %D%\juego\hd_kit\ >nul
copy /y installer\hd_kit\realesrgan\*.* %D%\juego\hd_kit\realesrgan\ >nul
copy /y installer\hd_kit\modelo\*.* %D%\juego\hd_kit\realesrgan\models\ >nul
copy /y "installer\bin\Instalar High Speed 4.exe" %D%\ >nul
copy /y installer\bin\instalar_hs4.exe %D%\ >nul
copy /y installer\herramientas\fondo.bmp %D%\herramientas\ >nul
copy /y installer\LEEME.txt %D%\ >nul
start "" "%D%\Instalar High Speed 4.exe"
echo Instalador abierto. Usa una carpeta de PRUEBA como destino.
pause
