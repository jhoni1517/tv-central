package main

import (
	"path/filepath"
	"testing"
)

func ids(apps []App) []string {
	l := []string{}
	for _, a := range apps {
		l = append(l, a.ID)
	}
	return l
}

func TestPerfisEModoCrianca(t *testing.T) {
	arq := filepath.Join(t.TempDir(), "ajustes.json")
	c, _ := carregarAjustes(arq)
	a := c.Ler()
	vis := ids(appsVisiveis(appsPadrao, a))
	if vis[0] != "netflix" || contem(vis, "youtubekids") {
		t.Fatalf("padrão: %v", vis)
	}
	// ordem e ocultos
	_ = c.Mudar(func(a *Ajustes) error {
		a.Ordem = []string{"youtube", "netflix"}
		a.Ocultos = []string{"prime"}
		return nil
	})
	vis = ids(appsVisiveis(appsPadrao, c.Ler()))
	if vis[0] != "youtube" || vis[1] != "netflix" || contem(vis, "prime") || !contem(vis, "youtubekids") {
		t.Fatalf("ordem/ocultos: %v", vis)
	}
	// perfil com apps escondidos e pasta de login própria
	_ = c.Mudar(func(a *Ajustes) error {
		a.Perfis = append(a.Perfis, Perfil{ID: "pana", Nome: "Ana", Ocultos: []string{"hbomax"}})
		return nil
	})
	if err := c.TrocarPerfil("pana"); err != nil {
		t.Fatal(err)
	}
	a = c.Ler()
	if contem(ids(appsVisiveis(appsPadrao, a)), "hbomax") {
		t.Error("perfil devia esconder hbomax")
	}
	if got := perfilNavegador(App{ID: "netflix"}, a); got != "streaming-pana" {
		t.Errorf("pasta do perfil: %s", got)
	}
	_ = c.TrocarPerfil(perfilPrincipal)
	if got := perfilNavegador(App{ID: "youtube", Perfil: "youtube"}, c.Ler()); got != "youtube" {
		t.Errorf("principal mantém as pastas antigas: %s", got)
	}
	// modo criança
	if err := c.EntrarCrianca("12", nil); err == nil {
		t.Error("PIN curto aceito")
	}
	if err := c.EntrarCrianca("1234", []string{"youtubekids"}); err != nil {
		t.Fatal(err)
	}
	a = c.Ler()
	if vis := ids(appsVisiveis(appsPadrao, a)); len(vis) != 1 || vis[0] != "youtubekids" {
		t.Errorf("criança vê: %v", vis)
	}
	if perfilNavegador(App{ID: "disney"}, a) != "streaming-crianca" {
		t.Error("criança devia ter logins próprios")
	}
	if err := c.Mudar(func(a *Ajustes) error { a.Ocultos = nil; return nil }); err != errModoCrianca {
		t.Error("mudou ajustes no modo criança")
	}
	if err := c.TrocarPerfil("pana"); err == nil {
		t.Error("trocou perfil no modo criança")
	}
	if err := c.SairCrianca("0000"); err == nil {
		t.Error("saiu com PIN errado")
	}
	if c.Publicos().Crianca.Pin != "" {
		t.Error("hash do PIN vazou")
	}
	if err := c.SairCrianca("1234"); err != nil || c.Ler().Crianca.Ativo {
		t.Error("não saiu com o PIN certo")
	}
	// tudo salvo em disco
	c2, _ := carregarAjustes(arq)
	if a := c2.Ler(); len(a.Perfis) != 2 || a.Crianca.Pin == "" || a.Ordem[0] != "youtube" {
		t.Errorf("não salvou: %+v", a)
	}
}

func TestLimitador(t *testing.T) {
	var l Limitador
	for i := 0; i < 5; i++ {
		if !l.Permitir() {
			t.Fatal("bloqueou cedo")
		}
		l.Errou()
	}
	if l.Permitir() {
		t.Error("devia bloquear depois de 5 erros")
	}
}

func contem(l []string, x string) bool {
	for _, v := range l {
		if v == x {
			return true
		}
	}
	return false
}
