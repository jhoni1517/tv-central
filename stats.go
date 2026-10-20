package main

import (
	"sync"
	"time"
)

// Estatísticas do PC (CPU, memória, disco) para a linha opcional embaixo do relógio.
type Stats struct {
	CPU        int     `json:"cpu"` // %
	MemLivreMB int     `json:"memLivreMB"`
	MemTotalMB int     `json:"memTotalMB"`
	DiscoGB    float64 `json:"discoLivreGB"`
}

type Medidor struct {
	mu      sync.Mutex
	ultimo  Stats
	pedido  time.Time
	rodando bool
}

// Atual devolve a última medição e mantém o medidor ligado só enquanto alguém está olhando.
func (m *Medidor) Atual() Stats {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pedido = time.Now()
	if !m.rodando {
		m.rodando = true
		go m.laco()
	}
	return m.ultimo
}

func (m *Medidor) laco() {
	ocioso, total := tempoCPU()
	for {
		time.Sleep(3 * time.Second)
		o2, t2 := tempoCPU()
		s := memoriaEDisco()
		if dt := t2 - total; dt > 0 {
			s.CPU = int(100 - 100*float64(o2-ocioso)/float64(dt))
			s.CPU = min(100, max(0, s.CPU))
		}
		ocioso, total = o2, t2
		m.mu.Lock()
		m.ultimo = s
		parar := time.Since(m.pedido) > 20*time.Second
		if parar {
			m.rodando = false
		}
		m.mu.Unlock()
		if parar {
			return
		}
	}
}
