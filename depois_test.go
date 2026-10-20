package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDestinoDoLink(t *testing.T) {
	casos := []struct{ link, app, contem string }{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=10", "youtube", "/tv#/watch?v=dQw4w9WgXcQ"},
		{"https://youtu.be/dQw4w9WgXcQ", "youtube", "/tv#/watch?v=dQw4w9WgXcQ"},
		{"https://www.netflix.com/title/80100172", "netflix", "netflix.com/title"},
		{"https://play.max.com/show/abc", "hbomax", "max.com"},
		{"https://www.primevideo.com/detail/xyz", "prime", "primevideo"},
		{"https://globoplay.globo.com/v/123/", "globoplay", "globoplay"},
		{"https://tv.apple.com/br/show/x", "appletv", "tv.apple.com"},
		{"https://www.disneyplus.com/pt-br/series/x", "disney", "disneyplus"},
		{"https://exemplo.com.br/x", "", "exemplo"},
	}
	for _, c := range casos {
		a, l := destinoDoLink(c.link, appsPadrao)
		if a.ID != c.app || !strings.Contains(l, c.contem) {
			t.Errorf("%s -> %q %q", c.link, a.ID, l)
		}
	}
}

func TestExtrairLink(t *testing.T) {
	if l := extrairLink("Assista Stranger Things na Netflix: https://www.netflix.com/title/80057281."); l != "https://www.netflix.com/title/80057281" {
		t.Errorf("link: %q", l)
	}
	if extrairLink("sem link nenhum") != "" {
		t.Error("achou link onde não tem")
	}
}

func TestDepoisEFotosPelaRede(t *testing.T) {
	s, dados := servidorTeste(t)
	h := s.rotas()
	tok := tokenTeste(t, s)
	cab := map[string]string{"X-Token": tok}
	if w := pedir(h, "POST", "http://x/api/depois", `{"Texto":"olha isso https://youtu.be/dQw4w9WgXcQ"}`, cab, celular); w.Code != 200 {
		t.Fatalf("depois: %d %s", w.Code, w.Body.String())
	}
	if w := pedir(h, "POST", "http://x/api/depois", `{"Texto":"https://youtu.be/dQw4w9WgXcQ"}`, cab, celular); w.Code != 409 {
		t.Errorf("duplicado aceito: %d", w.Code)
	}
	if w := pedir(h, "POST", "http://x/api/depois", `{"Texto":"nada aqui"}`, cab, celular); w.Code != 400 {
		t.Errorf("sem link aceito: %d", w.Code)
	}
	itens := s.depois.Itens()
	if len(itens) != 1 || itens[0].App != "youtube" {
		t.Fatalf("itens: %+v", itens)
	}
	c2, _ := os.ReadFile(filepath.Join(dados, "depois.json"))
	if !strings.Contains(string(c2), "youtu.be") {
		t.Error("não salvou em disco")
	}
	if w := pedir(h, "DELETE", "http://x/api/depois/"+itens[0].ID, "", cab, celular); w.Code != 200 || len(s.depois.Itens()) != 0 {
		t.Errorf("remover: %d", w.Code)
	}
	// fotos
	corpo := "--XX\r\nContent-Disposition: form-data; name=\"f\"; filename=\"praia.JPEG\"\r\n\r\nFOTO1\r\n" +
		"--XX\r\nContent-Disposition: form-data; name=\"f\"; filename=\"x.heic\"\r\n\r\nFOTO2\r\n--XX--\r\n"
	c := map[string]string{"X-Token": tok, "Content-Type": "multipart/form-data; boundary=XX"}
	w := pedir(h, "POST", "http://x/api/fotos", corpo, c, celular)
	var r struct{ Salvas, Falhas []string }
	_ = json.Unmarshal(w.Body.Bytes(), &r)
	if len(r.Salvas) != 1 || len(r.Falhas) != 1 || !strings.HasSuffix(r.Salvas[0], ".jpg") {
		t.Fatalf("fotos: %s", w.Body.String())
	}
	nome := r.Salvas[0]
	if w := pedir(h, "GET", "http://x/foto/"+nome, "", nil, celular); w.Code != 401 {
		t.Errorf("foto aberta sem pareamento: %d", w.Code)
	}
	if w := pedir(h, "GET", "http://x/foto/"+nome+"?token="+tok, "", nil, celular); w.Body.String() != "FOTO1" {
		t.Errorf("foto com token: %d", w.Code)
	}
	if w := pedir(h, "GET", "http://127.0.0.1:8765/foto/..%2fajustes.json", "", nil, local); w.Code != 404 {
		t.Errorf("escapou da pasta: %d", w.Code)
	}
	if w := pedir(h, "POST", "http://x/api/fotos/mostrar", `{"Nome":"`+nome+`"}`, cab, celular); w.Code != 200 || s.fotos.Agora() == nil {
		t.Errorf("mostrar: %d", w.Code)
	}
	if w := pedir(h, "DELETE", "http://x/api/fotos/"+nome, "", cab, celular); w.Code != 200 || len(s.fotos.Lista()) != 0 {
		t.Errorf("apagar foto: %d", w.Code)
	}
}
