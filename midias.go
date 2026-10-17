package main

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Vídeos de fundo que a pessoa coloca na pasta capas: <id>.mp4 ou <id>.webm
// (id de um app, "fundo" para a tela sem app em foco).
type Midias struct {
	mu     sync.Mutex
	pasta  string
	videos map[string]string // id -> arquivo
	quando time.Time
}

var extVideo = []string{".mp4", ".webm", ".m4v"}

// Videos devolve a lista atual (relê a pasta no máximo a cada 5 s).
func (m *Midias) Videos() map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if time.Since(m.quando) < 5*time.Second && m.videos != nil {
		return m.videos
	}
	m.quando = time.Now()
	m.videos = map[string]string{}
	itens, _ := os.ReadDir(m.pasta)
	for _, it := range itens {
		if it.IsDir() {
			continue
		}
		nome := strings.ToLower(it.Name())
		for _, ext := range extVideo {
			if id, ok := strings.CutSuffix(nome, ext); ok && idValido(id) {
				m.videos[id] = filepath.Join(m.pasta, it.Name())
			}
		}
	}
	return m.videos
}

func (m *Midias) IDs() []string {
	ids := []string{}
	for id := range m.Videos() {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Servir manda o vídeo com suporte a "pular" (Range), que o navegador usa para tocar em loop.
func (m *Midias) Servir(w http.ResponseWriter, r *http.Request) {
	arq, ok := m.Videos()[r.PathValue("id")]
	if !ok {
		http.NotFound(w, r)
		return
	}
	f, err := os.Open(arq)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	if strings.HasSuffix(strings.ToLower(arq), ".webm") {
		w.Header().Set("Content-Type", "video/webm") // o Go chutaria "audio/webm"
	} else {
		w.Header().Set("Content-Type", "video/mp4")
	}
	http.ServeContent(w, r, arq, fi.ModTime(), f)
}

// Esquecer faz a próxima consulta reler a pasta (depois de enviar ou apagar um arquivo).
func (m *Midias) Esquecer() {
	m.mu.Lock()
	m.quando = time.Time{}
	m.mu.Unlock()
}
