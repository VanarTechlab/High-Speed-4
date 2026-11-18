# Portada ORIGINAL de High Speed 4 (sin arte de EA). Genera launcher_assets/img/boxart.png y boxart.tga (512x512).
# Uso: python tools/hs4_portada.py [carpeta_fuentes]   (necesita Pillow y una fuente TTF, p. ej. Poppins Bold Italic)
import math, os, struct, sys
from PIL import Image, ImageDraw, ImageFilter, ImageFont
W = H = 512
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
OUT = os.path.join(ROOT, "launcher_assets", "img")
FONT_DIRS = [sys.argv[1]] if len(sys.argv) > 1 else []
FONT_DIRS += ["/usr/share/fonts/truetype/google-fonts", "C:/Windows/Fonts"]
def font(names, size):
    for d in FONT_DIRS:
        for n in names:
            p = os.path.join(d, n)
            if os.path.exists(p):
                return ImageFont.truetype(p, size)
    return ImageFont.load_default()
img = Image.new("RGB", (W, H))
px = img.load()
for y in range(H):                                   # cielo de atardecer
    t = y / (H * 0.62)
    if t <= 1:
        c0, c1 = (18, 10, 48), (255, 120, 70)
        r = [int(c0[i] + (c1[i] - c0[i]) * (t ** 1.6)) for i in range(3)]
    else:
        r = [14, 12, 22]
    for x in range(W):
        px[x, y] = tuple(r)
d = ImageDraw.Draw(img)
hy = int(H * 0.62)
d.ellipse((W/2-70, hy-70, W/2+70, hy+70), fill=(255, 196, 92))     # sol
d.rectangle((0, hy, W, H), fill=(16, 14, 26))
for i in range(-9, 10):                              # carretera en perspectiva
    x0 = W/2 + i * 8
    x1 = W/2 + i * 120
    col = (255, 150, 60) if abs(i) in (8, 9) else (70, 60, 110)
    d.line((x0, hy, x1, H), fill=col, width=2)
for k in range(1, 12):
    y = hy + (H - hy) * (k / 12) ** 2.2
    d.line((0, y, W, y), fill=(55, 45, 90), width=1)
for k in range(0, 8):                                # lineas centrales
    y0 = hy + (H - hy) * (k / 10) ** 2
    y1 = hy + (H - hy) * ((k + 0.5) / 10) ** 2
    w0 = 1 + 10 * (k / 10) ** 2
    d.polygon([(W/2-w0, y0), (W/2+w0, y0), (W/2+w0*1.3, y1), (W/2-w0*1.3, y1)], fill=(255, 220, 120))
glow = Image.new("RGBA", (W, H), (0, 0, 0, 0))
g = ImageDraw.Draw(glow)
f_big = font(["Poppins-BoldItalic.ttf", "arialbi.ttf"], 92)
f_num = font(["Poppins-BoldItalic.ttf", "arialbi.ttf"], 230)
f_small = font(["Poppins-MediumItalic.ttf", "Poppins-Italic.ttf", "ariali.ttf"], 22)
def centered(dr, text, f, y, fill):
    bb = dr.textbbox((0, 0), text, font=f)
    dr.text(((W - (bb[2] - bb[0])) / 2 - bb[0], y), text, font=f, fill=fill)
centered(g, "4", f_num, 170, (255, 80, 40, 255))
glow = glow.filter(ImageFilter.GaussianBlur(10))
img.paste(glow, (0, 0), glow)
d = ImageDraw.Draw(img)
centered(d, "4", f_num, 170, (255, 236, 200))
for y, txt in ((40, "HIGH"), (128, "SPEED")):
    centered(d, txt, f_big, y + 3, (40, 10, 30))
    centered(d, txt, f_big, y, (255, 255, 255))
centered(d, "VanarTechlab", f_small, H - 38, (0, 0, 0)); centered(d, "VanarTechlab", f_small, H - 40, (220, 210, 245))
img.save(os.path.join(OUT, "boxart.png"))
rgba = img.convert("RGBA")                            # TGA 32 bpp sin comprimir, origen abajo-izquierda (como el original)
with open(os.path.join(OUT, "boxart.tga"), "wb") as f:
    f.write(struct.pack("<BBBHHBHHHHBB", 0, 0, 2, 0, 0, 0, 0, 0, W, H, 32, 8))
    for y in range(H - 1, -1, -1):
        row = bytearray()
        for x in range(W):
            r, gg, b, a = rgba.getpixel((x, y))
            row += bytes((b, gg, r, a))
        f.write(row)
print("portada escrita en", OUT)
