package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const portaDepuracao = 9222

// Gerente abre e fecha os apps e mantém o launcher.
type Gerente struct {
	mu           sync.Mutex
	dados        string
	porta        int
	atual        *exec.Cmd
	atualID      string
	cancelar     context.CancelFunc
	launcherPID  int
	launcherVivo bool
}

func novoGerente(dados string, porta int) *Gerente {
	return &Gerente{dados: dados, porta: porta}
}

// acharNavegador procura o Edge e, se não houver, o Chrome (aceitam os mesmos parâmetros).
func acharNavegador() (caminho string, edge bool) {
	pf, pf86, local := os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)"), os.Getenv("LOCALAPPDATA")
	for _, c := range []string{
		filepath.Join(pf86, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(pf, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(local, `Microsoft\Edge\Application\msedge.exe`),
	} {
		if existe(c) {
			return c, true
		}
	}
	for _, c := range []string{
		filepath.Join(pf, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(pf86, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(local, `Google\Chrome\Application\chrome.exe`),
	} {
		if existe(c) {
			return c, false
		}
	}
	for _, n := range []string{"microsoft-edge", "google-chrome", "chromium"} { // para testar fora do Windows
		if p, err := exec.LookPath(n); err == nil {
			return p, false
		}
	}
	return "", false
}

func existe(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil
}

func (g *Gerente) perfil(nome string) string { return filepath.Join(g.dados, "perfis", nome) }

func argsBase(perfil string) []string {
	return []string{
		"--user-data-dir=" + perfil,
		"--no-first-run", "--no-default-browser-check",
		"--disable-session-crashed-bubble", "--hide-crash-restore-bubble",
		"--autoplay-policy=no-user-gesture-required",
		"--disable-features=Translate,msEdgeShoppingAssistant",
	}
}

// AbrirLauncher abre a tela inicial em modo quiosque (ou traz para frente se já estiver aberta).
func (g *Gerente) AbrirLauncher() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.launcherVivo {
		go focarProcesso(g.launcherPID)
		return
	}
	nav, edge := acharNavegador()
	if nav == "" {
		log.Printf("launcher: Edge/Chrome não encontrado")
		avisar("TV Central", "Não achei o Microsoft Edge nem o Chrome. Instale o Edge para usar a tela inicial.")
		return
	}
	url := fmt.Sprintf("http://127.0.0.1:%d/tv/", g.porta)
	args := append(argsBase(g.perfil("launcher")), "--kiosk", url, "--disable-pinch", "--overscroll-history-navigation=0")
	if edge {
		args = append(args, "--edge-kiosk-type=fullscreen")
	}
	cmd := exec.Command(nav, args...)
	if err := cmd.Start(); err != nil {
		log.Printf("launcher: %v", err)
		return
	}
	log.Printf("launcher aberto (pid %d)", cmd.Process.Pid)
	g.vigiarLauncher(cmd.Process)
	go focarProcesso(cmd.Process.Pid)
}

// vigiarLauncher marca o launcher como aberto até o processo terminar. Chame com g.mu travado.
func (g *Gerente) vigiarLauncher(p *os.Process) {
	g.launcherPID, g.launcherVivo = p.Pid, true
	go func() {
		_, _ = p.Wait()
		g.mu.Lock()
		if g.launcherPID == p.Pid {
			g.launcherVivo = false
		}
		g.mu.Unlock()
		log.Printf("launcher fechou")
	}()
}

// AdotarLauncher assume a tela inicial aberta pela versão anterior (depois de uma atualização).
func (g *Gerente) AdotarLauncher(pid int) bool {
	if pid <= 0 || !processoVivo(pid) {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.vigiarLauncher(p)
	log.Printf("tela inicial mantida da versão anterior (pid %d)", pid)
	return true
}

func (g *Gerente) LauncherPID() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.launcherVivo {
		return 0
	}
	return g.launcherPID
}

// Abrir fecha o app atual e abre o novo em tela cheia.
func (g *Gerente) Abrir(app App) error {
	g.Fechar()
	g.mu.Lock()
	defer g.mu.Unlock()

	var cmd *exec.Cmd
	switch app.Tipo {
	case "programa":
		cmd = exec.Command(app.Destino, app.Args...)
		cmd.Dir = filepath.Dir(app.Destino)
	default:
		nav, _ := acharNavegador()
		if nav == "" {
			return fmt.Errorf("Edge/Chrome não encontrado")
		}
		perfil := app.Perfil
		if perfil == "" {
			perfil = "streaming"
		}
		args := append(argsBase(g.perfil(perfil)), "--start-fullscreen", "--app="+app.Destino)
		if app.UserAgent != "" {
			args = append(args, "--user-agent="+app.UserAgent)
		}
		if app.usaSetas() {
			// Só escuta em 127.0.0.1: usado para colocar a navegação por setas nas páginas.
			args = append(args, "--remote-debugging-port="+strconv.Itoa(portaDepuracao))
		}
		cmd = exec.Command(nav, args...)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	inicio := time.Now()
	g.atual, g.atualID = cmd, app.ID
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelar = cancel
	log.Printf("abriu %s (pid %d)", app.ID, cmd.Process.Pid)

	if app.Tipo != "programa" && app.usaSetas() {
		go injetarSetas(ctx, portaDepuracao, scriptSetas())
	}
	go focarProcesso(cmd.Process.Pid)
	go func() {
		_ = cmd.Wait()
		cancel()
		g.mu.Lock()
		eraAtual := g.atual == cmd
		if eraAtual {
			g.atual, g.atualID = nil, ""
		}
		g.mu.Unlock()
		// Saiu logo de cara: o navegador só repassou para uma janela já aberta.
		if eraAtual && time.Since(inicio) > 4*time.Second {
			log.Printf("%s fechou, voltando para a tela inicial", app.ID)
			g.AbrirLauncher()
		}
	}()
	return nil
}

// Fechar encerra o app aberto (e os processos filhos).
func (g *Gerente) Fechar() {
	g.mu.Lock()
	cmd, cancel := g.atual, g.cancelar
	g.atual, g.atualID, g.cancelar = nil, "", nil
	g.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if cmd != nil && cmd.Process != nil {
		matarArvore(cmd.Process.Pid)
		_ = cmd.Process.Kill()
	}
}

// Inicio fecha o app e volta para a tela inicial.
func (g *Gerente) Inicio() {
	g.Fechar()
	g.AbrirLauncher()
}

func (g *Gerente) Atual() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.atualID
}
