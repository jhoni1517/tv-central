package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// De onde vêm as versões novas (Releases do GitHub).
const repoAtualizacao = "jhoni1517/tv-central"

var urlUltimaVersao = "https://api.github.com/repos/" + repoAtualizacao + "/releases/latest"

type Atualizador struct {
	mu       sync.Mutex
	status   string // texto para a tela inicial
	pronto   string // caminho do .exe novo já conferido
	novaVers string
	rodando  bool
	cli      *http.Client
}

func novoAtualizador() *Atualizador {
	return &Atualizador{cli: &http.Client{Timeout: 5 * time.Minute}}
}

func (a *Atualizador) Status() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.status
}

func (a *Atualizador) definir(s string) {
	a.mu.Lock()
	a.status = s
	a.mu.Unlock()
	if s != "" {
		log.Printf("atualização: %s", s)
	}
}

// numeros transforma "v1.2.3" em [1 2 3]. Versão "dev" (compilada na mão) nunca atualiza.
func numeros(v string) ([3]int, bool) {
	var n [3]int
	partes := strings.Split(strings.TrimPrefix(strings.TrimSpace(v), "v"), ".")
	if len(partes) != 3 {
		return n, false
	}
	for i, p := range partes {
		x, err := strconv.Atoi(p)
		if err != nil {
			return n, false
		}
		n[i] = x
	}
	return n, true
}

func maisNova(nova, atual string) bool {
	a, ok1 := numeros(nova)
	b, ok2 := numeros(atual)
	if !ok1 || !ok2 {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

type release struct {
	Tag    string `json:"tag_name"`
	Assets []struct {
		Nome string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func (a *Atualizador) pegar(url string, max int64) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "TVCentral/"+versao)
	resp, err := a.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, max))
}

// Verificar procura uma versão nova, baixa e confere o SHA-256. Devolve true se há versão pronta.
func (a *Atualizador) Verificar(dados string) (bool, error) {
	a.mu.Lock()
	if a.rodando {
		a.mu.Unlock()
		return false, errors.New("já estou verificando")
	}
	if a.pronto != "" {
		a.mu.Unlock()
		return true, nil
	}
	a.rodando = true
	a.mu.Unlock()
	defer func() { a.mu.Lock(); a.rodando = false; a.mu.Unlock() }()

	if _, ok := numeros(versao); !ok {
		return false, errors.New("versão de desenvolvimento não se atualiza")
	}
	b, err := a.pegar(urlUltimaVersao, 1<<20)
	if err != nil {
		if strings.Contains(err.Error(), "HTTP 404") {
			return false, nil // nenhuma versão publicada ainda
		}
		return false, err
	}
	var r release
	if err := json.Unmarshal(b, &r); err != nil {
		return false, err
	}
	if !maisNova(r.Tag, versao) {
		a.definir("")
		return false, nil
	}
	var urlExe, urlHash string
	for _, as := range r.Assets {
		switch as.Nome {
		case "TVCentral.exe":
			urlExe = as.URL
		case "TVCentral.exe.sha256":
			urlHash = as.URL
		}
	}
	if urlExe == "" || urlHash == "" {
		return false, fmt.Errorf("versão %s sem o programa ou sem o código de conferência", r.Tag)
	}
	a.definir("Baixando versão " + strings.TrimPrefix(r.Tag, "v") + "…")
	hb, err := a.pegar(urlHash, 4096)
	if err != nil {
		return false, err
	}
	esperado := strings.ToLower(strings.Fields(string(hb) + " x")[0])
	exe, err := a.pegar(urlExe, 200<<20)
	if err != nil {
		a.definir("")
		return false, err
	}
	soma := sha256.Sum256(exe)
	if hex.EncodeToString(soma[:]) != esperado {
		a.definir("")
		return false, errors.New("o arquivo baixado não confere com o SHA-256: descartado")
	}
	pasta := filepath.Join(dados, "atualizacao")
	_ = os.MkdirAll(pasta, 0o755)
	destino := filepath.Join(pasta, "TVCentral.exe")
	if err := os.WriteFile(destino, exe, 0o755); err != nil {
		return false, err
	}
	a.mu.Lock()
	a.pronto, a.novaVers = destino, strings.TrimPrefix(r.Tag, "v")
	a.mu.Unlock()
	a.definir("Versão " + strings.TrimPrefix(r.Tag, "v") + " pronta: instala quando nenhum app estiver aberto")
	return true, nil
}

// Aplicar troca o programa pela versão baixada e reinicia. Só retorna se der erro.
func (a *Atualizador) Aplicar(ger *Gerente) error {
	a.mu.Lock()
	novo, nv := a.pronto, a.novaVers
	a.mu.Unlock()
	if novo == "" {
		return errors.New("nenhuma versão baixada")
	}
	a.definir("Instalando versão " + nv + "…")
	if err := trocarPrograma(novo, ger.LauncherPID()); err != nil {
		a.mu.Lock()
		a.pronto = ""
		a.mu.Unlock()
		a.definir("")
		return err
	}
	log.Printf("atualizando para %s: reiniciando", nv)
	time.Sleep(300 * time.Millisecond) // deixa a resposta HTTP sair
	os.Exit(0)
	return nil
}

// Loop verifica 90 s depois de abrir e a cada 6 h; instala quando não há app aberto.
func (a *Atualizador) Loop(dados string, ger *Gerente) {
	time.Sleep(90 * time.Second)
	ultima := time.Time{}
	for {
		if time.Since(ultima) > 6*time.Hour {
			ultima = time.Now()
			if _, err := a.Verificar(dados); err != nil {
				log.Printf("atualização: %v", err)
			}
		}
		a.mu.Lock()
		pronto := a.pronto != ""
		a.mu.Unlock()
		if pronto && ger.Atual() == "" {
			if err := a.Aplicar(ger); err != nil {
				log.Printf("atualização falhou: %v", err)
			}
		}
		time.Sleep(30 * time.Second)
	}
}

func copiarArquivo(de, para string) error {
	b, err := os.ReadFile(de)
	if err != nil {
		return err
	}
	return os.WriteFile(para, b, 0o755)
}
