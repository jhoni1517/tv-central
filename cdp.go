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

// injetarPaginas conecta no navegador pela porta de depuração local e, em cada página (agora e
// nas próximas), coloca os scripts (setas, anúncios) e bloqueia os endereços de anúncio.
func injetarPaginas(ctx context.Context, porta int, scripts []string, bloquear []string) {
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
				if err := sessaoCDP(ctx, ws, scripts, bloquear); err != nil {
					log.Printf("injetar: página %s: %v", id, err)
				}
				mu.Lock()
				delete(ligados, id)
				mu.Unlock()
			}()
		}
	}
}

// sessaoCDP fica conectada enquanto a página existir (o script some se a conexão cair).
func sessaoCDP(ctx context.Context, wsURL string, scripts []string, bloquear []string) error {
	d := websocket.Dialer{HandshakeTimeout: 3 * time.Second}
	c, _, err := d.DialContext(ctx, wsURL, nil)
	if err != nil {
		return err
	}
	defer c.Close()
	go func() { <-ctx.Done(); c.Close() }()
	var msgs []map[string]any
	id := 0
	add := func(metodo string, params map[string]any) {
		id++
		msgs = append(msgs, map[string]any{"id": id, "method": metodo, "params": params})
	}
	if len(bloquear) > 0 {
		add("Network.enable", map[string]any{})
		add("Network.setBlockedURLs", map[string]any{"urls": bloquear})
	}
	for _, sc := range scripts {
		add("Page.addScriptToEvaluateOnNewDocument", map[string]any{"source": sc})
		add("Runtime.evaluate", map[string]any{"expression": sc})
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

// fecharNavegador manda o comando Browser.close pela porta de depuração local.
func fecharNavegador(porta int) bool {
	cli := &http.Client{Timeout: 2 * time.Second}
	resp, err := cli.Get(fmt.Sprintf("http://127.0.0.1:%d/json/version", porta))
	if err != nil {
		return false
	}
	var v struct {
		WS string `json:"webSocketDebuggerUrl"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&v)
	resp.Body.Close()
	if v.WS == "" {
		return false
	}
	d := websocket.Dialer{HandshakeTimeout: 2 * time.Second}
	c, _, err := d.Dial(v.WS, nil)
	if err != nil {
		return false
	}
	defer c.Close()
	return c.WriteJSON(map[string]any{"id": 1, "method": "Browser.close"}) == nil
}
