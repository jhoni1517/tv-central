"""Gera as artes de fundo (originais) de cada app da tela inicial.
Cada arte é abstrata, no estilo "cena de cinema", com as cores do app. Sem logos nem imagens de terceiros.
Rode: python3 recursos/gerar_artes.py
"""
import random, math
from PIL import Image, ImageDraw, ImageFilter, ImageChops

W, H = 1600, 900
APPS = {  # id: (cor principal, cor secundária, estilo)
    "netflix": ("#E50914", "#3d0205", "feixes"),
    "hbomax": ("#7B2FF7", "#1b0a3d", "ondas"),
    "disney": ("#0063E5", "#0b1a4a", "estrelas"),
    "appletv": ("#9a9aa8", "#111114", "aurora"),
    "youtube": ("#FF0033", "#7a0019", "ondas"),
    "prime": ("#1A98FF", "#0a2a4a", "feixes"),
    "globoplay": ("#F25C2B", "#4a1406", "aurora"),
    "youtubekids": ("#FF4E45", "#5b1a10", "estrelas"),
    "padrao": ("#3a6df0", "#0b1430", "estrelas"),
}

def rgb(h): h = h.lstrip('#'); return tuple(int(h[i:i+2], 16) for i in (0, 2, 4))

def brilho(img, cor, cx, cy, r, forca):
    m = Image.new('L', (W, H), 0)
    ImageDraw.Draw(m).ellipse([cx - r, cy - r, cx + r, cy + r], fill=int(255 * forca))
    m = m.filter(ImageFilter.GaussianBlur(r * 0.55))
    return Image.composite(Image.new('RGB', (W, H), cor), img, m)

def gerar(nome, c1, c2, estilo):
    rnd = random.Random(nome)
    a, b = rgb(c1), rgb(c2)
    img = Image.new('RGB', (W, H), (6, 7, 11))
    img = brilho(img, b, int(W * .75), int(H * .85), 700, .9)
    img = brilho(img, a, int(W * (.7 + rnd.uniform(-.06, .06))), int(H * .28), 420, .75)
    img = brilho(img, tuple(min(255, int(v * 1.4 + 40)) for v in a), int(W * .78), int(H * .3), 140, .55)

    luz = Image.new('L', (W, H), 0); d = ImageDraw.Draw(luz)
    if estilo == "feixes":  # feixes de luz de projetor
        for i in range(7):
            x = W * .78 + rnd.uniform(-60, 60); ang = rnd.uniform(-1.3, -.25); larg = rnd.uniform(.05, .13)
            p1 = (x + 2000 * math.cos(ang - larg), H * .3 + 2000 * math.sin(ang - larg) * -1)
            p2 = (x + 2000 * math.cos(ang + larg), H * .3 + 2000 * math.sin(ang + larg) * -1)
            d.polygon([(x, H * .3), (W - p1[0] + W * .55, p1[1] + H), (W - p2[0] + W * .55, p2[1] + H)], fill=rnd.randint(25, 60))
        luz = luz.filter(ImageFilter.GaussianBlur(28))
    elif estilo == "ondas":  # faixas de seda
        for i in range(9):
            y0 = H * rnd.uniform(.25, .95); amp = rnd.uniform(40, 120); fase = rnd.uniform(0, 6); esp = rnd.uniform(6, 26)
            pts = [(x, y0 + amp * math.sin(x / 260 + fase)) for x in range(0, W + 20, 20)]
            d.line(pts, fill=rnd.randint(60, 140), width=int(esp))
        luz = luz.filter(ImageFilter.GaussianBlur(10))
    elif estilo == "estrelas":  # céu com brilho
        for i in range(260):
            x, y = rnd.uniform(W * .3, W), rnd.uniform(0, H * .8); r = rnd.choice([1, 1, 1, 2, 2, 3])
            d.ellipse([x - r, y - r, x + r, y + r], fill=rnd.randint(90, 255))
        arco = Image.new('L', (W, H), 0)
        ImageDraw.Draw(arco).arc([W * .35, H * .35, W * 1.45, H * 1.9], 190, 300, fill=200, width=6)
        luz = ImageChops.add(luz.filter(ImageFilter.GaussianBlur(.6)), arco.filter(ImageFilter.GaussianBlur(5)))
    else:  # aurora
        for i in range(5):
            y0 = H * rnd.uniform(.15, .55)
            pts = [(x, y0 + 90 * math.sin(x / 300 + i)) for x in range(0, W + 40, 40)]
            d.line(pts, fill=rnd.randint(70, 130), width=rnd.randint(50, 110))
        luz = luz.filter(ImageFilter.GaussianBlur(45))
    claro = tuple(min(255, int(v * .6 + 110)) for v in a)
    img = Image.composite(Image.new('RGB', (W, H), claro), img, luz)

    # bokeh (círculos desfocados)
    bok = Image.new('L', (W, H), 0); d = ImageDraw.Draw(bok)
    for i in range(22):
        x, y, r = rnd.uniform(W * .45, W), rnd.uniform(0, H), rnd.uniform(10, 46)
        d.ellipse([x - r, y - r, x + r, y + r], fill=rnd.randint(20, 70))
    img = Image.composite(Image.new('RGB', (W, H), claro), img, bok.filter(ImageFilter.GaussianBlur(6)))

    # escurece a esquerda (onde fica o texto) e as bordas
    esc = Image.new('L', (W, H), 0); px = esc.load()
    for x in range(W):
        v = int(235 * max(0, 1 - x / (W * .62)) ** 1.6)
        for y in range(H): px[x, y] = v
    img = Image.composite(Image.new('RGB', (W, H), (6, 7, 11)), img, esc)
    vin = Image.new('L', (W, H), 255)
    ImageDraw.Draw(vin).ellipse([-W * .2, -H * .3, W * 1.2, H * 1.3], fill=0)
    img = Image.composite(Image.new('RGB', (W, H), (4, 5, 8)), img, vin.filter(ImageFilter.GaussianBlur(160)))

    # granulado de filme
    ruido = Image.effect_noise((W, H), 18).convert('RGB')
    img = ImageChops.blend(img, ImageChops.overlay(img, ruido), .12)

    img.save(f"web/tv/arte/{nome}.jpg", quality=82, optimize=True, progressive=True)
    img.crop((int(W * .38), 0, W, int((W * .62) * 9 / 16))).resize((480, 270), Image.LANCZOS).save(f"web/tv/arte/{nome}-p.jpg", quality=80, optimize=True)

import sys
for n, (c1, c2, e) in APPS.items():
    if len(sys.argv) > 1 and n not in sys.argv[1:]:
        continue
    gerar(n, c1, c2, e); print(n)
