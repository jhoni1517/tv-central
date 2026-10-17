package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// injetarSetas conecta no navegador pela porta de depuração local e coloca o script de
// navegação por setas em cada página (agora e nas próximas que abrirem).
func injetarSetas(ctx context.Context, porta int, script string) {
	var mu sync.Mutex
	ligados := map[string]bool{}
	cli := &http.Client{Timeout: 2 * time.Second}
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		resp, err := cli.Get(fmt.Sprintf("http://127.0.0.1:%d/json/list", porta))
		if err != nil {
			continue
		}
		var alvos []struct {
			ID, Type, URL        string
			WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&alvos)
		resp.Body.Close()
		for _, a := range alvos {
			if a.Type != "page" || a.WebSocketDebuggerURL == "" {
				continue
			}
			mu.Lock()
			ja := ligados[a.ID]
			ligados[a.ID] = true
			mu.Unlock()
			if ja {
				continue
			}
			id, ws := a.ID, a.WebSocketDebuggerURL
			go func() {
				if err := sessaoCDP(ctx, ws, script); err != nil {
					log.Printf("setas: página %s: %v", id, err)
				}
				mu.Lock()
				delete(ligados, id)
				mu.Unlock()
			}()
		}
	}
}

// sessaoCDP fica conectada enquanto a página existir (o script some se a conexão cair).
func sessaoCDP(ctx context.Context, wsURL, script string) error {
	d := websocket.Dialer{HandshakeTimeout: 3 * time.Second}
	c, _, err := d.DialContext(ctx, wsURL, nil)
	if err != nil {
		return err
	}
	defer c.Close()
	go func() { <-ctx.Done(); c.Close() }()
	msgs := []map[string]any{
		{"id": 1, "method": "Page.addScriptToEvaluateOnNewDocument", "params": map[string]any{"source": script}},
		{"id": 2, "method": "Runtime.evaluate", "params": map[string]any{"expression": script}},
	}
	for _, m := range msgs {
		if err := c.WriteJSON(m); err != nil {
			return err
		}
	}
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return nil // página fechou
		}
	}
}
