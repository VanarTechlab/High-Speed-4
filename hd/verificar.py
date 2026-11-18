# Verifica hd\salida contra dumps: tamano x4 y que el color se parezca al original. Escribe malas.txt
import os, sys
from PIL import Image, ImageChops, ImageStat
from concurrent.futures import ProcessPoolExecutor
base = os.path.dirname(os.path.abspath(__file__))
dumps = os.path.join(base, "..", "build-windows", "mods", "texture-packs", "SLUS-00826", "dumps")
def check(f):
    try:
        a = Image.open(os.path.join(dumps, f)).convert("RGBA")
        b = Image.open(os.path.join(base, os.environ.get("HD_SALIDA", "salida"), f)).convert("RGBA")
        if b.size != (a.size[0]*4, a.size[1]*4): return (f, "tamano %s" % (b.size,))
        s = b.resize(a.size, Image.BOX)
        ra, rb = Image.new("RGB", a.size), Image.new("RGB", a.size)
        ra.paste(a.convert("RGB"), mask=a.getchannel("A")); rb.paste(s.convert("RGB"), mask=a.getchannel("A"))
        d = ImageStat.Stat(ImageChops.difference(ra, rb).convert("L")).mean[0]
        return (f, "color %.1f" % d) if d > 20 else None
    except Exception as e:
        return (f, "error %s" % e)
if __name__ == "__main__":
    import time
    fs = sorted(x for x in os.listdir(dumps) if x.endswith(".png"))
    if os.environ.get("HD_SOLO_NUEVAS"):
        nuevas = set(os.listdir(os.path.join(base, "entrada_nuevas")))
        fs = [f for f in fs if f in nuevas]
    ini = int(sys.argv[1]) if len(sys.argv) > 1 else 0
    lim = float(sys.argv[2]) if len(sys.argv) > 2 else 1e9   # segundos maximos
    t0 = time.time(); i = ini; malas = open(os.path.join(base, "malas.txt"), "a" if (ini or os.environ.get("HD_SOLO_NUEVAS")) else "w")
    with ProcessPoolExecutor() as ex:
        while i < len(fs) and time.time() - t0 < lim:
            for r in ex.map(check, fs[i:i+256], chunksize=16):
                if r: malas.write("%s %s\n" % r)
            i = min(i + 256, len(fs)); malas.flush()
    print("siguiente", i, "de", len(fs))
