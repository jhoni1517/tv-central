// TV Central: transforma um PC com Windows numa central de TV,
// com launcher em tela cheia e controle pelo celular.
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var versao = "dev" // preenchida na compilação: -ldflags "-X main.versao=0.2.0"

func main() {
	cmd := ""
	if len(os.Args) > 1 {
		cmd = strings.ToLower(strings.TrimLeft(os.Args[1], "-/"))
	}
	switch cmd {
	case "instalar":
		sair(instalar())
	case "remover", "desinstalar":
		sair(remover())
	case "servidor", "rodar":
		rodar(lerOpcoes(os.Args[2:]))
	case "versao", "version":
		fmt.Println("TV Central", versao)
	default:
		// Duplo clique no .exe baixado: instala. O .exe instalado: roda.
		if precisaInstalar() {
			sair(instalar())
			return
		}
		rodar(Opcoes{})
	}
}

func sair(err error) {
	if err != nil {
		avisar("TV Central", "Erro: "+err.Error())
		os.Exit(1)
	}
}

// pastaDados guarda configuração, celulares pareados, perfis do navegador e logs.
func pastaDados() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		h, _ := os.UserHomeDir()
		return filepath.Join(h, ".tvcentral")
	}
	return filepath.Join(base, "TVCentral")
}

// Opcoes vêm da versão anterior quando ela se atualiza e abre esta.
type Opcoes struct {
	LauncherPID  int
	AtualizadoDe string
}

func lerOpcoes(args []string) Opcoes {
	var o Opcoes
	for _, a := range args {
		if v, ok := strings.CutPrefix(a, "--launcher="); ok {
			o.LauncherPID, _ = strconv.Atoi(v)
		}
		if v, ok := strings.CutPrefix(a, "--atualizado-de="); ok {
			o.AtualizadoDe = v
		}
	}
	return o
}

func rodar(op Opcoes) {
	dados := pastaDados()
	if err := os.MkdirAll(filepath.Join(dados, "logs"), 0o755); err != nil {
		avisar("TV Central", err.Error())
		return
	}
	arqLog := filepath.Join(dados, "logs", "tvcentral.log")
	if fi, err := os.Stat(arqLog); err == nil && fi.Size() > 5<<20 {
		_ = os.Rename(arqLog, arqLog+".1")
	}
	f, err := os.OpenFile(arqLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err == nil {
		log.SetOutput(io.MultiWriter(f, os.Stderr))
	}
	log.Printf("TV Central %s iniciando (dados em %s)", versao, dados)

	prepararCapas(dados)
	cfg, err := carregarConfig(dados)
	if err != nil {
		log.Printf("config: %v (usando padrão)", err)
	}
	par, err := novoPareamento(filepath.Join(dados, "celulares.json"))
	if err != nil {
		log.Printf("pareamento: %v", err)
	}
	ger := novoGerente(dados, cfg.Porta)
	aj, err := carregarAjustes(filepath.Join(dados, "ajustes.json"))
	if err != nil {
		log.Printf("ajustes: %v (usando padrão)", err)
	}
	srv := novoServidor(cfg, par, ger, aj)
	ger.bloquearAnuncios = srv.bloquearAnuncios
	if op.AtualizadoDe != "" {
		log.Printf("atualizado de %s para %s", op.AtualizadoDe, versao)
		srv.novidade = fmt.Sprintf("Atualizado para a versão %s", versao)
		srv.novidadeAte = time.Now().Add(2 * time.Minute)
	}

	// Já está rodando? (ex.: clicou no atalho de novo) Só traz a tela inicial para frente.
	if op.AtualizadoDe == "" && jaRodando(cfg.Porta) {
		log.Printf("já está aberto: só mostrando a tela inicial")
		return
	}
	ouvinte, err := escutar(cfg.Porta)
	if err != nil {
		log.Printf("porta %d: %v", cfg.Porta, err)
		if op.AtualizadoDe != "" && voltarVersaoAnterior() {
			log.Printf("versão nova não abriu: voltei para a anterior")
			os.Exit(1)
		}
		avisar("TV Central", fmt.Sprintf("Não consegui usar a porta %d. Outro programa (por exemplo o controle antigo) está usando ela.\n\n%v", cfg.Porta, err))
		os.Exit(1)
	}
	// Versão nova funcionando: a cópia antiga pode ir embora
	go func() {
		time.Sleep(2 * time.Minute)
		if exe, err := os.Executable(); err == nil {
			_ = os.Remove(exe + ".antigo")
		}
	}()

	// Reaproveita a tela inicial e o app que já estavam abertos (atualização ou reinício do programa)
	ger.AdotarApp(srv.todosApps())
	if !ger.AdotarLauncher(op.LauncherPID) && cfg.abrirLauncher() {
		go ger.AbrirLauncher()
	}
	atalhosGlobais(ger.Inicio)
	if cfg.abrirLauncher() {
		go ger.Vigia()
	}
	go srv.clima.Loop()
	if cfg.atualizarAuto() {
		go srv.atu.Loop(dados, ger, func() bool { return srv.timer.Restante() == 0 })
	}
	if err := srv.Rodar(ouvinte); err != nil {
		log.Printf("servidor parou: %v", err)
		os.Exit(1)
	}
}

// prepararCapas cria a pasta onde a pessoa pode pôr imagens próprias para os apps.
func prepararCapas(dados string) {
	pasta := filepath.Join(dados, "capas")
	_ = os.MkdirAll(pasta, 0o755)
	if sobras, _ := filepath.Glob(filepath.Join(pasta, ".envio-*")); len(sobras) > 0 {
		for _, f := range sobras { // envios interrompidos
			_ = os.Remove(f)
		}
	}
	leia := filepath.Join(pasta, "LEIA-ME.txt")
	{ // sempre regrava: o texto pode mudar entre versões
		_ = os.WriteFile(leia, []byte("Coloque aqui uma imagem e/ou um vídeo para cada app, com o id do app no nome:\r\n"+
			"  netflix.jpg, hbomax.jpg, disney.jpg, appletv.jpg, youtube.jpg, prime.jpg, globoplay.jpg\r\n"+
			"Pode ser .jpg, .png ou .webp (de preferência 1600x900, mais escura à esquerda).\r\n"+
			"A imagem aparece no fundo e no cartão do app.\r\n\r\n"+
			"VÍDEOS: netflix.mp4, disney.webm... tocam sem som no fundo quando o app está selecionado.\r\n"+
			"  fundo.mp4 toca quando nenhum app está selecionado. Use vídeos curtos (10 a 60 s), 720p ou 1080p,\r\n"+
			"  em H.264 (mp4): é o formato que a placa de vídeo do notebook toca sem esforço.\r\n\r\n"+
			"Apague o arquivo para voltar à arte padrão.\r\n"), 0o644)
	}
}

func jaRodando(porta int) bool {
	cli := &http.Client{Timeout: 1500 * time.Millisecond}
	req, _ := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:%d/tv/api/mostrar", porta), strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := cli.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}
