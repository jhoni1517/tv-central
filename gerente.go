package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const portaDepuracao = 39222 // pouco usada por outros programas

// Gerente abre e fecha os apps e mantém o launcher.
type Gerente struct {
	mu           sync.Mutex
	trocando     sync.Mutex // um abrir/fechar por vez (dois toques rápidos não abrem dois apps)
	dados        string
	porta        int
	atual        *os.Process
	atualWeb     bool
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
	_ = os.WriteFile(filepath.Join(g.dados, "launcher.pid"), []byte(strconv.Itoa(p.Pid)), 0o644)
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
// Também serve depois de reiniciar o programa: o PID fica salvo em launcher.pid.
func (g *Gerente) AdotarLauncher(pid int) bool {
	if pid <= 0 {
		b, _ := os.ReadFile(filepath.Join(g.dados, "launcher.pid"))
		pid, _ = strconv.Atoi(strings.TrimSpace(string(b)))
	}
	if pid <= 0 || !processoVivo(pid) || !ehNavegador(pid) {
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
	g.trocando.Lock()
	defer g.trocando.Unlock()
	g.fechar()
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
		// Porta de depuração (só em 127.0.0.1): coloca as setas nas páginas e fecha o
		// navegador do jeito certo, salvando os logins.
		args = append(args, "--remote-debugging-port="+strconv.Itoa(portaDepuracao))
		cmd = exec.Command(nav, args...)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	log.Printf("abriu %s (pid %d)", app.ID, cmd.Process.Pid)
	g.vigiarApp(cmd.Process, app)
	go focarProcesso(cmd.Process.Pid)
	return nil
}

// vigiarApp acompanha o app aberto e volta para a tela inicial quando ele fechar. Chame com g.mu travado.
func (g *Gerente) vigiarApp(p *os.Process, app App) {
	inicio := time.Now()
	g.atual, g.atualID, g.atualWeb = p, app.ID, app.Tipo != "programa"
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelar = cancel
	_ = os.WriteFile(filepath.Join(g.dados, "app.pid"), []byte(app.ID+" "+strconv.Itoa(p.Pid)), 0o644)
	if app.Tipo != "programa" && app.usaSetas() {
		go injetarSetas(ctx, portaDepuracao, scriptSetas())
	}
	go func() {
		_, _ = p.Wait()
		cancel()
		g.mu.Lock()
		eraAtual := g.atual == p
		if eraAtual {
			g.atual, g.atualID = nil, ""
			_ = os.Remove(filepath.Join(g.dados, "app.pid"))
		}
		g.mu.Unlock()
		// Saiu logo de cara: o navegador só repassou para uma janela já aberta.
		if eraAtual && time.Since(inicio) > 4*time.Second {
			log.Printf("%s fechou, voltando para a tela inicial", app.ID)
			g.AbrirLauncher()
		}
	}()
}

// AdotarApp retoma o app que estava aberto antes do programa reiniciar (ex.: depois de atualizar).
func (g *Gerente) AdotarApp(apps []App) {
	b, err := os.ReadFile(filepath.Join(g.dados, "app.pid"))
	if err != nil {
		return
	}
	var id string
	var pid int
	if _, err := fmt.Sscanf(string(b), "%s %d", &id, &pid); err != nil || !processoVivo(pid) {
		return
	}
	for _, a := range apps {
		if a.ID == id {
			if a.Tipo != "programa" && !ehNavegador(pid) {
				return // o Windows reaproveitou o número do processo
			}
			if p, err := os.FindProcess(pid); err == nil {
				g.mu.Lock()
				g.vigiarApp(p, a)
				g.mu.Unlock()
				log.Printf("app %s continua aberto (pid %d)", id, pid)
			}
			return
		}
	}
}

// Fechar encerra o app aberto (e os processos filhos).
func (g *Gerente) Fechar() {
	g.trocando.Lock()
	defer g.trocando.Unlock()
	g.fechar()
}

func (g *Gerente) fechar() {
	g.mu.Lock()
	p, cancel, web := g.atual, g.cancelar, g.atualWeb
	g.atual, g.atualID, g.cancelar = nil, "", nil
	g.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if p != nil {
		// Pede ao próprio navegador para fechar: ele grava cookies e logins antes de sair.
		if web && fecharNavegador(portaDepuracao) {
			for i := 0; i < 40 && processoVivo(p.Pid); i++ {
				time.Sleep(150 * time.Millisecond)
			}
		}
		matarArvore(p.Pid) // só age se ainda estiver aberto
		_ = p.Kill()
		_ = os.Remove(filepath.Join(g.dados, "app.pid"))
	}
}

// Inicio fecha o app e volta para a tela inicial.
func (g *Gerente) Inicio() {
	g.trocando.Lock()
	g.fechar()
	g.trocando.Unlock()
	g.AbrirLauncher()
}

func (g *Gerente) Atual() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.atualID
}
