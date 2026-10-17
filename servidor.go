package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strings"
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
	cfg Config
	par *Pareamento
	ger *Gerente
}

func novoServidor(cfg Config, par *Pareamento, ger *Gerente) *Servidor {
	return &Servidor{cfg: cfg, par: par, ger: ger}
}

func (s *Servidor) Rodar() error {
	mux := http.NewServeMux()
	celular, _ := fs.Sub(arquivosWeb, "web/celular")
	tv, _ := fs.Sub(arquivosWeb, "web/tv")

	// Celular (qualquer aparelho da rede local)
	mux.Handle("GET /", semCache(http.FileServer(http.FS(celular))))
	mux.HandleFunc("POST /api/parear/iniciar", s.iniciarPareamento)
	mux.HandleFunc("POST /api/parear", s.parear)
	mux.HandleFunc("GET /api/eu", s.autenticado(func(w http.ResponseWriter, r *http.Request) { responder(w, map[string]any{"ok": true}) }))
	mux.HandleFunc("GET /api/apps", s.autenticado(s.listarApps))
	mux.HandleFunc("GET /ws", s.autenticado(s.controle))

	// Tela da TV (só o próprio computador)
	mux.Handle("GET /tv/", soLocal(semCache(http.StripPrefix("/tv/", http.FileServer(http.FS(tv))))))
	mux.HandleFunc("GET /tv/api/estado", soLocalF(s.estadoTV))
	mux.HandleFunc("GET /tv/qr.png", soLocalF(s.qr))
	mux.HandleFunc("POST /tv/api/abrir", soLocalF(s.abrirPelaTV))
	mux.HandleFunc("POST /tv/api/sistema", soLocalF(s.sistema))

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.cfg.Porta),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("ouvindo na porta %d: %s", s.cfg.Porta, strings.Join(enderecos(s.cfg.Porta), " "))
	return srv.ListenAndServe()
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
		if !ehLocal(r) {
			erro(w, http.StatusForbidden, "só no próprio computador")
			return
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
	s.par.IniciarCodigo()
	s.ger.AbrirLauncher() // o código aparece na tela inicial
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
	ID, Nome, Cor, Cor2 string
}

func (s *Servidor) appsPublicos() []appPublico {
	l := make([]appPublico, 0, len(s.cfg.apps))
	for _, a := range s.cfg.apps {
		l = append(l, appPublico{a.ID, a.Nome, a.Cor, a.Cor2})
	}
	return l
}

func (s *Servidor) listarApps(w http.ResponseWriter, r *http.Request) {
	responder(w, map[string]any{"apps": s.appsPublicos(), "atual": s.ger.Atual()})
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
		"codigo":    s.par.Pendente(),
		"celulares": s.par.Quantos(),
		"enderecos": enderecos(s.cfg.Porta),
		"apps":      s.appsPublicos(),
		"versao":    versao,
	})
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
}

func (s *Servidor) controle(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	c.SetReadLimit(8192)
	for {
		var m Msg
		if err := c.ReadJSON(&m); err != nil {
			return
		}
		if err := s.executar(m); err != nil {
			log.Printf("controle: %s %s: %v", m.T, m.K, err)
			_ = c.WriteJSON(map[string]string{"erro": err.Error()})
		} else if m.T == "ping" {
			_ = c.WriteJSON(map[string]string{"t": "pong", "atual": s.ger.Atual()})
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
