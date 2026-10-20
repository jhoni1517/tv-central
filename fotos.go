package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

// Fotos do celular na TV: viram protetor de tela e dá para mostrar uma na hora.
var extFoto = []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}

type Fotos struct {
	mu    sync.Mutex
	pasta string
	agora struct {
		Nome   string `json:"nome"`
		Quando int64  `json:"quando"`
	}
}

func (f *Fotos) Lista() []string {
	itens, _ := os.ReadDir(f.pasta)
	type foto struct {
		nome string
		t    time.Time
	}
	var l []foto
	for _, it := range itens {
		if it.IsDir() || !slices.Contains(extFoto, strings.ToLower(filepath.Ext(it.Name()))) || strings.HasPrefix(it.Name(), ".") {
			continue
		}
		info, err := it.Info()
		if err != nil {
			continue
		}
		l = append(l, foto{it.Name(), info.ModTime()})
	}
	sort.Slice(l, func(i, j int) bool { return l[i].t.After(l[j].t) }) // mais novas primeiro
	nomes := make([]string, len(l))                                    // nunca nil: vira [] no JSON, não null
	for i, x := range l {
		nomes[i] = x.nome
	}
	return nomes
}

func nomeFotoValido(n string) bool {
	return n != "" && len(n) < 80 && !strings.ContainsAny(n, `/\:`) && !strings.HasPrefix(n, ".") &&
		slices.Contains(extFoto, strings.ToLower(filepath.Ext(n)))
}

func (f *Fotos) Mostrar(nome string) {
	f.mu.Lock()
	f.agora.Nome, f.agora.Quando = nome, time.Now().UnixMilli()
	f.mu.Unlock()
}

func (f *Fotos) Agora() any {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.agora.Nome == "" || time.Since(time.UnixMilli(f.agora.Quando)) > time.Minute {
		return nil
	}
	return f.agora
}

// localOuPareado: a TV (o próprio PC) ou um celular pareado (fotos são pessoais).
func (s *Servidor) localOuPareado(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ehLocal(r) || s.par.Valido(token(r)) {
			h(w, r)
			return
		}
		erro(w, http.StatusUnauthorized, "celular não pareado")
	}
}

func (s *Servidor) rotasFotos(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/fotos", s.localOuPareado(func(w http.ResponseWriter, r *http.Request) {
		responder(w, map[string]any{"fotos": s.fotos.Lista()})
	}))
	mux.HandleFunc("GET /foto/{nome}", s.localOuPareado(func(w http.ResponseWriter, r *http.Request) {
		nome := r.PathValue("nome")
		if !nomeFotoValido(nome) {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "private, max-age=86400")
		http.ServeFile(w, r, filepath.Join(s.fotos.pasta, nome))
	}))
	mux.HandleFunc("POST /api/fotos", s.autenticado(s.enviarFotos))
	mux.HandleFunc("DELETE /api/fotos/{nome}", s.autenticado(func(w http.ResponseWriter, r *http.Request) {
		nome := r.PathValue("nome")
		if s.aj.Ler().Crianca.Ativo {
			erro(w, http.StatusForbidden, errModoCrianca.Error())
			return
		}
		if !nomeFotoValido(nome) || os.Remove(filepath.Join(s.fotos.pasta, nome)) != nil {
			erro(w, http.StatusNotFound, "foto não encontrada")
			return
		}
		responder(w, map[string]any{"ok": true})
	}))
	mux.HandleFunc("POST /api/fotos/mostrar", s.autenticado(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Nome string }
		if !lerJSON(w, r, &req) {
			return
		}
		if !nomeFotoValido(req.Nome) || !existe(filepath.Join(s.fotos.pasta, req.Nome)) {
			erro(w, http.StatusNotFound, "foto não encontrada")
			return
		}
		s.fotos.Mostrar(req.Nome)
		go s.ger.AbrirLauncher() // traz a tela inicial para a frente para mostrar a foto
		responder(w, map[string]any{"ok": true})
	}))
}

func (s *Servidor) enviarFotos(w http.ResponseWriter, r *http.Request) {
	if s.aj.Ler().Crianca.Ativo {
		erro(w, http.StatusForbidden, errModoCrianca.Error())
		return
	}
	if len(s.fotos.Lista()) >= 500 {
		erro(w, http.StatusBadRequest, "já tem 500 fotos: apague algumas antes")
		return
	}
	const limite = 20 << 20
	r.Body = http.MaxBytesReader(w, r.Body, 30*limite)
	leitor, err := r.MultipartReader()
	if err != nil {
		erro(w, http.StatusBadRequest, "envie como formulário com arquivos")
		return
	}
	_ = os.MkdirAll(s.fotos.pasta, 0o755)
	salvas, falhas := []string{}, []string{}
	for {
		parte, err := leitor.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			falhas = append(falhas, "envio interrompido")
			break
		}
		ext := strings.ToLower(filepath.Ext(parte.FileName()))
		if ext == ".jpeg" {
			ext = ".jpg"
		}
		if !slices.Contains(extFoto, ext) {
			falhas = append(falhas, parte.FileName()+" (formato não aceito; no iPhone use JPG)")
			continue
		}
		nome := fmt.Sprintf("%s-%s%s", time.Now().Format("20060102-150405"), aleatorio(3), ext)
		tmp := filepath.Join(s.fotos.pasta, ".envio-"+nome)
		f, err := os.Create(tmp)
		if err != nil {
			falhas = append(falhas, parte.FileName())
			continue
		}
		n, err := io.Copy(f, io.LimitReader(parte, limite+1))
		f.Close()
		if err != nil || n == 0 || n > limite {
			os.Remove(tmp)
			falhas = append(falhas, parte.FileName()+" (vazia ou maior que 20 MB)")
			continue
		}
		if os.Rename(tmp, filepath.Join(s.fotos.pasta, nome)) != nil {
			os.Remove(tmp)
			falhas = append(falhas, parte.FileName())
			continue
		}
		salvas = append(salvas, nome)
	}
	log.Printf("fotos: %d recebidas, %d com problema", len(salvas), len(falhas))
	responder(w, map[string]any{"salvas": salvas, "falhas": falhas})
}
