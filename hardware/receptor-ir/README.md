# Receptor infravermelho USB (faça você mesmo)

Faz qualquer controle remoto comandar o TV Central: o Arduino recebe o infravermelho e vira um teclado USB.

## Peças

| Peça | Detalhe |
|---|---|
| **Arduino Pro Micro** (ATmega32U4, 5 V / 16 MHz) | Tem que ser 32U4 (Pro Micro ou Leonardo). Uno/Nano **não** viram teclado. |
| **Receptor IR 38 kHz** | VS1838B ou TSOP38238 (de preferência o módulo com 3 pinos e placa) |
| 3 fios jumper fêmea-fêmea | |
| Cabo micro-USB **de dados** | Cabo só de carga não funciona |

## Ligação

| Receptor | Pro Micro |
|---|---|
| OUT (ou S) | pino **2** |
| GND (ou −) | **GND** |
| VCC (ou +) | **VCC** |

Receptor solto (sem placa), olhando para a bolinha: pinos da esquerda para a direita = OUT, GND, VCC.

## Programar

1. Instale o **Arduino IDE 2**.
2. Em *Bibliotecas*, instale **IRremote** (Armin Joachimsmeyer) e **HID-Project** (NicoHood).
3. Ligue o Pro Micro no PC. Placa: **Arduino Leonardo** (ou "SparkFun Pro Micro" se instalar o pacote da SparkFun). Porta: a que aparecer.
4. Abra `leitor/leitor.ino`, envie, abra o **Monitor Serial** em 115200.
5. Aperte cada botão que vai usar e anote o **Endereço** e o **Comando**.
6. Abra `controle/controle.ino`, troque `ENDERECO` e os comandos da tabela `BOTOES`, e envie.
7. Pronto: o Pro Micro vira um teclado. Deixe ele num lugar com visão do controle (perto da TV).

## Qual controle usar

- **Melhor: um controle de outro aparelho** (DVD, home theater, TV Box antigo, controle universal). A TV ignora os códigos dele, então nada acontece duas vezes.
- **Controle da própria TV:** funciona, mas a TV também recebe os botões. Setas e OK costumam não fazer nada na TV quando ela está na entrada HDMI, mas **Home abre o menu da TV**. Use outro botão para "início" (por exemplo o azul) e **não** mapeie o volume (a TV já cuida dele).
- Controle de ar-condicionado não serve (manda um bloco de dados diferente a cada toque).

## Caixinha (impressão 3D)

`caixa/caixa_base.stl` e `caixa/caixa_tampa.stl` (49,6 × 23,2 × 14 mm montada). Feita para o Pro Micro **sem** barras de pinos e o receptor **solto** (3 perninhas) soldado com fios curtos, com a bolinha saindo pelo furo da frente. A tampa entra sob pressão; embaixo tem um rebaixo para fita dupla-face.

Medidas diferentes: ajuste o topo de `caixa/gerar_caixa.py` e rode `python3 gerar_caixa.py` (precisa de `pip install trimesh manifold3d`). Dá para mandar imprimir em serviços de impressão 3D sob encomenda.
