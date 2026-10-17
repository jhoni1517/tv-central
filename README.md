# TV Central

Transforma um PC ou notebook com Windows, ligado na TV pelo HDMI, numa **TV Box**:
tela inicial bonita em tela cheia, streamings com **navegação por setas** e **controle pelo celular**.

## O que faz

- **Tela inicial de TV**: abre junto com o Windows, em tela cheia, com Netflix, HBO Max, Disney+, Apple TV, YouTube, Prime Video e Globoplay.
- **Navegação por setas nos streamings**: nos sites feitos para mouse (Netflix, Disney+, HBO Max, Apple TV…), as setas pulam entre capas e botões, OK abre e Voltar volta. O YouTube abre na interface de TV dele.
- **Controle pelo celular** (app instalável): direcional, OK, voltar, início, volume, play/pausa, touchpad, lista de apps.
- **Digitar e falar**: o texto vai para a TV com Enter e o campo **limpa sozinho**. Busca por voz em português.
- **Pareamento seguro**: o código aparece na TV e só celulares pareados controlam o PC.
- **QR code na tela**: aponte a câmera do celular e pronto.
- **Instalador sem PowerShell**: um `.exe` só, funciona até em Windows "otimizado" sem PowerShell.

## Instalar

1. Baixe o `TVCentral.exe` em [Releases](../../releases).
2. Dê dois cliques e aceite a permissão de administrador (é para liberar a porta 8765 no firewall, só para a rede de casa).
3. A tela inicial abre. No celular, na mesma Wi-Fi, escaneie o QR code e toque em **Mostrar código na TV**.

Requisitos: Windows 10 ou 11 (64 bits) e Microsoft Edge (ou Google Chrome).

> Dica: reserve o IP do PC no roteador ("Reserva de endereço" / "DHCP estático") para o endereço do controle não mudar.

### Instalar o controle como app no Android

Endereços `http://` não podem ser instalados como app. Para instalar:
1. No Chrome do celular, abra `chrome://flags`, busque **insecure**, ative **Insecure origins treated as secure** e coloque o endereço do controle (ex.: `http://192.168.0.10:8765`).
2. Reinicie o Chrome, abra o endereço e toque em **⋮ → Instalar app**.

Isso também libera a busca por voz.

## Usar

| No celular | O que faz |
|---|---|
| Setas / OK | Anda e abre (na tela inicial e nos streamings) |
| Voltar | Volta uma tela |
| Início | Fecha o app e volta para a tela inicial |
| Campo de texto | Digita na TV e aperta Enter |
| 🎤 | Busca por voz |
| Touchpad | Mouse: arraste, toque para clicar, dois dedos para rolar |

Com vídeo tocando em tela grande, **OK pausa/continua** e as setas ficam com o player (avançar/voltar).

## Personalizar

Edite `%LOCALAPPDATA%\TVCentral\config.json` e reinicie o programa:

```json
{
  "porta": 8765,
  "abrirLauncher": true,
  "apps": [
    { "id": "prime", "oculto": true },
    { "id": "crunchyroll", "nome": "Crunchyroll", "destino": "https://www.crunchyroll.com", "cor": "#F47521" },
    { "id": "jogos", "nome": "Jogos", "tipo": "programa", "destino": "C:\\Jogos\\ES-DE\\ES-DE.exe", "cor": "#2E7D32" }
  ]
}
```

- Mesmo `id` de um app padrão: altera só os campos informados.
- `tipo`: `web` (abre no Edge em tela cheia) ou `programa` (um `.exe`).
- `setas: false` desliga a navegação por setas num site.
- `perfil`: perfil separado do navegador (logins ficam salvos por perfil).

## Remover

`TVCentral.exe remover` (pede administrador). Os dados ficam em `%LOCALAPPDATA%\TVCentral`.

## Desenvolver

```sh
go run .                     # roda fora do Windows também (teclas só aparecem no log)
GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui -s -w" -o dist/TVCentral.exe .
```

Estrutura:

| Arquivo | Função |
|---|---|
| `main.go` | Início, instalar/remover |
| `servidor.go` | HTTP, WebSocket do controle, rotas da TV |
| `gerente.go` | Abre/fecha apps e a tela inicial no Edge |
| `cdp.go` | Coloca a navegação por setas nas páginas (porta de depuração local) |
| `entrada_windows.go` | Teclado e mouse via `SendInput` |
| `sistema_windows.go` | Foco de janela, firewall, início automático, desligar |
| `web/tv` | Tela inicial |
| `web/celular` | Controle do celular (PWA) |
| `web/injetar/setas.js` | Navegação espacial por setas |

Uma versão nova é publicada em Releases automaticamente ao criar uma tag `v*`.

## Segurança

- A porta 8765 só é liberada para a **rede local**, e só para este programa.
- Sem pareamento, o celular não consegue mandar comandos. Os tokens ficam salvos só como hash.
- A tela inicial e as ações de sistema só respondem ao próprio computador.
- A porta de depuração do navegador (9222) escuta só em `127.0.0.1`.

## Licença

MIT
