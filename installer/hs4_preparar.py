# High Speed 4 - preparar el juego en el PC del jugador (VanarTechlab, MIT).
# Traduce y compila el codigo del juego DESDE EL DISCO DEL JUGADOR con el perfil AOT de PSXRecomp
# (hs4_aot/overlays.json: solo archivos, direcciones y huellas; ningun byte del juego viene en la descarga).
# Uso (lo lanza "Preparar juego.bat" o el instalador):
#   overlay_toolchain\python\python.exe hs4_aot\hs4_preparar.py [--disc RUTA.cue] [--clang RUTA\clang.exe] [--workers N]
# Sin --disc usa [disc] path de settings.toml. Sin --clang usa overlay_toolchain\compiler.txt.
import argparse, os, runpy, shutil, sys, time
from pathlib import Path

AQUI = Path(__file__).resolve().parent            # <juego>/hs4_aot
APP = AQUI.parent                                  # carpeta del juego
TK = APP / 'overlay_toolchain'
FW = AQUI / 'psxrecomp'                            # mini "framework": tools (vienen en la descarga) + include/bios/seeds (se copian)

def leer_disco_settings():
    f = APP / 'settings.toml'
    if not f.is_file():
        return None
    import tomllib
    try:
        return tomllib.loads(f.read_text(encoding='utf-8-sig')).get('disc', {}).get('path')
    except Exception:
        return None

def copiar_arbol(src, dst):
    if dst.exists():
        shutil.rmtree(dst)
    shutil.copytree(src, dst)

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--disc')
    ap.add_argument('--clang')
    ap.add_argument('--workers', type=int, default=max(1, (os.cpu_count() or 2) - 1))
    a = ap.parse_args()
    disco = a.disc or leer_disco_settings()
    if not disco or not Path(disco).is_file():
        print('ERROR: no encuentro la imagen del disco (.cue). Eligela en el lanzador o pasala con --disc.')
        return 2
    clang = a.clang
    if not clang and (TK / 'compiler.txt').is_file():
        clang = (TK / 'compiler.txt').read_text(encoding='utf-8').strip()
    if not clang or not Path(clang).is_file():
        print('ERROR: falta el compilador clang. Reinstala High Speed 4 marcando "Descargar el compilador clang".')
        return 3
    print('High Speed 4: preparando el juego desde tu disco.')
    print('  Disco :', disco)
    print('  clang :', clang)
    print('  Esto tarda unos minutos la primera vez (3 en el PC de Vanar). No cierres esta ventana.')
    # Piezas del framework que ya trae el juego (mismas que usa la compilacion en segundo plano)
    copiar_arbol(TK / 'include', FW / 'runtime' / 'include')
    copiar_arbol(TK / 'bios', FW / 'bios')
    if (TK / 'recompiler').is_dir():
        copiar_arbol(TK / 'recompiler', FW / 'recompiler')
    shutil.copy2(APP / 'game.toml', AQUI / 'game.toml')
    trabajo, escenario = AQUI / 'trabajo', AQUI / 'escenario'
    for d in (trabajo, escenario):
        if d.exists():
            shutil.rmtree(d)
    sys.path[:0] = [str(FW / 'tools'), str(FW / 'tools' / 'aot_overlay_spike')]
    sys.argv = ['aot_overlay_pipeline.py', 'release',
                '--profile', str(AQUI / 'overlays.json'), '--game-toml', str(AQUI / 'game.toml'),
                '--disc', str(Path(disco).resolve()), '--recompiler', str(TK / 'psxrecomp-game.exe'),
                '--work-dir', str(trabajo), '--stage', str(escenario), '--gcc', clang,
                '--workers', str(a.workers), '--cps', '--runtime-config', str(APP / 'game.toml')]
    t0 = time.time()
    try:
        runpy.run_path(str(FW / 'tools' / 'aot_overlay_pipeline.py'), run_name='__main__')
        codigo = 0
    except SystemExit as e:
        codigo = e.code if isinstance(e.code, int) else (0 if e.code is None else 1)
    if codigo != 0:
        print('\nERROR preparando el juego (codigo %s). Registros en %s' % (codigo, trabajo))
        if codigo == 3:
            print('Tu disco no es el original USA (SLUS-00826) que espera High Speed 4.')
        print('El juego funcionara igual, compilando poco a poco mientras juegas.')
        return 1
    origen = escenario / 'cache'
    destino = APP / 'cache'
    for f in origen.rglob('*'):
        if f.is_file():
            d = destino / f.relative_to(origen)
            d.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(f, d)
    shutil.copy2(escenario / 'AOT_CACHE_AUDIT.json', APP / 'AOT_CACHE_AUDIT.json')
    shutil.rmtree(trabajo, ignore_errors=True)
    shutil.rmtree(escenario, ignore_errors=True)
    print('\nListo en %.0f minutos. Ya puedes jugar a High Speed 4.' % ((time.time() - t0) / 60))
    return 0

if __name__ == '__main__':
    sys.exit(main())
