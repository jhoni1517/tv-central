// TV Central - Receptor IR, passo 1: LEITOR
// Mostra o código de cada botão do controle no Monitor Serial (115200).
// Biblioteca: IRremote (Armin Joachimsmeyer), versão 4.x
#include <IRremote.hpp>

const int PINO_IR = 2;  // fio OUT do receptor

void setup() {
  Serial.begin(115200);
  while (!Serial) {}  // espera abrir o Monitor Serial
  IrReceiver.begin(PINO_IR, ENABLE_LED_FEEDBACK);
  Serial.println(F("Aponte o controle para o receptor e aperte um botao de cada vez."));
}

void loop() {
  if (!IrReceiver.decode()) return;
  IRData &d = IrReceiver.decodedIRData;
  if (!(d.flags & IRDATA_FLAGS_IS_REPEAT) && d.protocol != UNKNOWN) {
    Serial.print(F("Protocolo: "));
    Serial.print(getProtocolString(d.protocol));
    Serial.print(F("   Endereco: 0x"));
    Serial.print(d.address, HEX);
    Serial.print(F("   Comando: 0x"));
    Serial.println(d.command, HEX);
  }
  IrReceiver.resume();
}
