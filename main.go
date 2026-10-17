// TV Central: transforma um PC com Windows numa central de TV,
// com launcher em tela cheia e controle pelo celular.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var versao = "0.1.0"

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
		rodar()
	case "versao", "version":
		fmt.Println("TV Central", versao)
	default:
		// Duplo clique no .exe baixado: instala. O .exe instalado: roda.
		if precisaInstalar() {
			sair(instalar())
			return
		}
		rodar()
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

func rodar() {
	dados := pastaDados()
	if err := os.MkdirAll(filepath.Join(dados, "logs"), 0o755); err != nil {
		avisar("TV Central", err.Error())
		return
	}
	f, err := os.OpenFile(filepath.Join(dados, "logs", "tvcentral.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
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

	if cfg.abrirLauncher() {
		go ger.AbrirLauncher()
	}
	if err := srv.Rodar(); err != nil {
		log.Printf("servidor parou: %v", err)
		avisar("TV Central", fmt.Sprintf("Não consegui usar a porta %d. Outro programa (por exemplo o controle antigo) está usando ela.\n\n%v", cfg.Porta, err))
		os.Exit(1)
	}
}
