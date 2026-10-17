"""Gera a caixinha do receptor IR (base + tampa) em STL para impressão 3D.

Medidas em mm. Feita para: Arduino Pro Micro (sem barras de pinos soldadas) + receptor
VS1838B solto (3 perninhas) com a bolinha saindo pelo furo da frente.
Rode: python3 gerar_caixa.py   (precisa de: pip install trimesh manifold3d)
"""
from manifold3d import CrossSection, Manifold, JoinType
import trimesh

# ---- Ajuste aqui se a sua placa for diferente ----
PLACA_C, PLACA_L = 34.0, 18.6     # Pro Micro: comprimento x largura
ESPACO_RECEPTOR = 11.0            # espaço na frente para o receptor
ALTURA_INTERNA = 9.0
PAREDE = 1.8
FUNDO = 1.6
RAIO = 3.5                        # cantos arredondados
FURO_RECEPTOR = 6.4               # diâmetro do furo da bolinha do receptor
USB_L, USB_A = 11.0, 6.5          # abertura do conector USB (serve micro-USB e USB-C)
FOLGA = 0.25                      # folga do encaixe da tampa
# ---------------------------------------------------

IC = PLACA_C + ESPACO_RECEPTOR + 1.0   # comprimento interno
IL = PLACA_L + 1.0                     # largura interna
EC, EL = IC + 2 * PAREDE, IL + 2 * PAREDE
EA = FUNDO + ALTURA_INTERNA


def ret_redondo(c, l, r):
    r = min(r, c / 2 - 0.01, l / 2 - 0.01)
    base = CrossSection.square((c - 2 * r, l - 2 * r), center=True)
    return base.offset(r, JoinType.Round, circular_segments=48)


def caixa(c, l, a, r):
    return Manifold.extrude(ret_redondo(c, l, r), a)


# Base: caixa aberta em cima
base = caixa(EC, EL, EA, RAIO) - caixa(IC, IL, EA, max(RAIO - PAREDE, 0.6)).translate((0, 0, FUNDO))
# Furo do receptor (frente, eixo X negativo)
furo = Manifold.cylinder(PAREDE * 3, FURO_RECEPTOR / 2, circular_segments=48).rotate((0, 90, 0))
base -= furo.translate((-EC / 2 - PAREDE, 0, FUNDO + 4.2))
# Abertura do USB (fundo, eixo X positivo)
usb = Manifold.cube((PAREDE * 3, USB_L, USB_A), center=True)
base -= usb.translate((EC / 2, 0, FUNDO + 0.6 + USB_A / 2))
# Trilhos que seguram a placa (2 batentes laterais perto do USB)
for y in (-IL / 2 + 0.6, IL / 2 - 0.6):
    base += Manifold.cube((PLACA_C * 0.6, 1.2, 1.2), center=True).translate((EC / 2 - PAREDE - PLACA_C * 0.3 - 1, y, FUNDO + 0.6))
# Pé antiderrapante: rebaixo de 0,4 mm para colar fita dupla-face
base -= caixa(EC - 8, EL - 8, 0.4, 2)

# Tampa: placa + aba que entra na caixa
tampa = caixa(EC, EL, 1.6, RAIO)
aba = caixa(IC - 2 * FOLGA, IL - 2 * FOLGA, 2.2, max(RAIO - PAREDE, 0.6)) - caixa(IC - 2 * FOLGA - 2.4, IL - 2 * FOLGA - 2.4, 2.2, 0.6)
tampa += aba.translate((0, 0, 1.6))
# Detalhe decorativo: friso com uma "tela" em relevo baixo
tampa -= caixa(EC * 0.45, EL * 0.42, 0.5, 1.5).translate((-EC * 0.12, 0, 0))

for nome, m in (("caixa_base", base), ("caixa_tampa", tampa)):
    mesh = m.to_mesh()
    t = trimesh.Trimesh(vertices=mesh.vert_properties[:, :3], faces=mesh.tri_verts)
    t.export(f"{nome}.stl")
    print(nome, "ok", "medidas externas: %.1f x %.1f x %.1f mm" % tuple(t.extents), "fechado:", t.is_watertight)
