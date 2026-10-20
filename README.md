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
- **Visual de TV Box**: arte de fundo para cada app, cartões com imagem, abertura com música, sons de navegação e protetor de tela.
- **Desligar em…**: 30, 60, 90 ou 120 minutos, pela TV ou pelo celular.
- **Painel de ajustes no celular** (⚙): ordem e visibilidade dos apps, enviar imagem e vídeo de cada app, clima, sons, abertura, protetor.
- **Perfis**: "Quem está assistindo?", cada um com seus apps e **logins separados** nos streamings.
- **Modo criança**: só os apps escolhidos, sem desligar/ajustes, e para sair só com PIN no celular.
- **Bloqueio de anúncios** nas janelas do TV Central (sites e YouTube), sem mexer no seu navegador normal.
- **Adicione qualquer site** pelo celular (Crunchyroll, Twitch, Pluto TV…).
- **Assistir depois**: compartilhe um link no celular → vira cartão na TV e abre no app certo.
- **Fotos na TV**: mande fotos pelo celular; viram protetor de tela e dá para mostrar na hora.
- **Espelhar o celular** pelo "Conectar" do Windows.
- **Modo desempenho** automático para PC fraco e linha opcional de CPU/memória/disco.
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
| Gestos | Deslize = setas, toque = OK, segure = Voltar |
| Touchpad | Mouse: arraste, toque para clicar, dois dedos para rolar |

Com vídeo tocando em tela grande, **OK pausa/continua** e as setas ficam com o player (avançar/voltar).

## Controle remoto da TV

O controle da TV fala com a TV por infravermelho; o PC não recebe esse sinal sozinho. Com um **receptor infravermelho USB que aprende botões** (ex.: FLIRC), ensine:

| Botão do controle | Tecla |
|---|---|
| Setas / OK | Setas / Enter |
| Return | Esc |
| Home | Home do navegador (tecla de mídia) |
| Play/Pause | Play/Pause (tecla de mídia) |
| Vol + / − / Mudo | teclas de volume |

A tecla **Home do navegador** e **Ctrl+Alt+H** voltam para a tela inicial de qualquer lugar.

Para mexer no Windows: **Sistema → Área de trabalho** fecha a tela inicial até você apertar Início de novo. Se a tela inicial fechar por acidente, ela reabre sozinha.

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
- `"intro": false` desliga a abertura; `"volumeIntro"` vai de 0 a 1 (padrão 0.5).
- `"cidade": "Curitiba"` mostra o clima no topo da tela inicial.
- `"sons": false` desliga os sons de navegação; `"protetorMin"`: minutos até o protetor de tela (0 desliga).

## Atualizações

O programa procura versão nova sozinho (90 s depois de abrir e a cada 6 horas), confere o SHA-256 e instala quando **nenhum app está aberto**. A tela inicial e o controle do celular recarregam sozinhos. Também dá para forçar em **Sistema → Atualizar** na tela inicial. Se a versão nova não abrir, ele volta para a anterior.

Para desligar: `"atualizarAuto": false` no `config.json`.

Para publicar uma versão: mude o número no arquivo `VERSION` e envie para a `main`. O GitHub Actions compila e publica em Releases.

## Problemas

No celular: **Apps → Diagnóstico** mostra a versão, a rede e as últimas linhas do registro do PC. O registro completo fica em `%LOCALAPPDATA%\TVCentral\logs\tvcentral.log`.

O medidor ao lado do nome no controle mostra o tempo de resposta do PC: verde até 80 ms. Se ficar vermelho, o problema é o Wi-Fi.

## Imagens e vídeos dos apps

As artes que vêm no programa são próprias (sem logos). Para usar a sua imagem num app, coloque em `%LOCALAPPDATA%\TVCentral\capas\` um arquivo com o id do app: `netflix.jpg`, `hbomax.jpg`, `disney.jpg`, `appletv.jpg`, `youtube.jpg`, `prime.jpg`, `globoplay.jpg` (ou `.png`/`.webp`, de preferência 1600×900 e mais escura à esquerda).

**Vídeos de fundo:** coloque `netflix.mp4`, `disney.webm` etc. na mesma pasta. O vídeo toca sem som no fundo quando o app está selecionado e no protetor de tela; `fundo.mp4` toca quando nenhum app está selecionado. Prefira vídeos curtos (10 a 60 s), 720p ou 1080p, em H.264 (`.mp4`). O vídeo para sozinho enquanto um app está aberto. Para desligar: `"videosFundo": false`.

## Remover

`TVCentral.exe remover` (pede administrador). Os dados ficam em `%LOCALAPPDATA%\TVCentral`.

## Desenvolver

```sh
go run .                     # roda fora do Windows também (teclas só aparecem no log)
go test ./...
GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui -s -w -X main.versao=$(cat VERSION)" -o dist/TVCentral.exe .
```

Estrutura:

| Arquivo | Função |
|---|---|
| `main.go` | Início, instalar/remover |
| `servidor.go` | HTTP, WebSocket do controle, rotas da TV |
| `gerente.go` | Abre/fecha apps e a tela inicial no Edge |
| `atualizar.go` | Atualização automática pelo GitHub Releases |
| `cdp.go` | Coloca a navegação por setas nas páginas (porta de depuração local) |
| `entrada_windows.go` | Teclado e mouse via `SendInput` |
| `sistema_windows.go` | Foco de janela, firewall, início automático, desligar |
| `web/tv` | Tela inicial |
| `web/celular` | Controle do celular (PWA) |
| `web/injetar/setas.js` | Navegação espacial por setas |


## Segurança

- A porta 8765 só é liberada para a **rede local**, e só para este programa.
- Sem pareamento, o celular não consegue mandar comandos. Os tokens ficam salvos só como hash.
- A tela inicial e as ações de sistema só respondem ao próprio computador.
- A porta de depuração do navegador (9222) escuta só em `127.0.0.1`.

## Licença

MIT
