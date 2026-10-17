package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	qrcode "github.com/skip2/go-qrcode"
)

//go:embed web
var arquivosWeb embed.FS

func scriptSetas() string {
	b, _ := arquivosWeb.ReadFile("web/injetar/setas.js")
	return string(b)
}

type Servidor struct {
	cfg         Config
	par         *Pareamento
	ger         *Gerente
	atu         *Atualizador
	timer       Timer
	midias      *Midias
	clima       *Clima
	conectados  atomic.Int32 // celulares com o controle aberto
	novidade    string
	novidadeAte time.Time
}

func novoServidor(cfg Config, par *Pareamento, ger *Gerente) *Servidor {
	return &Servidor{cfg: cfg, par: par, ger: ger, atu: novoAtualizador(), midias: &Midias{pasta: filepath.Join(ger.dados, "capas")}, clima: novoClima(cfg.Cidade)}
}

// escutar tenta a porta por alguns segundos: depois de uma atualização a versão antiga ainda está saindo.
func escutar(porta int) (net.Listener, error) {
	var err error
	for i := 0; i < 30; i++ {
		var l net.Listener
		if l, err = net.Listen("tcp", fmt.Sprintf(":%d", porta)); err == nil {
			return l, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil, err
}

func (s *Servidor) Rodar(l net.Listener) error {
	mux := http.NewServeMux()
	celular, _ := fs.Sub(arquivosWeb, "web/celular")
	tv, _ := fs.Sub(arquivosWeb, "web/tv")

	// Celular (qualquer aparelho da rede local)
	mux.Handle("GET /", semCache(http.FileServer(http.FS(celular))))
	mux.HandleFunc("POST /api/parear/iniciar", s.iniciarPareamento)
	mux.HandleFunc("POST /api/parear", s.parear)
	mux.HandleFunc("GET /api/eu", s.autenticado(func(w http.ResponseWriter, r *http.Request) {
		responder(w, map[string]any{"ok": true, "versao": versao})
	}))
	mux.HandleFunc("GET /api/apps", s.autenticado(s.listarApps))
	mux.HandleFunc("GET /capa/{id}", s.capa)
	mux.HandleFunc("GET /video/{id}", s.midias.Servir)
	mux.HandleFunc("GET /api/diagnostico", s.autenticado(s.diagnostico))
	mux.HandleFunc("GET /ws", s.autenticado(s.controle))

	// Tela da TV (só o próprio computador)
	mux.Handle("GET /tv/", soLocal(semCache(http.StripPrefix("/tv/", http.FileServer(http.FS(tv))))))
	mux.HandleFunc("GET /tv/api/estado", soLocalF(s.estadoTV))
	mux.HandleFunc("GET /tv/qr.png", soLocalF(s.qr))
	mux.HandleFunc("POST /tv/api/abrir", soLocalF(s.abrirPelaTV))
	mux.HandleFunc("POST /tv/api/sistema", soLocalF(s.sistema))
	mux.HandleFunc("POST /tv/api/atualizar", soLocalF(s.atualizarAgora))
	mux.HandleFunc("POST /tv/api/mostrar", soLocalF(func(w http.ResponseWriter, r *http.Request) {
		go s.ger.AbrirLauncher()
		responder(w, map[string]any{"ok": true, "versao": versao})
	}))
	mux.HandleFunc("POST /tv/api/timer", soLocalF(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Min int }
		_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&req)
		s.timer.Definir(req.Min)
		responder(w, map[string]any{"timer": s.timer.Restante()})
	}))

	srv := &http.Server{Handler: mesmaOrigem(mux), ReadHeaderTimeout: 10 * time.Second}
	log.Printf("ouvindo na porta %d: %s", s.cfg.Porta, strings.Join(enderecos(s.cfg.Porta), " "))
	return srv.Serve(l)
}

func responder(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}

func erro(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"erro": msg})
}

func semCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		h.ServeHTTP(w, r)
	})
}

func ehLocal(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func soLocal(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Além de vir do próprio PC, o endereço tem que ser 127.0.0.1/localhost: impede que um site
		// qualquer aberto no navegador use um nome de domínio apontado para cá (DNS rebinding).
		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			host = r.Host
		}
		if !ehLocal(r) || (host != "127.0.0.1" && host != "localhost" && host != "[::1]" && host != "::1") {
			erro(w, http.StatusForbidden, "só no próprio computador")
			return
		}
		h.ServeHTTP(w, r)
	})
}

// mesmaOrigem barra pedidos POST disparados por outros sites (CSRF): o navegador manda o
// cabeçalho Origin, e ele tem que ser o próprio TV Central.
func mesmaOrigem(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if o := r.Header.Get("Origin"); o != "" && o != "http://"+r.Host && o != "https://"+r.Host {
				erro(w, http.StatusForbidden, "pedido de outro site recusado")
				return
			}
			if s := r.Header.Get("Sec-Fetch-Site"); s == "cross-site" || s == "same-site" {
				erro(w, http.StatusForbidden, "pedido de outro site recusado")
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func soLocalF(f http.HandlerFunc) http.HandlerFunc { return soLocal(f).ServeHTTP }

func token(r *http.Request) string {
	if t := r.Header.Get("X-Token"); t != "" {
		return t
	}
	return r.URL.Query().Get("token")
}

func (s *Servidor) autenticado(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.par.Valido(token(r)) {
			erro(w, http.StatusUnauthorized, "celular não pareado")
			return
		}
		f(w, r)
	}
}

func (s *Servidor) iniciarPareamento(w http.ResponseWriter, r *http.Request) {
	if s.par.IniciarCodigo() {
		s.ger.AbrirLauncher() // o código aparece na tela inicial
	} else if s.par.Pendente() == "" {
		erro(w, http.StatusTooManyRequests, "muitas tentativas: espere 1 minuto")
		return
	}
	responder(w, map[string]any{"ok": true})
}

func (s *Servidor) parear(w http.ResponseWriter, r *http.Request) {
	var req struct{ Codigo, Nome string }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		erro(w, http.StatusBadRequest, "pedido inválido")
		return
	}
	t, err := s.par.Confirmar(strings.TrimSpace(req.Codigo), req.Nome)
	if err != nil {
		erro(w, http.StatusForbidden, err.Error())
		return
	}
	log.Printf("celular pareado: %q", req.Nome)
	responder(w, map[string]string{"token": t})
}

type appPublico struct {
	ID, Nome, Cor, Cor2, Descricao string
}

func (s *Servidor) appsPublicos() []appPublico {
	l := make([]appPublico, 0, len(s.cfg.apps))
	for _, a := range s.cfg.apps {
		l = append(l, appPublico{a.ID, a.Nome, a.Cor, a.Cor2, a.Descricao})
	}
	return l
}

func (s *Servidor) listarApps(w http.ResponseWriter, r *http.Request) {
	responder(w, map[string]any{"apps": s.appsPublicos(), "atual": s.ger.Atual()})
}

// diagnostico mostra no celular o que está acontecendo no PC (para achar bugs).
func (s *Servidor) diagnostico(w http.ResponseWriter, r *http.Request) {
	nav, _ := acharNavegador()
	responder(w, map[string]any{
		"versao":      versao,
		"navegador":   nav,
		"enderecos":   enderecos(s.cfg.Porta),
		"appAberto":   s.ger.Atual(),
		"telaInicial": s.ger.LauncherPID() != 0,
		"atualizacao": s.atu.Status(),
		"log":         ultimasLinhas(filepath.Join(s.ger.dados, "logs", "tvcentral.log"), 60),
	})
}

func ultimasLinhas(arq string, n int) []string {
	b, err := os.ReadFile(arq)
	if err != nil {
		return nil
	}
	if len(b) > 64<<10 {
		b = b[len(b)-64<<10:]
	}
	linhas := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	if len(linhas) > n {
		linhas = linhas[len(linhas)-n:]
	}
	return linhas
}

// capa: imagem de fundo do app. Uma imagem sua em capas\<id>.jpg (ou .png/.webp) tem prioridade;
// senão usa a arte própria que vem no programa. ?p=1 devolve a versão pequena (cartão).
func (s *Servidor) capa(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !idValido(id) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".webp"} {
		arq := filepath.Join(s.ger.dados, "capas", id+ext)
		if f, err := os.Open(arq); err == nil {
			defer f.Close()
			if fi, err := f.Stat(); err == nil && fi.Size() < 15<<20 {
				http.ServeContent(w, r, arq, fi.ModTime(), f)
				return
			}
		}
	}
	sufixo := ""
	if r.URL.Query().Get("p") == "1" {
		sufixo = "-p"
	}
	b, err := arquivosWeb.ReadFile("web/tv/arte/" + id + sufixo + ".jpg")
	if err != nil {
		b, _ = arquivosWeb.ReadFile("web/tv/arte/padrao" + sufixo + ".jpg")
	}
	w.Header().Set("Content-Type", "image/jpeg")
	_, _ = w.Write(b)
}

func idValido(id string) bool {
	if id == "" || len(id) > 40 {
		return false
	}
	for _, c := range id {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func (s *Servidor) acharApp(id string) (App, bool) {
	for _, a := range s.cfg.apps {
		if a.ID == id {
			return a, true
		}
	}
	return App{}, false
}

func (s *Servidor) abrirApp(id string) error {
	a, ok := s.acharApp(id)
	if !ok {
		return fmt.Errorf("app %q não existe", id)
	}
	return s.ger.Abrir(a)
}

func (s *Servidor) estadoTV(w http.ResponseWriter, r *http.Request) {
	responder(w, map[string]any{
		"codigo":      s.par.Pendente(),
		"celulares":   s.par.Quantos(),
		"enderecos":   enderecos(s.cfg.Porta),
		"apps":        s.appsPublicos(),
		"versao":      versao,
		"atualizacao": s.atu.Status(),
		"novidade":    s.textoNovidade(),
		"intro":       s.cfg.intro(),
		"volumeIntro": s.cfg.volumeIntro(),
		"timer":       s.timer.Restante(),
		"videos":      s.midias.IDs(),
		"clima":       s.clima.Atual(),
		"conectados":  s.conectados.Load(),
		"videosFundo": s.cfg.videosFundo(),
		"sons":        s.cfg.sons(),
		"protetorMin": s.cfg.protetorMin(),
	})
}

func (s *Servidor) textoNovidade() string {
	if time.Now().Before(s.novidadeAte) {
		return s.novidade
	}
	return ""
}

// atualizarAgora: botão "Atualizar" da tela inicial (nenhum app aberto, então já instala).
func (s *Servidor) atualizarAgora(w http.ResponseWriter, r *http.Request) {
	tem, err := s.atu.Verificar(s.ger.dados)
	if err != nil {
		responder(w, map[string]any{"mensagem": "Não deu para verificar: " + err.Error()})
		return
	}
	if !tem {
		responder(w, map[string]any{"mensagem": "Você já tem a versão mais nova (" + versao + ")"})
		return
	}
	responder(w, map[string]any{"mensagem": "Instalando a versão nova…"})
	go func() {
		if err := s.atu.Aplicar(s.ger); err != nil {
			log.Printf("atualização: %v", err)
		}
	}()
}

func (s *Servidor) qr(w http.ResponseWriter, r *http.Request) {
	end := enderecos(s.cfg.Porta)
	if len(end) == 0 {
		http.NotFound(w, r)
		return
	}
	png, err := qrcode.Encode(end[0], qrcode.Medium, 320)
	if err != nil {
		erro(w, 500, err.Error())
		return
	}
	w.Header().Set("Content-Type", "image/png")
	_, _ = w.Write(png)
}

func (s *Servidor) abrirPelaTV(w http.ResponseWriter, r *http.Request) {
	var req struct{ ID string }
	_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req)
	if err := s.abrirApp(req.ID); err != nil {
		erro(w, 400, err.Error())
		return
	}
	responder(w, map[string]any{"ok": true})
}

func (s *Servidor) sistema(w http.ResponseWriter, r *http.Request) {
	var req struct{ Acao string }
	_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req)
	if err := acaoSistema(req.Acao); err != nil {
		erro(w, 400, err.Error())
		return
	}
	responder(w, map[string]any{"ok": true})
}

var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024} // aceita só a mesma origem

// Mensagem do controle do celular.
type Msg struct {
	T     string `json:"t"`     // tecla | texto | mover | clique | rolar | abrir | inicio | sistema | ping
	K     string `json:"k"`     // nome da tecla / botão / ação
	S     string `json:"s"`     // texto
	Enter bool   `json:"enter"` // texto: apertar Enter depois
	DX    int    `json:"dx"`
	DY    int    `json:"dy"`
	ID    int64  `json:"id"`  // ping: devolvido no pong para medir a latência
	Min   int    `json:"min"` // timer: minutos até desligar (0 cancela)
}

func (s *Servidor) controle(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	s.conectados.Add(1)
	defer s.conectados.Add(-1)
	c.SetReadLimit(8192)
	for {
		var m Msg
		if err := c.ReadJSON(&m); err != nil {
			return
		}
		switch m.T {
		case "ping":
			_ = c.WriteJSON(map[string]any{"t": "pong", "id": m.ID, "atual": s.ger.Atual(), "versao": versao, "timer": s.timer.Restante()})
		case "abrir", "inicio", "sistema":
			// Abrir/fechar app demora: roda à parte para as próximas teclas não ficarem na fila.
			go func(m Msg) {
				if err := s.executar(m); err != nil {
					log.Printf("controle: %s %s: %v", m.T, m.K, err)
				}
			}(m)
		default:
			if err := s.executar(m); err != nil {
				log.Printf("controle: %s %s: %v", m.T, m.K, err)
				_ = c.WriteJSON(map[string]string{"erro": err.Error()})
			}
		}
	}
}

func (s *Servidor) executar(m Msg) error {
	switch m.T {
	case "tecla":
		return apertarTecla(m.K)
	case "texto":
		if len([]rune(m.S)) > 500 {
			return fmt.Errorf("texto grande demais")
		}
		digitarTexto(m.S)
		if m.Enter {
			return apertarTecla("ok")
		}
	case "mover":
		moverMouse(limitar(m.DX, 400), limitar(m.DY, 400))
	case "clique":
		clicar(m.K)
	case "rolar":
		rolar(limitar(m.DY, 20))
	case "abrir":
		return s.abrirApp(m.K)
	case "inicio":
		s.ger.Inicio()
	case "sistema":
		return acaoSistema(m.K)
	case "timer":
		s.timer.Definir(m.Min)
	case "ping":
	default:
		return fmt.Errorf("comando desconhecido %q", m.T)
	}
	return nil
}

func limitar(v, max int) int {
	if v > max {
		return max
	}
	if v < -max {
		return -max
	}
	return v
}

// enderecos lista os endereços da rede local (o primeiro vai no QR code).
func enderecos(porta int) []string {
	var lista []string
	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}
		nome := strings.ToLower(i.Name)
		if strings.Contains(nome, "vethernet") || strings.Contains(nome, "virtual") || strings.Contains(nome, "vmware") || strings.HasPrefix(nome, "docker") {
			continue
		}
		addrs, _ := i.Addrs()
		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok || ipn.IP.To4() == nil || !ipn.IP.IsPrivate() {
				continue
			}
			lista = append(lista, fmt.Sprintf("http://%s:%d", ipn.IP, porta))
		}
	}
	return lista
}
