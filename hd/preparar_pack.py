# Pack final mixto desde hd\salida (x4): x4 si la textura tiene mucho detalle (>= umbral), x2 si no;
# las de malas.txt (la IA invento un dibujo distinto) se quedan fuera y el juego usa la original.
# Reanudable: preparar_pack.py INICIO SEGUNDOS  -> hd\pack_final
import os, sys, time, shutil
from PIL import Image, ImageChops, ImageStat
from concurrent.futures import ProcessPoolExecutor
base = os.path.dirname(os.path.abspath(__file__))
src = os.path.join(base, os.environ.get("HD_SALIDA", "salida")); dst = os.path.join(base, os.environ.get("HD_PACK", "pack_final"))
UMBRAL = 3.0
def one(f):
    im = Image.open(os.path.join(src, f)).convert("RGBA")
    w, h = im.size
    small = im.convert("RGBa").resize((max(1, w // 2), max(1, h // 2)), Image.LANCZOS).convert("RGBA")
    rgb = im.convert("RGB")
    back = small.convert("RGB").resize((w, h), Image.BICUBIC)
    d = ImageStat.Stat(ImageChops.difference(rgb, back).convert("L")).mean[0]
    if d >= UMBRAL:
        shutil.copyfile(os.path.join(src, f), os.path.join(dst, f)); return "x4"
    small.putalpha(small.getchannel("A").point(lambda v: 255 if v >= 128 else 0))
    small.save(os.path.join(dst, f)); return "x2"
if __name__ == "__main__":
    os.makedirs(dst, exist_ok=True)
    malas = {l.split()[0] for l in open(os.path.join(base, "malas.txt"))}
    fs = sorted(f for f in os.listdir(src) if f.endswith(".png") and f not in malas and not f.startswith("texupload-ST"))
    if os.environ.get("HD_SOLO_NUEVAS"):
        nuevas = set(os.listdir(os.path.join(base, "entrada_nuevas")))
        fs = [f for f in fs if f in nuevas]
    ini = int(sys.argv[1]) if len(sys.argv) > 1 else 0
    lim = float(sys.argv[2]) if len(sys.argv) > 2 else 1e9
    t0 = time.time(); i = ini; cnt = {"x4": 0, "x2": 0}
    with ProcessPoolExecutor() as ex:
        while i < len(fs) and time.time() - t0 < lim:
            for r in ex.map(one, fs[i:i+128], chunksize=8): cnt[r] += 1
            i = min(i + 128, len(fs))
    print("siguiente", i, "de", len(fs), cnt)
