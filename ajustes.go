package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"unicode/utf8"
)

// Ajustes feitos pelo painel do celular (ficam em ajustes.json; o config.json continua valendo
// para o que não for mudado aqui).
type Ajustes struct {
	Ordem            []string `json:"ordem,omitempty"`   // ordem dos apps na tela inicial
	Ocultos          []string `json:"ocultos,omitempty"` // apps escondidos para todos
	Cidade           *string  `json:"cidade,omitempty"`
	Sons             *bool    `json:"sons,omitempty"`
	Intro            *bool    `json:"intro,omitempty"`
	VolumeIntro      *float64 `json:"volumeIntro,omitempty"`
	ProtetorMin      *int     `json:"protetorMin,omitempty"`
	VideosFundo      *bool    `json:"videosFundo,omitempty"`
	BloquearAnuncios *bool    `json:"bloquearAnuncios,omitempty"`
	Desempenho       string   `json:"desempenho,omitempty"` // "auto" (padrão), "leve" ou "completo"
	MostrarStats     bool     `json:"mostrarStats,omitempty"`
	Idioma           string   `json:"idioma,omitempty"` // auto, pt, en, es

	Extras          []AppExtra `json:"extras,omitempty"` // sites adicionados pelo painel
	Perfis          []Perfil   `json:"perfis"`
	PerfilAtual     string     `json:"perfilAtual"`
	PerguntarPerfil bool       `json:"perguntarPerfil"` // "Quem está assistindo?" ao ligar

	Crianca ModoCrianca `json:"crianca"`
}

type Perfil struct {
	ID      string   `json:"id"`
	Nome    string   `json:"nome"`
	Emoji   string   `json:"emoji"`
	Cor     string   `json:"cor"`
	Ocultos []string `json:"ocultos,omitempty"` // apps que este perfil não vê
}

type ModoCrianca struct {
	Ativo bool     `json:"ativo"`
	Apps  []string `json:"apps"`
	Sal   string   `json:"sal,omitempty"`
	Pin   string   `json:"pin,omitempty"` // hash do PIN
}

// Público: o que o celular e a TV podem ver (sem o hash do PIN).
type AjustesPublicos struct {
	Ajustes
	TemPin bool `json:"temPin"`
}

const perfilPrincipal = "principal"

type CofreAjustes struct {
	mu      sync.Mutex
	arquivo string
	a       Ajustes
}

func carregarAjustes(arquivo string) (*CofreAjustes, error) {
	c := &CofreAjustes{arquivo: arquivo}
	b, err := os.ReadFile(arquivo)
	if err == nil {
		err = json.Unmarshal(b, &c.a)
	} else if errors.Is(err, os.ErrNotExist) {
		err = nil
		c.a.Ocultos = []string{"youtubekids"} // só aparece no modo criança, a não ser que mostrem
	}
	c.normalizar()
	return c, err
}

func (c *CofreAjustes) normalizar() {
	if len(c.a.Perfis) == 0 {
		c.a.Perfis = []Perfil{{ID: perfilPrincipal, Nome: "Principal", Emoji: "😎", Cor: "#3d7bff"}}
	}
	if c.perfil(c.a.PerfilAtual) == nil {
		c.a.PerfilAtual = c.a.Perfis[0].ID
	}
	if c.a.Crianca.Apps == nil {
		c.a.Crianca.Apps = []string{"youtubekids", "disney"}
	}
}

func (c *CofreAjustes) perfil(id string) *Perfil {
	for i := range c.a.Perfis {
		if c.a.Perfis[i].ID == id {
			return &c.a.Perfis[i]
		}
	}
	return nil
}

func (c *CofreAjustes) salvar() error {
	b, _ := json.MarshalIndent(c.a, "", "  ")
	tmp := c.arquivo + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, c.arquivo)
}

// Ler devolve uma cópia (para usar sem segurar a trava).
func (c *CofreAjustes) Ler() Ajustes {
	c.mu.Lock()
	defer c.mu.Unlock()
	b, _ := json.Marshal(c.a)
	var copia Ajustes
	_ = json.Unmarshal(b, &copia)
	return copia
}

func (c *CofreAjustes) Publicos() AjustesPublicos {
	a := c.Ler()
	tem := a.Crianca.Pin != ""
	a.Crianca.Pin, a.Crianca.Sal = "", ""
	return AjustesPublicos{Ajustes: a, TemPin: tem}
}

// Mudar aplica uma alteração e salva. Com o modo criança ativo, nada muda.
func (c *CofreAjustes) Mudar(f func(a *Ajustes) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.a.Crianca.Ativo {
		return errModoCrianca
	}
	antes, _ := json.Marshal(c.a)
	if err := f(&c.a); err != nil {
		_ = json.Unmarshal(antes, &c.a)
		return err
	}
	c.normalizar()
	return c.salvar()
}

var errModoCrianca = errors.New("modo criança ativo: digite o PIN para sair antes de mudar algo")

func hashPin(sal, pin string) string {
	h := sha256.Sum256([]byte(sal + ":" + pin))
	return hex.EncodeToString(h[:])
}

func pinValido(pin string) bool {
	if len(pin) < 4 || len(pin) > 8 {
		return false
	}
	for _, c := range pin {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// EntrarCrianca liga o modo criança. Na primeira vez, define o PIN.
func (c *CofreAjustes) EntrarCrianca(pin string, apps []string) error {
	return c.Mudar(func(a *Ajustes) error {
		if a.Crianca.Pin == "" {
			if !pinValido(pin) {
				return errors.New("crie um PIN de 4 a 8 números")
			}
			a.Crianca.Sal = aleatorio(8)
			a.Crianca.Pin = hashPin(a.Crianca.Sal, pin)
		}
		if apps != nil {
			a.Crianca.Apps = limparIDs(apps)
		}
		if len(a.Crianca.Apps) == 0 {
			return errors.New("escolha pelo menos um app para o modo criança")
		}
		a.Crianca.Ativo = true
		return nil
	})
}

// SairCrianca desliga o modo criança se o PIN estiver certo.
func (c *CofreAjustes) SairCrianca(pin string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.a.Crianca.Ativo {
		return nil
	}
	if subtle.ConstantTimeCompare([]byte(hashPin(c.a.Crianca.Sal, pin)), []byte(c.a.Crianca.Pin)) != 1 {
		return errors.New("PIN errado")
	}
	c.a.Crianca.Ativo = false
	return c.salvar()
}

// TrocarPIN exige o PIN atual (se existir).
func (c *CofreAjustes) TrocarPIN(atual, novo string) error {
	return c.Mudar(func(a *Ajustes) error {
		if a.Crianca.Pin != "" && subtle.ConstantTimeCompare([]byte(hashPin(a.Crianca.Sal, atual)), []byte(a.Crianca.Pin)) != 1 {
			return errors.New("PIN atual errado")
		}
		if !pinValido(novo) {
			return errors.New("o PIN novo precisa ter de 4 a 8 números")
		}
		a.Crianca.Sal = aleatorio(8)
		a.Crianca.Pin = hashPin(a.Crianca.Sal, novo)
		return nil
	})
}

// TrocarPerfil funciona mesmo fora do painel (tela da TV); bloqueado no modo criança.
func (c *CofreAjustes) TrocarPerfil(id string) error {
	return c.Mudar(func(a *Ajustes) error {
		for _, p := range a.Perfis {
			if p.ID == id {
				a.PerfilAtual = id
				return nil
			}
		}
		return fmt.Errorf("perfil %q não existe", id)
	})
}

func limparIDs(ids []string) []string {
	out := []string{}
	for _, id := range ids {
		if idValido(id) && !slices.Contains(out, id) {
			out = append(out, id)
		}
	}
	return out
}

func limparTexto(s string, max int) string {
	s = strings.TrimSpace(strings.Map(func(r rune) rune {
		if r < 32 {
			return -1
		}
		return r
	}, s))
	for utf8.RuneCountInString(s) > max {
		r := []rune(s)
		s = string(r[:max])
	}
	return s
}

// ---------- o que cada um vê ----------

// AppsVisiveis aplica ordem, ocultos, perfil e modo criança.
func appsVisiveis(todos []App, a Ajustes) []App {
	porID := map[string]App{}
	for _, ap := range todos {
		porID[ap.ID] = ap
	}
	ordem := []App{}
	usados := map[string]bool{}
	for _, id := range a.Ordem {
		if ap, ok := porID[id]; ok && !usados[id] {
			ordem = append(ordem, ap)
			usados[id] = true
		}
	}
	for _, ap := range todos {
		if !usados[ap.ID] {
			ordem = append(ordem, ap)
		}
	}
	vis := []App{}
	if a.Crianca.Ativo {
		for _, ap := range ordem {
			if slices.Contains(a.Crianca.Apps, ap.ID) {
				vis = append(vis, ap)
			}
		}
		return vis
	}
	var ocultosPerfil []string
	for _, p := range a.Perfis {
		if p.ID == a.PerfilAtual {
			ocultosPerfil = p.Ocultos
		}
	}
	for _, ap := range ordem {
		if !slices.Contains(a.Ocultos, ap.ID) && !slices.Contains(ocultosPerfil, ap.ID) {
			vis = append(vis, ap)
		}
	}
	return vis
}

// perfilNavegador: cada perfil (e o modo criança) tem os próprios logins nos streamings.
// O perfil principal continua usando as pastas de antes, para ninguém perder o login.
func perfilNavegador(app App, a Ajustes) string {
	base := app.Perfil
	if base == "" {
		base = "streaming"
	}
	switch {
	case a.Crianca.Ativo:
		return base + "-crianca"
	case a.PerfilAtual == "" || a.PerfilAtual == perfilPrincipal:
		return base
	default:
		return base + "-" + a.PerfilAtual
	}
}
