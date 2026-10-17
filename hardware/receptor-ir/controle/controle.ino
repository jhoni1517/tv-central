// TV Central - Receptor IR, passo 2: CONTROLE
// Transforma os botões do controle remoto em teclas do PC (o Arduino vira um teclado USB).
// Placa: Arduino Pro Micro / Leonardo (ATmega32U4). NÃO funciona com Uno/Nano.
// Bibliotecas: IRremote 4.x (Armin Joachimsmeyer) e HID-Project (NicoHood)
#include <IRremote.hpp>
#include <HID-Project.h>

const int PINO_IR = 2;  // fio OUT do receptor

// ===== PREENCHA COM O QUE O LEITOR MOSTROU =====
// Endereço do seu controle (aparece igual em todos os botões dele).
const uint16_t ENDERECO = 0x1;

enum Tipo : uint8_t { TECLA, MIDIA };
struct Botao {
  uint16_t comando;  // "Comando" que o leitor mostrou
  Tipo tipo;
  uint16_t tecla;
  bool repete;       // segurar o botão repete (setas e volume)
};

// Os valores abaixo são de exemplo (controle de TV Sony). Troque pelos do SEU controle.
const Botao BOTOES[] = {
  {0x74, TECLA, KEY_UP_ARROW, true},       // cima
  {0x75, TECLA, KEY_DOWN_ARROW, true},     // baixo
  {0x34, TECLA, KEY_LEFT_ARROW, true},     // esquerda
  {0x33, TECLA, KEY_RIGHT_ARROW, true},    // direita
  {0x65, TECLA, KEY_ENTER, false},         // OK
  {0x63, TECLA, KEY_ESC, false},           // voltar (Return)
  {0x60, MIDIA, CONSUMER_BROWSER_HOME, false}, // início -> tela inicial do TV Central
  {0x1A, MIDIA, MEDIA_PLAY_PAUSE, false},  // play/pausa
  // Volume: só use se o controle NÃO for o da própria TV (senão o volume muda duas vezes)
  // {0x12, MIDIA, MEDIA_VOLUME_UP, true},
  // {0x13, MIDIA, MEDIA_VOLUME_DOWN, true},
  // {0x14, MIDIA, MEDIA_VOLUME_MUTE, false},
};
// ================================================

const unsigned long PRIMEIRA_REPETICAO = 380;  // ms segurando até começar a repetir
const unsigned long INTERVALO_REPETICAO = 110; // ms entre repetições
const unsigned long FIM_DO_TOQUE = 160;        // ms sem sinal = botão solto

int atual = -1;
unsigned long inicioToque = 0, ultimoSinal = 0, ultimaRepeticao = 0;

void apertar(const Botao &b) {
  if (b.tipo == TECLA) Keyboard.write(KeyboardKeycode(b.tecla));
  else Consumer.write(ConsumerKeycode(b.tecla));
}

void setup() {
  IrReceiver.begin(PINO_IR, DISABLE_LED_FEEDBACK);
  Keyboard.begin();
  Consumer.begin();
}

void loop() {
  unsigned long agora = millis();
  if (atual >= 0 && agora - ultimoSinal > FIM_DO_TOQUE) atual = -1;  // soltou

  if (!IrReceiver.decode()) return;
  IRData d = IrReceiver.decodedIRData;
  IrReceiver.resume();
  if (d.protocol == UNKNOWN || d.address != ENDERECO) return;

  for (int i = 0; i < (int)(sizeof(BOTOES) / sizeof(BOTOES[0])); i++) {
    if (BOTOES[i].comando != d.command) continue;
    ultimoSinal = agora;
    if (atual != i) {  // toque novo
      atual = i;
      inicioToque = ultimaRepeticao = agora;
      apertar(BOTOES[i]);
    } else if (BOTOES[i].repete && agora - inicioToque > PRIMEIRA_REPETICAO &&
               agora - ultimaRepeticao > INTERVALO_REPETICAO) {
      ultimaRepeticao = agora;  // segurando o botão
      apertar(BOTOES[i]);
    }
    return;
  }
}
