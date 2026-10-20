package main

import (
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

// Sites que a pessoa adiciona pelo painel do celular.
type AppExtra struct {
	ID      string `json:"id"`
	Nome    string `json:"nome"`
	Destino string `json:"destino"`
	Cor     string `json:"cor"`
}

func (e AppExtra) app() App {
	return App{ID: e.ID, Nome: e.Nome, Tipo: "web", Destino: e.Destino, Cor: e.Cor, Cor2: "#101018", Descricao: hostDe(e.Destino)}
}

func hostDe(u string) string {
	p, err := url.Parse(u)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(p.Hostname(), "www.")
}

// todosApps: os do programa + config.json + os adicionados pelo painel.
func (s *Servidor) todosApps() []App {
	l := append([]App(nil), s.cfg.apps...)
	for _, e := range s.aj.Ler().Extras {
		if !slices.ContainsFunc(l, func(a App) bool { return a.ID == e.ID }) {
			l = append(l, e.app())
		}
	}
	return l
}

func normalizarEndereco(bruto string) (string, error) {
	bruto = strings.TrimSpace(bruto)
	if bruto == "" || len(bruto) > 500 {
		return "", errors.New("coloque o endereço do site")
	}
	if !strings.Contains(bruto, "://") {
		bruto = "https://" + bruto
	}
	u, err := url.Parse(bruto)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || !strings.Contains(u.Hostname(), ".") {
		return "", errors.New("endereço inválido (ex.: crunchyroll.com)")
	}
	return u.String(), nil
}

func (s *Servidor) mudarExtras(w http.ResponseWriter, r *http.Request) {
	var req struct{ Acao, ID, Nome, Destino, Cor string }
	if !lerJSON(w, r, &req) {
		return
	}
	var criado string
	err := s.aj.Mudar(func(a *Ajustes) error {
		switch req.Acao {
		case "criar":
			destino, err := normalizarEndereco(req.Destino)
			if err != nil {
				return err
			}
			nome := limparTexto(req.Nome, 24)
			if nome == "" {
				nome = hostDe(destino)
			}
			if len(a.Extras) >= 30 {
				return errors.New("máximo de 30 sites")
			}
			cor := req.Cor
			if !corValida(cor) {
				cor = coresPerfil[len(a.Extras)%len(coresPerfil)]
			}
			criado = "s" + aleatorio(3)
			a.Extras = append(a.Extras, AppExtra{ID: criado, Nome: nome, Destino: destino, Cor: cor})
		case "remover":
			antes := len(a.Extras)
			a.Extras = slices.DeleteFunc(a.Extras, func(e AppExtra) bool { return e.ID == req.ID })
			if len(a.Extras) == antes {
				return errors.New("site não existe")
			}
			a.Ordem = slices.DeleteFunc(a.Ordem, func(id string) bool { return id == req.ID })
		default:
			return errors.New("ação desconhecida")
		}
		return nil
	})
	if s.responderErro(w, err) {
		return
	}
	if req.Acao == "remover" && idValido(req.ID) {
		s.removerMidia(req.ID, "capa")
		s.removerMidia(req.ID, "video")
	}
	responder(w, map[string]any{"ok": true, "id": criado})
}
