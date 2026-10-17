package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// App é um item da tela inicial.
type App struct {
	ID        string   `json:"id"`
	Nome      string   `json:"nome"`
	Tipo      string   `json:"tipo"`    // "web" (abre no Edge em tela cheia) ou "programa" (.exe)
	Destino   string   `json:"destino"` // endereço do site ou caminho do .exe
	Args      []string `json:"args,omitempty"`
	Cor       string   `json:"cor"`
	Cor2      string   `json:"cor2,omitempty"`
	Descricao string   `json:"descricao,omitempty"` // frase que aparece na tela inicial
	UserAgent string   `json:"userAgent,omitempty"`
	Perfil    string   `json:"perfil,omitempty"` // perfil do navegador (logins ficam salvos nele)
	Setas     *bool    `json:"setas,omitempty"`  // navegação por setas injetada no site
	Oculto    bool     `json:"oculto,omitempty"`
}

func (a App) usaSetas() bool {
	if a.Setas != nil {
		return *a.Setas
	}
	return a.Tipo == "web"
}

func falso() *bool { b := false; return &b }

// YouTube tem interface de TV própria quando o navegador se apresenta como Smart TV.
const uaSmartTV = "Mozilla/5.0 (SMART-TV; Linux; Tizen 6.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.31 TV Safari/537.36"

var appsPadrao = []App{
	{ID: "netflix", Nome: "Netflix", Tipo: "web", Destino: "https://www.netflix.com/browse", Cor: "#E50914", Cor2: "#3d0205", Descricao: "Séries, filmes e documentários"},
	{ID: "hbomax", Nome: "HBO Max", Tipo: "web", Destino: "https://play.hbomax.com", Cor: "#7B2FF7", Cor2: "#1b0a3d", Descricao: "Séries e filmes da HBO e da Warner"},
	{ID: "disney", Nome: "Disney+", Tipo: "web", Destino: "https://www.disneyplus.com", Cor: "#0063E5", Cor2: "#0b1a4a", Descricao: "Disney, Pixar, Marvel, Star Wars e National Geographic"},
	{ID: "appletv", Nome: "Apple TV", Tipo: "web", Destino: "https://tv.apple.com", Cor: "#5c5c66", Cor2: "#111114", Descricao: "Séries e filmes originais da Apple"},
	{ID: "youtube", Nome: "YouTube", Tipo: "web", Destino: "https://www.youtube.com/tv", Cor: "#FF0033", Cor2: "#7a0019", UserAgent: uaSmartTV, Perfil: "youtube", Setas: falso(), Descricao: "Vídeos, música e canais, com tela de TV"},
	{ID: "youtubekids", Nome: "YouTube Kids", Tipo: "web", Destino: "https://www.youtubekids.com", Cor: "#FF4E45", Cor2: "#5b1a10", Descricao: "Vídeos para crianças"},
	{ID: "prime", Nome: "Prime Video", Tipo: "web", Destino: "https://www.primevideo.com", Cor: "#1A98FF", Cor2: "#0a2a4a", Descricao: "Filmes, séries e esportes da Amazon"},
	{ID: "globoplay", Nome: "Globoplay", Tipo: "web", Destino: "https://globoplay.globo.com", Cor: "#F25C2B", Cor2: "#4a1406", Descricao: "Novelas, séries e a Globo ao vivo"},
}

type Config struct {
	Porta         int               `json:"porta"`
	AbrirLauncher *bool             `json:"abrirLauncher,omitempty"`
	AtualizarAuto *bool             `json:"atualizarAuto,omitempty"`
	Intro         *bool             `json:"intro,omitempty"`       // abertura com música ao ligar
	VolumeIntro   *float64          `json:"volumeIntro,omitempty"` // 0 a 1 (padrão 0.5)
	Sons          *bool             `json:"sons,omitempty"`        // sons ao navegar na tela inicial
	VideosFundo   *bool             `json:"videosFundo,omitempty"` // tocar os vídeos da pasta capas no fundo
	Cidade        string            `json:"cidade,omitempty"`      // clima no topo (ex.: "Curitiba")
	ProtetorMin   *int              `json:"protetorMin,omitempty"` // minutos parado até o protetor de tela (0 desliga)
	Apps          []json.RawMessage `json:"apps"`
	apps          []App
}

func (c Config) abrirLauncher() bool { return c.AbrirLauncher == nil || *c.AbrirLauncher }
func (c Config) atualizarAuto() bool { return c.AtualizarAuto == nil || *c.AtualizarAuto }
func (c Config) intro() bool         { return c.Intro == nil || *c.Intro }
func (c Config) sons() bool          { return c.Sons == nil || *c.Sons }
func (c Config) videosFundo() bool   { return c.VideosFundo == nil || *c.VideosFundo }
func (c Config) protetorMin() int {
	if c.ProtetorMin == nil {
		return 10
	}
	return max(0, *c.ProtetorMin)
}
func (c Config) volumeIntro() float64 {
	if c.VolumeIntro == nil {
		return 0.5
	}
	return min(1, max(0, *c.VolumeIntro))
}

const exemploConfig = `{
  "porta": 8765,
  "abrirLauncher": true,
  "intro": true,
  "volumeIntro": 0.5,
  "sons": true,
  "cidade": "",
  "protetorMin": 10,
  "apps": []
}
`

// carregarConfig lê config.json. Cada item de "apps" com o mesmo id de um padrão
// altera só os campos informados (ex.: {"id":"prime","oculto":true}); ids novos viram apps novos.
func carregarConfig(dados string) (Config, error) {
	cfg := Config{Porta: 8765}
	arq := filepath.Join(dados, "config.json")
	b, err := os.ReadFile(arq)
	if errors.Is(err, os.ErrNotExist) {
		_ = os.WriteFile(arq, []byte(exemploConfig), 0o644)
		b, err = []byte(exemploConfig), nil
	}
	var errLeitura error
	if err == nil {
		errLeitura = json.Unmarshal(b, &cfg)
	} else {
		errLeitura = err
	}
	if cfg.Porta == 0 {
		cfg.Porta = 8765
	}
	cfg.apps = juntarApps(appsPadrao, cfg.Apps)
	return cfg, errLeitura
}

func juntarApps(padrao []App, extras []json.RawMessage) []App {
	lista := append([]App(nil), padrao...)
	for _, raw := range extras {
		var cab struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(raw, &cab) != nil || cab.ID == "" {
			continue
		}
		achou := false
		for i := range lista {
			if lista[i].ID == cab.ID {
				_ = json.Unmarshal(raw, &lista[i])
				achou = true
				break
			}
		}
		if !achou {
			var a App
			if json.Unmarshal(raw, &a) == nil && a.Destino != "" {
				if a.Tipo == "" {
					a.Tipo = "web"
				}
				if a.Cor == "" {
					a.Cor = "#3a3a48"
				}
				lista = append(lista, a)
			}
		}
	}
	vis := lista[:0]
	for _, a := range lista {
		if !a.Oculto {
			vis = append(vis, a)
		}
	}
	return vis
}
