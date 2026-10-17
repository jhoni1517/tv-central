package main

import (
	"log"
	"sync"
	"time"
)

// Desligamento programado ("desligar em 30 min"), bom para dormir vendo filme.
type Timer struct {
	mu  sync.Mutex
	ate time.Time
	t   *time.Timer
}

// Definir agenda para daqui a min minutos; 0 cancela.
func (d *Timer) Definir(min int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.t != nil {
		d.t.Stop()
		d.t = nil
	}
	d.ate = time.Time{}
	if min <= 0 {
		log.Printf("desligamento programado cancelado")
		return
	}
	if min > 600 {
		min = 600
	}
	d.ate = time.Now().Add(time.Duration(min) * time.Minute)
	d.t = time.AfterFunc(time.Duration(min)*time.Minute, func() {
		log.Printf("desligamento programado: desligando")
		if err := acaoSistema("desligar"); err != nil {
			log.Printf("desligar: %v", err)
		}
	})
	log.Printf("desligamento programado para daqui a %d min", min)
}

// Restante em segundos (0 = sem timer).
func (d *Timer) Restante() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.ate.IsZero() {
		return 0
	}
	return max(0, int(time.Until(d.ate).Seconds()))
}
