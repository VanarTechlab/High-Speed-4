# Fondo ORIGINAL de la ventana del instalador de High Speed 4 (640x300). VanarTechlab, MIT.
# Uso: python installer/fondo.py   -> installer/herramientas/fondo.bmp (+ fondo_vista.png)
import os
from PIL import Image, ImageDraw, ImageFilter, ImageFont
W, H = 640, 300
BG = (14, 10, 24)                       # mismo color que la ventana (gui_windows.go)
AQUI = os.path.dirname(os.path.abspath(__file__))
def font(names, size):
    for d in ("/usr/share/fonts/truetype/google-fonts", "C:/Windows/Fonts"):
        for n in names:
            p = os.path.join(d, n)
            if os.path.exists(p):
                return ImageFont.truetype(p, size)
    return ImageFont.load_default()
img = Image.new("RGB", (W, H))
px = img.load()
hy = 205
for y in range(H):
    for x in range(W):
        if y < hy:
            t = (y / hy) ** 1.5
            c0, c1 = (24, 12, 52), (255, 118, 70)
            px[x, y] = tuple(int(c0[i] + (c1[i] - c0[i]) * t) for i in range(3))
        else:
            px[x, y] = (18, 14, 30)
d = ImageDraw.Draw(img)
d.ellipse((W / 2 - 64, hy - 64, W / 2 + 64, hy + 64), fill=(255, 196, 92))
d.rectangle((0, hy, W, H), fill=(18, 14, 30))
for i in range(-14, 15):
    col = (255, 150, 60) if abs(i) in (13, 14) else (78, 62, 120)
    d.line((W / 2 + i * 9, hy, W / 2 + i * 110, H), fill=col, width=1)
for k in range(1, 9):
    y = hy + (H - hy) * (k / 9) ** 2.1
    d.line((0, y, W, y), fill=(60, 48, 96), width=1)
glow = Image.new("RGBA", (W, H), (0, 0, 0, 0))
g = ImageDraw.Draw(glow)
big = font(["Poppins-BoldItalic.ttf", "arialbi.ttf"], 78)
sub = font(["Poppins-MediumItalic.ttf", "ariali.ttf"], 20)
def centro(dr, t, f, y, fill):
    bb = dr.textbbox((0, 0), t, font=f)
    dr.text(((W - (bb[2] - bb[0])) / 2 - bb[0], y), t, font=f, fill=fill)
centro(g, "HIGH SPEED 4", big, 48, (255, 80, 40, 255))
glow = glow.filter(ImageFilter.GaussianBlur(9))
img.paste(glow, (0, 0), glow)
d = ImageDraw.Draw(img)
centro(d, "HIGH SPEED 4", big, 51, (40, 10, 30))
centro(d, "HIGH SPEED 4", big, 48, (255, 244, 225))
centro(d, "Port para PC  -  por Vanar (VanarTechlab)", sub, 150, (40, 12, 30))
centro(d, "Port para PC  -  por Vanar (VanarTechlab)", sub, 148, (255, 226, 200))
for y in range(H - 50, H):                     # fundido con el fondo de la ventana
    f = (y - (H - 50)) / 50
    for x in range(W):
        r, gg, b = img.getpixel((x, y))
        img.putpixel((x, y), (int(r * (1 - f) + BG[0] * f), int(gg * (1 - f) + BG[1] * f), int(b * (1 - f) + BG[2] * f)))
img.save(os.path.join(AQUI, "herramientas", "fondo.bmp"))
img.save(os.path.join(AQUI, "herramientas", "fondo_vista.png"))
print("fondo listo")
