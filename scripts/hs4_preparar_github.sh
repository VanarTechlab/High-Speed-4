#!/bin/bash
# High Speed 4: prepara los commits para GitHub SIN subir nada (la subida se hace con GitHub Desktop,
# que pide la cuenta en una ventana normal). No conecta a internet.
#  - psxrecomp: tus parches en un commit de la rama high-speed-4 (se publicara en tu fork)
#  - proyecto: rama "main" NUEVA sin historial (el historial antiguo tiene generated/ con codigo de EA)
#    con el submodulo psxrecomp apuntando a VanarTechlab/psxrecomp y remoto origin = VanarTechlab/High-Speed-4
# Commits a nombre de Vanar con el correo anonimo de GitHub (no tu gmail).
# Uso (MSYS2, desde la carpeta del proyecto), sin preguntas:
#   bash scripts/hs4_preparar_github.sh            -> pasos 1-3 y muestra la lista (NO crea el commit del proyecto)
#   bash scripts/hs4_preparar_github.sh confirmar  -> crea el commit de la rama main (despues de revisar la lista)
set -e
cd "$(dirname "$0")/.."
REPO_JUEGO="https://github.com/VanarTechlab/High-Speed-4.git"
REPO_PSXRECOMP="https://github.com/VanarTechlab/psxrecomp.git"
RAMA_PSX="high-speed-4"
NOMBRE="Vanar"; CORREO="VanarTechlab@users.noreply.github.com"
MODO="${1:-preparar}"
# Archivos que NO se publican aunque esten en la carpeta:
#  mods de 4 jugadores y desbloqueos = solo en tu PC
#  perfil/w_86.txt = volcado de codigo del juego (de EA); .github = CI de plantilla que fallaria sin disco
QUITAR="perfil/w_86.txt .github src/mods/nfshs_4p.c src/mods/nfshs_unlocks.c mods/preloaded/packages/nfshs.4p mods/preloaded/packages/nfshs.unlocks"
quitar() { for q in $QUITAR; do git rm -r -q --cached --ignore-unmatch "$q" >/dev/null; done; }
if [ "$MODO" = "confirmar" ]; then
  [ "$(git symbolic-ref --short HEAD)" = "main" ] || { echo "Primero ejecuta el script sin confirmar."; exit 1; }
  git add -A
  quitar
  git commit -q -m "High Speed 4 v$(cat VERSION)"
  git remote get-url origin >/dev/null 2>&1 || git remote add origin "$REPO_JUEGO"
  git log --oneline -1
  echo "Commit creado. Ahora sube con GitHub Desktop (ver instrucciones del chat)."
  echo "Tu historial antiguo sigue en la rama master y NO se sube."
  exit 0
fi

echo "== 1/3 psxrecomp: guardar tus parches en la rama $RAMA_PSX"
( cd psxrecomp
  git config user.name "$NOMBRE"; git config user.email "$CORREO"
  git status --short | grep -v '^??' || true
  git checkout -B "$RAMA_PSX"
  git add -u
  git commit -q -m "High Speed 4: parches locales (texturas HD, cdrom/jukebox, overlay spans, presets de resolucion)" || echo "(sin cambios nuevos)"
  git log --oneline -1 )

echo "== 2/3 Submodulo psxrecomp -> tu fork"
git config -f .gitmodules submodule.psxrecomp.url "$REPO_PSXRECOMP"
git config -f .gitmodules submodule.psxrecomp.branch "$RAMA_PSX"
git config submodule.psxrecomp.url "$REPO_PSXRECOMP"

echo "== 3/3 Rama main sin historial"
git config user.name "$NOMBRE"; git config user.email "$CORREO"
if [ "$(git symbolic-ref --short HEAD)" != "main" ]; then
  git checkout --orphan main
fi
git rm -r -q --cached generated 2>/dev/null || true
git add -A
quitar
mkdir -p analysis
git diff --cached --name-only > analysis/lista_publicar.txt
echo "--- Lo que se va a publicar, por carpetas (lista completa en analysis/lista_publicar.txt):"
cut -d/ -f1 analysis/lista_publicar.txt | sort | uniq -c
echo "--- Total de archivos: $(git diff --cached --name-only | wc -l)"
echo "--- Archivos de mas de 5 MB:"
git diff --cached --name-only -z | xargs -0 -I{} sh -c 'test -f "{}" && s=$(stat -c %s "{}") && [ "$s" -gt 5000000 ] && echo "  {} $s" || true'
echo
echo "Revisa la lista. Si esta bien:  bash scripts/hs4_preparar_github.sh confirmar"
