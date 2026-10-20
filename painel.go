package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// Painel de ajustes do celular (todas as rotas exigem celular pareado).
func (s *Servidor) rotasPainel(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/ajustes", s.autenticado(s.lerAjustes))
	mux.HandleFunc("POST /api/ajustes", s.autenticado(s.mudarAjustes))
	mux.HandleFunc("POST /api/perfis", s.autenticado(s.mudarPerfis))
	mux.HandleFunc("POST /api/sites", s.autenticado(s.mudarExtras))
	mux.HandleFunc("POST /api/perfil/atual", s.autenticado(s.perfilAtual))
	mux.HandleFunc("POST /api/crianca/entrar", s.autenticado(s.criancaEntrar))
	mux.HandleFunc("POST /api/crianca/sair", s.autenticado(s.criancaSair))
	mux.HandleFunc("POST /api/crianca/pin", s.autenticado(s.criancaPin))
	mux.HandleFunc("POST /api/midia/{id}", s.autenticado(s.enviarMidia))
	mux.HandleFunc("DELETE /api/midia/{id}", s.autenticado(s.apagarMidia))
	// Tela da TV
	mux.HandleFunc("POST /tv/api/perfil", soLocalF(s.perfilAtual))
}

func lerJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(v); err != nil {
		erro(w, http.StatusBadRequest, "pedido inválido")
		return false
	}
	return true
}

func (s *Servidor) responderErro(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	status := http.StatusBadRequest
	if errors.Is(err, errModoCrianca) {
		status = http.StatusForbidden
	}
	erro(w, status, err.Error())
	return true
}

type appPainel struct {
	ID, Nome, Cor, Cor2 string
	TemCapa, TemVideo   bool
	Extra               bool
	Destino             string
}

func (s *Servidor) lerAjustes(w http.ResponseWriter, r *http.Request) {
	videos := s.midias.Videos()
	apps := []appPainel{}
	for _, a := range s.todosApps() {
		extra := slices.ContainsFunc(s.aj.Ler().Extras, func(e AppExtra) bool { return e.ID == a.ID })
		destino := ""
		if extra {
			destino = a.Destino
		}
		apps = append(apps, appPainel{a.ID, a.Nome, a.Cor, a.Cor2, s.capaPropria(a.ID) != "", videos[a.ID] != "", extra, destino})
	}
	ef := s.efetivo()
	responder(w, map[string]any{
		"ajustes": s.aj.Publicos(),
		"apps":    apps,
		"efetivo": ef,
		"fundo":   videos["fundo"] != "",
	})
}

// efetivo: o valor que está valendo (ajuste do painel ou, se não houver, o config.json).
func (s *Servidor) efetivo() map[string]any {
	return map[string]any{
		"cidade": s.cidade(), "sons": s.sons(), "intro": s.intro(), "volumeIntro": s.volumeIntro(),
		"protetorMin": s.protetorMin(), "videosFundo": s.videosFundo(), "bloquearAnuncios": s.bloquearAnuncios(), "desempenho": s.desempenho(), "mostrarStats": s.aj.Ler().MostrarStats, "idioma": s.idioma(),
	}
}

func (s *Servidor) cidade() string {
	if a := s.aj.Ler(); a.Cidade != nil {
		return *a.Cidade
	}
	return s.cfg.Cidade
}
func (s *Servidor) sons() bool {
	if a := s.aj.Ler(); a.Sons != nil {
		return *a.Sons
	}
	return s.cfg.sons()
}
func (s *Servidor) intro() bool {
	if a := s.aj.Ler(); a.Intro != nil {
		return *a.Intro
	}
	return s.cfg.intro()
}
func (s *Servidor) volumeIntro() float64 {
	if a := s.aj.Ler(); a.VolumeIntro != nil {
		return min(1, max(0, *a.VolumeIntro))
	}
	return s.cfg.volumeIntro()
}
func (s *Servidor) protetorMin() int {
	if a := s.aj.Ler(); a.ProtetorMin != nil {
		return max(0, *a.ProtetorMin)
	}
	return s.cfg.protetorMin()
}
func (s *Servidor) idioma() string {
	switch l := s.aj.Ler().Idioma; l {
	case "pt", "en", "es":
		return l
	}
	return "auto"
}

func (s *Servidor) desempenho() string {
	if d := s.aj.Ler().Desempenho; d == "leve" || d == "completo" {
		return d
	}
	return "auto"
}

func (s *Servidor) bloquearAnuncios() bool {
	if a := s.aj.Ler(); a.BloquearAnuncios != nil {
		return *a.BloquearAnuncios
	}
	return true
}

func (s *Servidor) videosFundo() bool {
	if a := s.aj.Ler(); a.VideosFundo != nil {
		return *a.VideosFundo
	}
	return s.cfg.videosFundo()
}

func (s *Servidor) mudarAjustes(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Ordem            *[]string
		Ocultos          *[]string
		Cidade           *string
		Sons             *bool
		Intro            *bool
		VolumeIntro      *float64
		ProtetorMin      *int
		VideosFundo      *bool
		PerguntarPerfil  *bool
		BloquearAnuncios *bool
		Desempenho       *string
		MostrarStats     *bool
		Idioma           *string
	}
	if !lerJSON(w, r, &req) {
		return
	}
	err := s.aj.Mudar(func(a *Ajustes) error {
		if req.Ordem != nil {
			a.Ordem = limparIDs(*req.Ordem)
		}
		if req.Ocultos != nil {
			a.Ocultos = limparIDs(*req.Ocultos)
		}
		if req.Cidade != nil {
			c := limparTexto(*req.Cidade, 60)
			a.Cidade = &c
		}
		if req.Sons != nil {
			a.Sons = req.Sons
		}
		if req.Intro != nil {
			a.Intro = req.Intro
		}
		if req.VolumeIntro != nil {
			v := min(1, max(0, *req.VolumeIntro))
			a.VolumeIntro = &v
		}
		if req.ProtetorMin != nil {
			m := min(240, max(0, *req.ProtetorMin))
			a.ProtetorMin = &m
		}
		if req.VideosFundo != nil {
			a.VideosFundo = req.VideosFundo
		}
		if req.Desempenho != nil {
			switch *req.Desempenho {
			case "auto", "leve", "completo":
				a.Desempenho = *req.Desempenho
			default:
				return errors.New("modo de desempenho inválido")
			}
		}
		if req.Idioma != nil {
			switch *req.Idioma {
			case "auto", "pt", "en", "es":
				a.Idioma = *req.Idioma
			default:
				return errors.New("idioma inválido")
			}
		}
		if req.MostrarStats != nil {
			a.MostrarStats = *req.MostrarStats
		}
		if req.BloquearAnuncios != nil {
			a.BloquearAnuncios = req.BloquearAnuncios
		}
		if req.PerguntarPerfil != nil {
			a.PerguntarPerfil = *req.PerguntarPerfil
		}
		return nil
	})
	if s.responderErro(w, err) {
		return
	}
	s.clima.Definir(s.cidade())
	responder(w, map[string]any{"ok": true})
}

var emojisPerfil = []string{"😎", "🦊", "🐼", "🦁", "🐸", "🦄", "🐱", "🐶", "🌸", "⚡", "🎮", "🎬"}
var coresPerfil = []string{"#3d7bff", "#e5484d", "#30a46c", "#f5a524", "#8e4ec6", "#12a594", "#e93d82", "#6e56cf"}

func (s *Servidor) mudarPerfis(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Acao   string // criar | editar | remover
		Perfil Perfil
	}
	if !lerJSON(w, r, &req) {
		return
	}
	p := req.Perfil
	p.Nome = limparTexto(p.Nome, 20)
	p.Emoji = limparTexto(p.Emoji, 4)
	p.Ocultos = limparIDs(p.Ocultos)
	var criado string
	err := s.aj.Mudar(func(a *Ajustes) error {
		switch req.Acao {
		case "criar":
			if p.Nome == "" {
				return errors.New("dê um nome ao perfil")
			}
			if len(a.Perfis) >= 8 {
				return errors.New("máximo de 8 perfis")
			}
			p.ID = "p" + aleatorio(3)
			if p.Emoji == "" {
				p.Emoji = emojisPerfil[len(a.Perfis)%len(emojisPerfil)]
			}
			if p.Cor == "" || !corValida(p.Cor) {
				p.Cor = coresPerfil[len(a.Perfis)%len(coresPerfil)]
			}
			a.Perfis = append(a.Perfis, p)
			if len(a.Perfis) == 2 {
				a.PerguntarPerfil = true // com mais de um, pergunta quem está assistindo
			}
			criado = p.ID
		case "editar":
			alvo := (&CofreAjustes{a: *a}).perfil(p.ID)
			if alvo == nil {
				return errors.New("perfil não existe")
			}
			for i := range a.Perfis {
				if a.Perfis[i].ID == p.ID {
					if p.Nome != "" {
						a.Perfis[i].Nome = p.Nome
					}
					if p.Emoji != "" {
						a.Perfis[i].Emoji = p.Emoji
					}
					if corValida(p.Cor) {
						a.Perfis[i].Cor = p.Cor
					}
					a.Perfis[i].Ocultos = p.Ocultos
				}
			}
		case "remover":
			if p.ID == perfilPrincipal {
				return errors.New("o perfil principal não pode ser apagado")
			}
			a.Perfis = slices.DeleteFunc(a.Perfis, func(x Perfil) bool { return x.ID == p.ID })
			if a.PerfilAtual == p.ID {
				a.PerfilAtual = perfilPrincipal
			}
		default:
			return errors.New("ação desconhecida")
		}
		return nil
	})
	if s.responderErro(w, err) {
		return
	}
	responder(w, map[string]any{"ok": true, "id": criado})
}

func corValida(c string) bool {
	if len(c) != 7 || c[0] != '#' {
		return false
	}
	for _, x := range strings.ToLower(c[1:]) {
		if !(x >= '0' && x <= '9' || x >= 'a' && x <= 'f') {
			return false
		}
	}
	return true
}

func (s *Servidor) perfilAtual(w http.ResponseWriter, r *http.Request) {
	var req struct{ ID string }
	if !lerJSON(w, r, &req) {
		return
	}
	antes := s.aj.Ler().PerfilAtual
	if s.responderErro(w, s.aj.TrocarPerfil(req.ID)) {
		return
	}
	if antes != req.ID {
		log.Printf("perfil: %s", req.ID)
		go s.ger.Inicio() // fecha o app do perfil anterior
	}
	responder(w, map[string]any{"ok": true})
}

func (s *Servidor) criancaEntrar(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pin  string
		Apps []string
	}
	if !lerJSON(w, r, &req) {
		return
	}
	if s.responderErro(w, s.aj.EntrarCrianca(req.Pin, req.Apps)) {
		return
	}
	log.Printf("modo criança ligado")
	go s.ger.Inicio()
	responder(w, map[string]any{"ok": true})
}

func (s *Servidor) criancaSair(w http.ResponseWriter, r *http.Request) {
	var req struct{ Pin string }
	if !lerJSON(w, r, &req) {
		return
	}
	// Erros de PIN contam como tentativas: depois de 5, espera 1 minuto
	if !s.tentativasPin.Permitir() {
		erro(w, http.StatusTooManyRequests, "muitas tentativas: espere 1 minuto")
		return
	}
	if err := s.aj.SairCrianca(req.Pin); err != nil {
		s.tentativasPin.Errou()
		erro(w, http.StatusForbidden, err.Error())
		return
	}
	s.tentativasPin.Zerar()
	log.Printf("modo criança desligado")
	go s.ger.Inicio()
	responder(w, map[string]any{"ok": true})
}

func (s *Servidor) criancaPin(w http.ResponseWriter, r *http.Request) {
	var req struct{ Atual, Novo string }
	if !lerJSON(w, r, &req) {
		return
	}
	if !s.tentativasPin.Permitir() {
		erro(w, http.StatusTooManyRequests, "muitas tentativas: espere 1 minuto")
		return
	}
	if err := s.aj.TrocarPIN(req.Atual, req.Novo); err != nil {
		if strings.Contains(err.Error(), "atual errado") {
			s.tentativasPin.Errou()
		}
		s.responderErro(w, err)
		return
	}
	responder(w, map[string]any{"ok": true})
}

// ---------- imagens e vídeos enviados pelo celular ----------

var extImagem = []string{".jpg", ".jpeg", ".png", ".webp"}

func (s *Servidor) capaPropria(id string) string {
	for _, ext := range extImagem {
		arq := filepath.Join(s.ger.dados, "capas", id+ext)
		if existe(arq) {
			return arq
		}
	}
	return ""
}

func (s *Servidor) enviarMidia(w http.ResponseWriter, r *http.Request) {
	if s.aj.Ler().Crianca.Ativo {
		erro(w, http.StatusForbidden, errModoCrianca.Error())
		return
	}
	id := r.PathValue("id")
	tipo := r.URL.Query().Get("tipo")
	if !idValido(id) || (tipo != "capa" && tipo != "video") {
		erro(w, http.StatusBadRequest, "pedido inválido")
		return
	}
	limite := int64(15 << 20)
	permitidas := extImagem
	if tipo == "video" {
		limite, permitidas = 400<<20, extVideo
	}
	r.Body = http.MaxBytesReader(w, r.Body, limite+1<<20)
	leitor, err := r.MultipartReader()
	if err != nil {
		erro(w, http.StatusBadRequest, "envie como formulário com arquivo")
		return
	}
	parte, err := leitor.NextPart()
	if err != nil {
		erro(w, http.StatusBadRequest, "arquivo não veio")
		return
	}
	ext := strings.ToLower(filepath.Ext(parte.FileName()))
	if !slices.Contains(permitidas, ext) {
		erro(w, http.StatusBadRequest, fmt.Sprintf("formato não aceito (use %s)", strings.Join(permitidas, ", ")))
		return
	}
	pasta := filepath.Join(s.ger.dados, "capas")
	_ = os.MkdirAll(pasta, 0o755)
	tmp, err := os.CreateTemp(pasta, ".envio-*")
	if err != nil {
		erro(w, 500, err.Error())
		return
	}
	n, err := io.Copy(tmp, io.LimitReader(parte, limite+1))
	tmp.Close()
	if err != nil || n > limite || n == 0 {
		os.Remove(tmp.Name())
		erro(w, http.StatusBadRequest, fmt.Sprintf("arquivo vazio ou maior que %d MB", limite>>20))
		return
	}
	s.removerMidia(id, tipo) // tira a versão anterior (pode ter outra extensão)
	destino := filepath.Join(pasta, id+ext)
	// No Windows o arquivo antigo pode estar aberto (tocando na TV): tenta de novo algumas vezes
	var errMover error
	for i := 0; i < 10; i++ {
		if errMover = os.Rename(tmp.Name(), destino); errMover == nil {
			break
		}
		s.removerMidia(id, tipo)
		time.Sleep(300 * time.Millisecond)
	}
	if errMover != nil {
		os.Remove(tmp.Name())
		erro(w, http.StatusConflict, "o arquivo antigo está em uso na TV: tente de novo em alguns segundos")
		return
	}
	s.midias.Esquecer()
	log.Printf("mídia nova: %s (%d KB)", filepath.Base(destino), n>>10)
	responder(w, map[string]any{"ok": true, "quando": time.Now().Unix()})
}

func (s *Servidor) removerMidia(id, tipo string) {
	exts := extImagem
	if tipo == "video" {
		exts = extVideo
	}
	for _, ext := range exts {
		_ = os.Remove(filepath.Join(s.ger.dados, "capas", id+ext))
	}
}

func (s *Servidor) apagarMidia(w http.ResponseWriter, r *http.Request) {
	if s.aj.Ler().Crianca.Ativo {
		erro(w, http.StatusForbidden, errModoCrianca.Error())
		return
	}
	id, tipo := r.PathValue("id"), r.URL.Query().Get("tipo")
	if !idValido(id) || (tipo != "capa" && tipo != "video") {
		erro(w, http.StatusBadRequest, "pedido inválido")
		return
	}
	s.removerMidia(id, tipo)
	s.midias.Esquecer()
	responder(w, map[string]any{"ok": true})
}
