package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Clima no topo da tela inicial (Open-Meteo: gratuito, sem cadastro).
// Liga com "cidade": "Curitiba" no config.json.
type Clima struct {
	mu     sync.Mutex
	cidade string
	Temp   *float64 `json:"temp,omitempty"`
	Icone  string   `json:"icone,omitempty"`
	Texto  string   `json:"texto,omitempty"`
	lat    float64
	lon    float64
	achou  bool
	cli    *http.Client
	mudou  chan struct{}
}

func novoClima(cidade string) *Clima {
	return &Clima{cidade: cidade, cli: &http.Client{Timeout: 10 * time.Second}, mudou: make(chan struct{}, 1)}
}

func (c *Clima) Atual() map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Temp == nil {
		return nil
	}
	return map[string]any{"temp": *c.Temp, "icone": c.Icone, "texto": c.Texto}
}

// Definir troca a cidade (pelo painel do celular); vazio desliga o clima.
func (c *Clima) Definir(cidade string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if cidade == c.cidade {
		return
	}
	c.cidade, c.achou, c.Temp = cidade, false, nil
	select {
	case c.mudou <- struct{}{}:
	default:
	}
}

func (c *Clima) Loop() {
	for {
		c.mu.Lock()
		cidade := c.cidade
		c.mu.Unlock()
		espera := 30 * time.Minute
		if cidade != "" {
			if err := c.atualizar(cidade); err != nil {
				log.Printf("clima: %v", err)
				espera = 5 * time.Minute
			}
		}
		select {
		case <-c.mudou:
		case <-time.After(espera):
		}
	}
}

func (c *Clima) pegar(u string, v any) error {
	resp, err := c.cli.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(v)
}

func (c *Clima) atualizar(cidade string) error {
	c.mu.Lock()
	achou, lat, lon := c.achou, c.lat, c.lon
	c.mu.Unlock()
	if !achou {
		var g struct {
			Results []struct{ Latitude, Longitude float64 } `json:"results"`
		}
		if err := c.pegar("https://geocoding-api.open-meteo.com/v1/search?count=1&language=pt&name="+url.QueryEscape(cidade), &g); err != nil {
			return err
		}
		if len(g.Results) == 0 {
			return fmt.Errorf("cidade %q não encontrada", cidade)
		}
		lat, lon = g.Results[0].Latitude, g.Results[0].Longitude
		c.mu.Lock()
		if c.cidade == cidade {
			c.lat, c.lon, c.achou = lat, lon, true
		}
		c.mu.Unlock()
	}
	var f struct {
		Current struct {
			Temp float64 `json:"temperature_2m"`
			Code int     `json:"weather_code"`
			Dia  int     `json:"is_day"`
		} `json:"current"`
	}
	u := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,weather_code,is_day&timezone=auto", lat, lon)
	if err := c.pegar(u, &f); err != nil {
		return err
	}
	icone, texto := tempo(f.Current.Code, f.Current.Dia == 1)
	c.mu.Lock()
	if c.cidade == cidade { // a cidade pode ter mudado no meio do caminho
		t := f.Current.Temp
		c.Temp, c.Icone, c.Texto = &t, icone, texto
	}
	c.mu.Unlock()
	return nil
}

// Códigos WMO usados pelo Open-Meteo.
func tempo(code int, dia bool) (string, string) {
	switch {
	case code == 0:
		if dia {
			return "☀️", "Céu limpo"
		}
		return "🌙", "Céu limpo"
	case code <= 2:
		if dia {
			return "🌤️", "Poucas nuvens"
		}
		return "☁️", "Poucas nuvens"
	case code == 3:
		return "☁️", "Nublado"
	case code == 45 || code == 48:
		return "🌫️", "Neblina"
	case code >= 51 && code <= 57:
		return "🌦️", "Garoa"
	case code >= 61 && code <= 67, code >= 80 && code <= 82:
		return "🌧️", "Chuva"
	case code >= 71 && code <= 77, code == 85 || code == 86:
		return "🌨️", "Neve"
	case code >= 95:
		return "⛈️", "Tempestade"
	}
	return "🌡️", ""
}
