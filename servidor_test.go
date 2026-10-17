package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func servidorTeste(t *testing.T) (*Servidor, string) {
	t.Helper()
	dados := t.TempDir()
	cfg, _ := carregarConfig(dados)
	par, _ := novoPareamento(filepath.Join(dados, "c.json"))
	ger := novoGerente(dados, cfg.Porta)
	ger.naAreaDeTrabalho = true // sem navegador nos testes
	aj, _ := carregarAjustes(filepath.Join(dados, "ajustes.json"))
	return novoServidor(cfg, par, ger, aj), dados
}

func pedir(h http.Handler, metodo, url, corpo string, cab map[string]string, remoto string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(metodo, url, strings.NewReader(corpo))
	r.RemoteAddr = remoto
	for k, v := range cab {
		if k == "Host" {
			r.Host = v
		} else {
			r.Header.Set(k, v)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

const local, celular = "127.0.0.1:5000", "192.168.0.50:5000"

func TestSegurancaRotas(t *testing.T) {
	s, _ := servidorTeste(t)
	h := s.rotas()
	casos := []struct {
		nome, metodo, url, remoto string
		cab                       map[string]string
		quer                      int
	}{
		{"tela inicial pelo PC", "GET", "http://127.0.0.1:8765/tv/api/estado", local, nil, 200},
		{"tela inicial pelo celular", "GET", "http://192.168.0.10:8765/tv/api/estado", celular, nil, 403},
		{"DNS rebinding", "GET", "http://127.0.0.1:8765/tv/api/estado", local, map[string]string{"Host": "malicioso.com:8765"}, 403},
		{"CSRF por Origin", "POST", "http://127.0.0.1:8765/tv/api/timer", local, map[string]string{"Origin": "https://malicioso.com"}, 403},
		{"CSRF por Sec-Fetch", "POST", "http://127.0.0.1:8765/tv/api/timer", local, map[string]string{"Sec-Fetch-Site": "cross-site"}, 403},
		{"timer pela própria tela", "POST", "http://127.0.0.1:8765/tv/api/timer", local, map[string]string{"Origin": "http://127.0.0.1:8765"}, 200},
		{"apps sem pareamento", "GET", "http://192.168.0.10:8765/api/apps", celular, nil, 401},
		{"websocket sem pareamento", "GET", "http://192.168.0.10:8765/ws", celular, nil, 401},
		{"capa pública", "GET", "http://192.168.0.10:8765/capa/netflix", celular, nil, 200},
		{"capa com id inválido", "GET", "http://192.168.0.10:8765/capa/..%2f..%2fx", celular, nil, 404},
		{"vídeo que não existe", "GET", "http://192.168.0.10:8765/video/netflix", celular, nil, 404},
	}
	for _, c := range casos {
		if w := pedir(h, c.metodo, c.url, `{"min":0}`, c.cab, c.remoto); w.Code != c.quer {
			t.Errorf("%s: %d, queria %d (%s)", c.nome, w.Code, c.quer, w.Body.String())
		}
	}
}

func TestPareamentoPelaRede(t *testing.T) {
	s, _ := servidorTeste(t)
	h := s.rotas()
	if w := pedir(h, "POST", "http://192.168.0.10:8765/api/parear/iniciar", "", nil, celular); w.Code != 200 {
		t.Fatalf("iniciar: %d", w.Code)
	}
	cod := s.par.Pendente()
	w := pedir(h, "POST", "http://192.168.0.10:8765/api/parear", `{"codigo":"`+cod+`","nome":"teste"}`, nil, celular)
	var r struct{ Token string }
	_ = json.Unmarshal(w.Body.Bytes(), &r)
	if r.Token == "" {
		t.Fatalf("sem token: %s", w.Body.String())
	}
	if w := pedir(h, "GET", "http://192.168.0.10:8765/api/apps", "", map[string]string{"X-Token": r.Token}, celular); w.Code != 200 {
		t.Errorf("apps com token: %d", w.Code)
	}
}

func TestCapaPersonalizadaEVideos(t *testing.T) {
	s, dados := servidorTeste(t)
	h := s.rotas()
	_ = os.MkdirAll(filepath.Join(dados, "capas"), 0o755)
	_ = os.WriteFile(filepath.Join(dados, "capas", "netflix.png"), []byte("\x89PNG minha imagem"), 0o644)
	_ = os.WriteFile(filepath.Join(dados, "capas", "Disney.WEBM"), []byte("video"), 0o644)
	_ = os.WriteFile(filepath.Join(dados, "capas", "ruim nome.mp4"), []byte("video"), 0o644)
	if w := pedir(h, "GET", "http://x/capa/netflix", "", nil, celular); !strings.Contains(w.Body.String(), "minha imagem") {
		t.Error("não usou a imagem da pessoa")
	}
	ids := s.midias.IDs()
	if len(ids) != 1 || ids[0] != "disney" {
		t.Errorf("vídeos: %v", ids)
	}
	w := pedir(h, "GET", "http://x/video/disney", "", map[string]string{"Range": "bytes=0-1"}, celular)
	if w.Code != 206 || w.Header().Get("Content-Type") != "video/webm" {
		t.Errorf("vídeo: %d %s", w.Code, w.Header().Get("Content-Type"))
	}
}

func TestTimer(t *testing.T) {
	var d Timer
	d.Definir(30)
	if r := d.Restante(); r < 29*60 || r > 30*60 {
		t.Errorf("restante %d", r)
	}
	d.Definir(0)
	if d.Restante() != 0 {
		t.Error("não cancelou")
	}
	d.Definir(99999)
	if d.Restante() > 600*60 {
		t.Error("sem limite")
	}
	d.Definir(0)
}

func TestClimaCodigos(t *testing.T) {
	for _, c := range []int{0, 2, 3, 45, 53, 63, 81, 73, 95} {
		if ic, tx := tempo(c, true); ic == "" || tx == "" {
			t.Errorf("código %d sem texto", c)
		}
	}
}

func TestIDValido(t *testing.T) {
	for _, id := range []string{"netflix", "meu-app_2"} {
		if !idValido(id) {
			t.Errorf("%q devia valer", id)
		}
	}
	for _, id := range []string{"", "../x", "a b", "NETFLIX", strings.Repeat("a", 50)} {
		if idValido(id) {
			t.Errorf("%q não devia valer", id)
		}
	}
	_ = time.Now
}

func tokenTeste(t *testing.T, s *Servidor) string {
	s.par.IniciarCodigo()
	tok, err := s.par.Confirmar(s.par.Pendente(), "teste")
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func TestPainelEnvioECrianca(t *testing.T) {
	s, dados := servidorTeste(t)
	h := s.rotas()
	tok := tokenTeste(t, s)
	cab := map[string]string{"X-Token": tok}

	// envio de imagem pelo celular
	corpo := "--XX\r\nContent-Disposition: form-data; name=\"arquivo\"; filename=\"foto.JPG\"\r\nContent-Type: image/jpeg\r\n\r\nIMAGEM-NOVA\r\n--XX--\r\n"
	c := map[string]string{"X-Token": tok, "Content-Type": "multipart/form-data; boundary=XX"}
	if w := pedir(h, "POST", "http://x/api/midia/netflix?tipo=capa", corpo, c, celular); w.Code != 200 {
		t.Fatalf("envio: %d %s", w.Code, w.Body.String())
	}
	if w := pedir(h, "GET", "http://x/capa/netflix", "", nil, celular); w.Body.String() != "IMAGEM-NOVA" {
		t.Error("capa enviada não aparece")
	}
	ruim := strings.Replace(corpo, "foto.JPG", "virus.exe", 1)
	if w := pedir(h, "POST", "http://x/api/midia/netflix?tipo=capa", ruim, c, celular); w.Code != 400 {
		t.Errorf("aceitou .exe: %d", w.Code)
	}
	if w := pedir(h, "DELETE", "http://x/api/midia/netflix?tipo=capa", "", cab, celular); w.Code != 200 {
		t.Errorf("apagar: %d", w.Code)
	}
	if s.capaPropria("netflix") != "" {
		t.Error("não apagou")
	}
	_ = dados

	// ajustes e modo criança pela rede
	if w := pedir(h, "POST", "http://x/api/ajustes", `{"Cidade":"Curitiba","ProtetorMin":15}`, cab, celular); w.Code != 200 {
		t.Fatalf("ajustes: %d", w.Code)
	}
	if s.cidade() != "Curitiba" || s.protetorMin() != 15 {
		t.Error("ajuste não valeu")
	}
	if w := pedir(h, "POST", "http://x/api/crianca/entrar", `{"Pin":"4321","Apps":["youtubekids"]}`, cab, celular); w.Code != 200 {
		t.Fatalf("entrar: %d %s", w.Code, w.Body.String())
	}
	if err := s.abrirApp("netflix"); err == nil || !strings.Contains(err.Error(), "não está liberado") {
		t.Errorf("abriu app bloqueado: %v", err)
	}
	if w := pedir(h, "POST", "http://x/api/midia/netflix?tipo=capa", corpo, c, celular); w.Code != 403 {
		t.Errorf("enviou mídia no modo criança: %d", w.Code)
	}
	for i := 0; i < 5; i++ {
		pedir(h, "POST", "http://x/api/crianca/sair", `{"Pin":"0000"}`, cab, celular)
	}
	if w := pedir(h, "POST", "http://x/api/crianca/sair", `{"Pin":"4321"}`, cab, celular); w.Code != 429 {
		t.Errorf("sem bloqueio após 5 erros: %d", w.Code)
	}
	s.tentativasPin = Limitador{}
	if w := pedir(h, "POST", "http://x/api/crianca/sair", `{"Pin":"4321"}`, cab, celular); w.Code != 200 {
		t.Errorf("sair: %d", w.Code)
	}
}
