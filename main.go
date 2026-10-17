// TV Central: transforma um PC com Windows numa central de TV,
// com launcher em tela cheia e controle pelo celular.
package main

import (
	"fmt"
	"io"
	"log"
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

	cfg, err := carregarConfig(dados)
	if err != nil {
		log.Printf("config: %v (usando padrão)", err)
	}
	par, err := novoPareamento(filepath.Join(dados, "celulares.json"))
	if err != nil {
		log.Printf("pareamento: %v", err)
	}
	ger := novoGerente(dados, cfg.Porta)
	srv := novoServidor(cfg, par, ger)
	if op.AtualizadoDe != "" {
		log.Printf("atualizado de %s para %s", op.AtualizadoDe, versao)
		srv.novidade = fmt.Sprintf("Atualizado para a versão %s", versao)
		srv.novidadeAte = time.Now().Add(2 * time.Minute)
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

	if !(op.LauncherPID > 0 && ger.AdotarLauncher(op.LauncherPID)) && cfg.abrirLauncher() {
		go ger.AbrirLauncher()
	}
	if cfg.atualizarAuto() {
		go srv.atu.Loop(dados, ger)
	}
	if err := srv.Rodar(ouvinte); err != nil {
		log.Printf("servidor parou: %v", err)
		os.Exit(1)
	}
}
