//go:build windows

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	procMessageBox          = user32.NewProc("MessageBoxW")
	procEnumWindows         = user32.NewProc("EnumWindows")
	procGetWindowThreadPID  = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible     = user32.NewProc("IsWindowVisible")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procBringWindowToTop    = user32.NewProc("BringWindowToTop")
	procShowWindow          = user32.NewProc("ShowWindow")
	procIsIconic            = user32.NewProc("IsIconic")
)

const nomeRegra = "TV Central (rede local)"

func avisar(titulo, msg string) {
	t, _ := windows.UTF16PtrFromString(titulo)
	m, _ := windows.UTF16PtrFromString(msg)
	procMessageBox.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), 0x40|0x10000) // info + primeiro plano
}

// semJanela roda um comando do sistema sem abrir janela preta.
func semJanela(nome string, args ...string) *exec.Cmd {
	c := exec.Command(nome, args...)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	return c
}

func matarArvore(pid int) {
	// Primeiro pede para fechar (o navegador grava logins e cookies); só força se não fechar.
	if !processoVivo(pid) {
		return
	}
	_ = semJanela("taskkill", "/PID", strconv.Itoa(pid), "/T").Run()
	for i := 0; i < 35 && processoVivo(pid); i++ {
		time.Sleep(150 * time.Millisecond)
	}
	if processoVivo(pid) {
		_ = semJanela("taskkill", "/PID", strconv.Itoa(pid), "/T", "/F").Run()
	}
}

var (
	muJanela     sync.Mutex
	pidProcura   uint32
	achada       uintptr
	callbackEnum = syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
		var pid uint32
		procGetWindowThreadPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
		if pid == pidProcura {
			if v, _, _ := procIsWindowVisible.Call(hwnd); v != 0 {
				achada = hwnd
				return 0
			}
		}
		return 1
	})
)

func janelaDo(pid int) uintptr {
	muJanela.Lock()
	defer muJanela.Unlock()
	pidProcura, achada = uint32(pid), 0
	procEnumWindows.Call(callbackEnum, 0)
	return achada
}

// focarProcesso traz a janela do processo para a frente (espera ela aparecer).
func focarProcesso(pid int) {
	for i := 0; i < 40; i++ {
		if h := janelaDo(pid); h != 0 {
			// O Windows só deixa trocar o primeiro plano depois de uma tecla: um Alt "solto" libera.
			_ = apertarTecla("alt")
			if min, _, _ := procIsIconic.Call(h); min != 0 {
				procShowWindow.Call(h, 9) // SW_RESTORE
			}
			procBringWindowToTop.Call(h)
			procSetForegroundWindow.Call(h)
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func acaoSistema(acao string) error {
	switch acao {
	case "desligar":
		return semJanela("shutdown", "/s", "/t", "0").Run()
	case "reiniciar":
		return semJanela("shutdown", "/r", "/t", "0").Run()
	}
	return fmt.Errorf("ação %q desconhecida", acao)
}

// ---------- instalação ----------

func pastaInstalacao() string { return filepath.Join(os.Getenv("LOCALAPPDATA"), "TVCentral") }
func exeInstalado() string    { return filepath.Join(pastaInstalacao(), "TVCentral.exe") }

func precisaInstalar() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	return !strings.EqualFold(filepath.Clean(exe), filepath.Clean(exeInstalado()))
}

func elevado() bool { return windows.GetCurrentProcessToken().IsElevated() }

// pedirAdmin roda este .exe de novo como administrador (aparece a pergunta do Windows).
func pedirAdmin(arg string) error {
	exe, _ := os.Executable()
	verbo, _ := windows.UTF16PtrFromString("runas")
	arq, _ := windows.UTF16PtrFromString(exe)
	par, _ := windows.UTF16PtrFromString(arg)
	if err := windows.ShellExecute(0, verbo, arq, par, nil, windows.SW_HIDE); err != nil {
		return fmt.Errorf("permissão de administrador negada")
	}
	return nil
}

func portaOcupada(porta int) bool {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", porta))
	if err != nil {
		return true
	}
	l.Close()
	return false
}

func instalar() error {
	if !elevado() {
		return pedirAdmin("instalar")
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	// Para a versão que estiver rodando (menos este processo)
	_ = semJanela("taskkill", "/F", "/IM", "TVCentral.exe", "/FI", fmt.Sprintf("PID ne %d", os.Getpid())).Run()
	time.Sleep(time.Second)

	if err := os.MkdirAll(pastaInstalacao(), 0o755); err != nil {
		return err
	}
	if !strings.EqualFold(filepath.Clean(exe), filepath.Clean(exeInstalado())) {
		b, err := os.ReadFile(exe)
		if err != nil {
			return err
		}
		if err := os.WriteFile(exeInstalado(), b, 0o755); err != nil {
			return fmt.Errorf("copiar programa: %w", err)
		}
	}
	_ = os.Remove(exeInstalado() + ":Zone.Identifier") // tira a marca de "baixado da internet"

	// Firewall: só a porta do controle, só este programa, só a rede local
	_ = semJanela("netsh", "advfirewall", "firewall", "delete", "rule", "name="+nomeRegra).Run()
	if out, err := semJanela("netsh", "advfirewall", "firewall", "add", "rule", "name="+nomeRegra,
		"dir=in", "action=allow", "program="+exeInstalado(), "protocol=TCP", "localport=8765",
		"remoteip=localsubnet", "profile=any").CombinedOutput(); err != nil {
		log.Printf("firewall: %v %s", err, out)
	}

	// Wi-Fi em desempenho máximo na tomada: o modo econômico atrasa os comandos do celular
	_ = semJanela("powercfg", "/setacvalueindex", "SCHEME_CURRENT", "19cbb8fa-5279-450e-9fac-8a3d5fedd0c1", "12bbebe6-58d6-4636-95bb-3217ef867c1a", "0").Run()
	_ = semJanela("powercfg", "/setactive", "SCHEME_CURRENT").Run()

	// Abrir junto com o Windows (sem Agendador de Tarefas e sem PowerShell)
	k, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("início automático: %w", err)
	}
	_ = k.SetStringValue("TVCentral", `"`+exeInstalado()+`"`)
	k.Close()

	aviso := ""
	if portaOcupada(8765) {
		aviso = "\n\nATENÇÃO: a porta 8765 está em uso por outro programa (provavelmente o controle antigo). Remova o controle antigo e reinicie o computador."
	}
	// Abre como usuário comum (o explorer não herda o administrador)
	_ = exec.Command("explorer.exe", exeInstalado()).Start()

	avisar("TV Central", "Instalado!\n\nA tela inicial vai abrir agora e sempre que o Windows ligar.\nNo celular (mesma Wi-Fi), escaneie o QR code da tela inicial."+aviso)
	return nil
}

func remover() error {
	if !elevado() {
		return pedirAdmin("remover")
	}
	_ = semJanela("taskkill", "/F", "/IM", "TVCentral.exe", "/FI", fmt.Sprintf("PID ne %d", os.Getpid())).Run()
	_ = semJanela("netsh", "advfirewall", "firewall", "delete", "rule", "name="+nomeRegra).Run()
	if k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE); err == nil {
		_ = k.DeleteValue("TVCentral")
		k.Close()
	}
	avisar("TV Central", "Removido do início do Windows e do firewall.\nSeus dados ficaram em "+pastaDados()+" (pode apagar a pasta se quiser).")
	return nil
}

// ---------- atualização ----------

func processoVivo(pid int) bool {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)
	var codigo uint32
	if windows.GetExitCodeProcess(h, &codigo) != nil {
		return false
	}
	return codigo == 259 // STILL_ACTIVE
}

// trocarPrograma: o Windows não deixa sobrescrever um .exe em uso, mas deixa renomear.
// O atual vira TVCentral.exe.antigo, o novo entra no lugar e é aberto (ele espera a porta liberar).
func trocarPrograma(novo string, launcherPID int) error {
	exe := exeInstalado()
	atual, _ := os.Executable()
	if !strings.EqualFold(filepath.Clean(atual), filepath.Clean(exe)) {
		return fmt.Errorf("só atualizo a cópia instalada (%s)", exe)
	}
	antigo := exe + ".antigo"
	_ = os.Remove(antigo)
	if err := os.Rename(exe, antigo); err != nil {
		return err
	}
	if err := copiarArquivo(novo, exe); err != nil {
		_ = os.Rename(antigo, exe)
		return err
	}
	cmd := exec.Command(exe, "servidor", "--launcher="+strconv.Itoa(launcherPID), "--atualizado-de="+versao)
	if err := cmd.Start(); err != nil {
		_ = os.Remove(exe)
		_ = os.Rename(antigo, exe)
		return err
	}
	return nil
}

// voltarVersaoAnterior: se a versão nova não conseguir abrir, volta a antiga.
func voltarVersaoAnterior() bool {
	exe := exeInstalado()
	antigo := exe + ".antigo"
	if !existe(antigo) {
		return false
	}
	ruim := exe + ".falhou"
	_ = os.Remove(ruim)
	if os.Rename(exe, ruim) != nil || os.Rename(antigo, exe) != nil {
		return false
	}
	return exec.Command(exe, "servidor").Start() == nil
}

// ehNavegador confere se o PID é mesmo do Edge/Chrome (o Windows reaproveita números de processo).
func ehNavegador(pid int) bool {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)
	buf := make([]uint16, windows.MAX_PATH)
	n := uint32(len(buf))
	if windows.QueryFullProcessImageName(h, 0, &buf[0], &n) != nil {
		return false
	}
	nome := strings.ToLower(filepath.Base(windows.UTF16ToString(buf[:n])))
	return nome == "msedge.exe" || nome == "chrome.exe"
}
