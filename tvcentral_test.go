package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestMaisNova(t *testing.T) {
	casos := []struct {
		nova, atual string
		quer        bool
	}{
		{"v0.2.0", "0.1.0", true}, {"v0.1.0", "0.1.0", false}, {"v0.10.0", "0.9.9", true},
		{"v1.0.0", "1.0.1", false}, {"v0.2.0", "dev", false}, {"lixo", "0.1.0", false},
	}
	for _, c := range casos {
		if got := maisNova(c.nova, c.atual); got != c.quer {
			t.Errorf("maisNova(%q,%q)=%v", c.nova, c.atual, got)
		}
	}
}

func TestJuntarApps(t *testing.T) {
	extras := []json.RawMessage{
		json.RawMessage(`{"id":"prime","oculto":true}`),
		json.RawMessage(`{"id":"netflix","nome":"Netflix BR"}`),
		json.RawMessage(`{"id":"novo","nome":"Novo","destino":"https://x"}`),
		json.RawMessage(`{"id":"semdestino","nome":"X"}`),
	}
	l := juntarApps(appsPadrao, extras)
	if l[0].Nome != "Netflix BR" || l[0].Destino == "" {
		t.Errorf("override parcial falhou: %+v", l[0])
	}
	for _, a := range l {
		if a.ID == "prime" || a.ID == "semdestino" {
			t.Errorf("%s deveria sumir", a.ID)
		}
	}
	if l[len(l)-1].ID != "novo" || l[len(l)-1].Tipo != "web" {
		t.Errorf("app novo: %+v", l[len(l)-1])
	}
	if appsPadrao[0].Nome != "Netflix" {
		t.Error("padrão foi alterado")
	}
}

func TestPareamento(t *testing.T) {
	p, _ := novoPareamento(filepath.Join(t.TempDir(), "c.json"))
	if _, err := p.Confirmar("123456", "x"); err == nil {
		t.Fatal("sem código ativo devia falhar")
	}
	p.IniciarCodigo()
	cod := p.Pendente()
	for i := 0; i < 4; i++ {
		if _, err := p.Confirmar("errado", "x"); err == nil {
			t.Fatal("código errado aceito")
		}
	}
	tok, err := p.Confirmar(cod, "cel")
	if err != nil || !p.Valido(tok) || p.Valido("x"+tok[1:]) {
		t.Fatalf("token: %v", err)
	}
	if p.Pendente() != "" {
		t.Error("código devia ser de uso único")
	}
}

func TestVerificarAtualizacao(t *testing.T) {
	exe := []byte("programa novo")
	soma := sha256.Sum256(exe)
	hashErrado := false
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest":
			fmt.Fprintf(w, `{"tag_name":"v9.0.0","assets":[{"name":"TVCentral.exe","browser_download_url":"%s/exe"},{"name":"TVCentral.exe.sha256","browser_download_url":"%s/sha"}]}`, srv.URL, srv.URL)
		case "/exe":
			w.Write(exe)
		case "/sha":
			if hashErrado {
				fmt.Fprint(w, "00 TVCentral.exe")
				return
			}
			fmt.Fprintf(w, "%s  TVCentral.exe\n", hex.EncodeToString(soma[:]))
		}
	}))
	defer srv.Close()
	urlUltimaVersao = srv.URL + "/latest"
	versao = "0.1.0"
	defer func() { versao = "dev" }()

	dados := t.TempDir()
	a := novoAtualizador()
	tem, err := a.Verificar(dados)
	if err != nil || !tem {
		t.Fatalf("devia achar versão: %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(dados, "atualizacao", "TVCentral.exe"))
	if string(b) != string(exe) {
		t.Error("arquivo baixado diferente")
	}

	hashErrado = true
	a2 := novoAtualizador()
	if tem, err := a2.Verificar(t.TempDir()); tem || err == nil {
		t.Error("devia recusar arquivo com SHA-256 errado")
	}
}
