#!/bin/bash
# High Speed 4: compila el codigo del juego desde el disco con el perfil AOT (aot/overlays.json).
# Uso (MSYS2 MINGW64, desde la carpeta del proyecto):
#   bash scripts/hs4_aot.sh extract   -> solo comprueba el disco y cuenta las recetas
#   bash scripts/hs4_aot.sh release   -> compila la cache en build-aot/stage (con clang)
# Registro completo en build-aot/hs4_aot_<accion>.log
set -u
cd "$(dirname "$0")/.."
ACCION="${1:-extract}"
DISCO="${HS4_DISCO:-D:/Proyectos/NFS4/disco/Need for Speed - High Stakes (USA).cue}"
CLANG="${HS4_CLANG:-$(cat build-windows/overlay_toolchain/compiler.txt 2>/dev/null)}"
mkdir -p build-aot
EXTRA=()
if [ "$ACCION" = "release" ]; then
  rm -rf build-aot/stage 2>/dev/null
  EXTRA=(--stage build-aot/stage --gcc "$CLANG" --workers 6 --cps --runtime-config build-windows/game.toml)
fi
python psxrecomp/tools/aot_overlay_pipeline.py "$ACCION" \
  --profile aot/overlays.json --game-toml game.toml --disc "$DISCO" \
  --recompiler build-windows/overlay_toolchain/psxrecomp-game.exe \
  --work-dir build-aot "${EXTRA[@]}" 2>&1 | tee "build-aot/hs4_aot_${ACCION}.log"
echo "Terminado ($ACCION). Registro: build-aot/hs4_aot_${ACCION}.log"
