package main

import (
	"encoding/json"
	"errors"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
)

// "Assistir depois": links mandados pelo celular viram cartões na tela inicial.
type ItemDepois struct {
	ID     string    `json:"id"`
	URL    string    `json:"url"`
	Titulo string    `json:"titulo"`
	App    string    `json:"app"` // app que abre o link (para usar o login certo)
	Quando time.Time `json:"quando"`
}

type ListaDepois struct {
	mu      sync.Mutex
	arquivo string
	itens   []ItemDepois
}

func carregarDepois(arquivo string) *ListaDepois {
	l := &ListaDepois{arquivo: arquivo}
	if b, err := os.ReadFile(arquivo); err == nil {
		_ = json.Unmarshal(b, &l.itens)
	}
	return l
}

func (l *ListaDepois) salvar() {
	b, _ := json.MarshalIndent(l.itens, "", "  ")
	tmp := l.arquivo + ".tmp"
	if os.WriteFile(tmp, b, 0o600) == nil {
		_ = os.Rename(tmp, l.arquivo)
	}
}

func (l *ListaDepois) Itens() []ItemDepois {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]ItemDepois{}, l.itens...)
}

func (l *ListaDepois) Adicionar(it ItemDepois) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, x := range l.itens {
		if x.URL == it.URL {
			return errors.New("esse link já está na lista")
		}
	}
	if len(l.itens) >= 50 {
		l.itens = l.itens[:49]
	}
	l.itens = append([]ItemDepois{it}, l.itens...)
	l.salvar()
	return nil
}

func (l *ListaDepois) Mudar(id string, f func(*ItemDepois)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := range l.itens {
		if l.itens[i].ID == id {
			f(&l.itens[i])
			l.salvar()
			return
		}
	}
}

func (l *ListaDepois) Remover(id string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	antes := len(l.itens)
	l.itens = slices.DeleteFunc(l.itens, func(x ItemDepois) bool { return x.ID == id })
	if len(l.itens) != antes {
		l.salvar()
		return true
	}
	return false
}

// achar um link dentro do texto compartilhado (o Android costuma mandar "Assista X: https://...")
var reLink = regexp.MustCompile(`https?://[^\s<>"']+`)

func extrairLink(texto string) string {
	return strings.TrimRight(reLink.FindString(texto), ".,);!?")
}

var reYoutube = regexp.MustCompile(`(?:youtube\.com/(?:watch\?(?:.*&)?v=|shorts/|live/|embed/)|youtu\.be/)([A-Za-z0-9_-]{11})`)

// destinoDoLink decide qual app abre o link (pelo endereço) e ajusta o link quando precisa.
func destinoDoLink(link string, apps []App) (App, string) {
	if m := reYoutube.FindStringSubmatch(link); m != nil {
		for _, a := range apps {
			if a.ID == "youtube" {
				return a, "https://www.youtube.com/tv#/watch?v=" + m[1] // abre na tela de TV do YouTube
			}
		}
	}
	u, err := url.Parse(link)
	if err != nil {
		return App{}, link
	}
	host := strings.TrimPrefix(u.Hostname(), "www.")
	apelidos := map[string]string{"max.com": "hbomax", "play.max.com": "hbomax", "primevideo.com": "prime", "amazon.com.br": "prime", "globo.com": "globoplay"}
	for dom, id := range apelidos {
		if host == dom || strings.HasSuffix(host, "."+dom) {
			for _, a := range apps {
				if a.ID == id {
					return a, link
				}
			}
		}
	}
	for _, a := range apps {
		h := strings.TrimPrefix(hostDe(a.Destino), "play.")
		if h != "" && (host == h || strings.HasSuffix(host, "."+h) || strings.HasSuffix(h, "."+host)) {
			return a, link
		}
	}
	return App{}, link
}

var reTitulo = regexp.MustCompile(`(?is)<meta[^>]+property=["']og:title["'][^>]+content=["']([^"']+)|<title[^>]*>([^<]+)</title>`)

// buscarTitulo lê o título da página (em segundo plano; se falhar, fica o endereço).
func buscarTitulo(link string) string {
	if m := reYoutube.FindStringSubmatch(link); m != nil { // YouTube tem um endereço próprio para isso
		if t := tituloYoutube(m[1]); t != "" {
			return t
		}
	}
	t := tituloPagina(link)
	baixo := strings.ToLower(t)
	for _, ruim := range []string{"http", "ops", "404", "não encontrad", "not found", "erro", "error", "access denied", "just a moment", "sign in", "entrar"} {
		if strings.Contains(baixo, ruim) {
			return ""
		}
	}
	return t
}

func tituloYoutube(id string) string {
	cli := &http.Client{Timeout: 6 * time.Second}
	resp, err := cli.Get("https://www.youtube.com/oembed?format=json&url=" + url.QueryEscape("https://www.youtube.com/watch?v="+id))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var r struct{ Title string }
	if json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&r) != nil {
		return ""
	}
	return limparTexto(r.Title, 80)
}

func tituloPagina(link string) string {
	u, err := url.Parse(link)
	if err != nil || u.Scheme != "https" {
		return ""
	}
	if ips, err := net.LookupIP(u.Hostname()); err != nil || len(ips) == 0 || slices.ContainsFunc(ips, func(ip net.IP) bool {
		return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsUnspecified()
	}) {
		return "" // não consulta endereços da rede de casa
	}
	cli := &http.Client{Timeout: 6 * time.Second}
	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120 Safari/537.36")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9")
	resp, err := cli.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 512<<10))
	m := reTitulo.FindSubmatch(b)
	if m == nil {
		return ""
	}
	t := string(m[1])
	if t == "" {
		t = string(m[2])
	}
	t = strings.TrimSpace(html.UnescapeString(t))
	for _, fim := range []string{" - YouTube", " | Netflix", " | Disney+", " | Max", " | Prime Video"} {
		t = strings.TrimSuffix(t, fim)
	}
	return limparTexto(t, 80)
}

// ---------- rotas ----------

func (s *Servidor) rotasDepois(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/depois", s.autenticado(func(w http.ResponseWriter, r *http.Request) {
		responder(w, map[string]any{"itens": s.depois.Itens()})
	}))
	mux.HandleFunc("POST /api/depois", s.autenticado(s.adicionarDepois))
	mux.HandleFunc("DELETE /api/depois/{id}", s.autenticado(func(w http.ResponseWriter, r *http.Request) {
		if s.aj.Ler().Crianca.Ativo {
			erro(w, http.StatusForbidden, errModoCrianca.Error())
			return
		}
		if !s.depois.Remover(r.PathValue("id")) {
			erro(w, http.StatusNotFound, "não está na lista")
			return
		}
		responder(w, map[string]any{"ok": true})
	}))
	// O celular abre /compartilhar quando a pessoa usa "Compartilhar → TV Central"
	mux.HandleFunc("GET /compartilhar", func(w http.ResponseWriter, r *http.Request) {
		b, _ := arquivosWeb.ReadFile("web/celular/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(b)
	})
	mux.HandleFunc("POST /tv/api/depois/abrir", soLocalF(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ ID string }
		if !lerJSON(w, r, &req) {
			return
		}
		if err := s.abrirDepois(req.ID); err != nil {
			erro(w, 400, err.Error())
			return
		}
		responder(w, map[string]any{"ok": true})
	}))
}

func (s *Servidor) adicionarDepois(w http.ResponseWriter, r *http.Request) {
	if s.aj.Ler().Crianca.Ativo {
		erro(w, http.StatusForbidden, errModoCrianca.Error())
		return
	}
	var req struct{ Texto string }
	if !lerJSON(w, r, &req) {
		return
	}
	link := extrairLink(req.Texto)
	if link == "" {
		if l, err := normalizarEndereco(req.Texto); err == nil && strings.Contains(req.Texto, ".") && !strings.Contains(strings.TrimSpace(req.Texto), " ") {
			link = l
		}
	}
	if link == "" || len(link) > 1000 {
		erro(w, http.StatusBadRequest, "não achei um link aí: copie o link do filme, série ou vídeo")
		return
	}
	app, _ := destinoDoLink(link, s.todosApps())
	it := ItemDepois{ID: "d" + aleatorio(4), URL: link, Titulo: hostDe(link), App: app.ID, Quando: time.Now()}
	if err := s.depois.Adicionar(it); err != nil {
		erro(w, http.StatusConflict, err.Error())
		return
	}
	log.Printf("assistir depois: %s", link)
	go func() {
		if t := buscarTitulo(link); t != "" {
			s.depois.Mudar(it.ID, func(x *ItemDepois) { x.Titulo = t })
		}
	}()
	responder(w, map[string]any{"ok": true, "item": it})
}

func (s *Servidor) abrirDepois(id string) error {
	if s.aj.Ler().Crianca.Ativo {
		return errModoCrianca
	}
	for _, it := range s.depois.Itens() {
		if it.ID != id {
			continue
		}
		apps := s.todosApps()
		app, link := destinoDoLink(it.URL, apps)
		if app.ID == "" {
			app = App{ID: "web", Nome: "Site", Tipo: "web"}
		}
		app.Destino = link
		app.Perfil = perfilNavegador(app, s.aj.Ler())
		return s.ger.Abrir(app)
	}
	return errors.New("não está mais na lista")
}
