package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestBloqueioReal(t *testing.T) {
	var anuncio, normal atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "doubleclick.net"):
			anuncio.Add(1)
		case r.URL.Path == "/normal.js":
			normal.Add(1)
		case r.URL.Path == "/":
			fmt.Fprint(w, "<html><body>oi</body></html>")
		}
	}))
	defer srv.Close()
	c := exec.Command("/opt/pw-browsers/chromium-1194/chrome-linux/chrome", "--headless=new", "--no-sandbox", "--remote-debugging-port=39333", "--user-data-dir="+dirTeste(t), srv.URL)
	if err := c.Start(); err != nil {
		t.Skip(err)
	}
	defer func() { c.Process.Kill(); c.Wait() }()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go injetarPaginas(ctx, 39333, []string{"window.__injetado = 1"}, padroesBloqueio())
	time.Sleep(3500 * time.Millisecond)
	// a página pede um anúncio e um script normal (pelo CDP, avaliando na página)
	ws := ""
	for i := 0; i < 20 && ws == ""; i++ {
		resp, err := http.Get("http://127.0.0.1:39333/json/list")
		if err == nil {
			var l []struct{ Type, WebSocketDebuggerURL string }
			_ = jsonDecode(resp, &l)
			for _, x := range l {
				if x.Type == "page" {
					ws = x.WebSocketDebuggerURL
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	if ws == "" {
		t.Skip("sem página")
	}
	js := fmt.Sprintf("fetch('%s/doubleclick.net/ad.js').catch(()=>{}); fetch('%s/normal.js').catch(()=>{}); 1", srv.URL, srv.URL)
	_ = avaliar(ws, js)
	time.Sleep(1500 * time.Millisecond)
	if normal.Load() == 0 {
		t.Fatal("pedido normal não chegou (teste quebrado)")
	}
	if anuncio.Load() != 0 {
		t.Errorf("anúncio não foi bloqueado (%d pedidos)", anuncio.Load())
	}
}

func jsonDecode(r *http.Response, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func avaliar(ws, js string) error {
	c, _, err := websocket.DefaultDialer.Dial(ws, nil)
	if err != nil {
		return err
	}
	defer c.Close()
	return c.WriteJSON(map[string]any{"id": 1, "method": "Runtime.evaluate", "params": map[string]any{"expression": js}})
}

func dirTeste(t *testing.T) string {
	d, _ := os.MkdirTemp("", "cdp")
	t.Cleanup(func() { time.Sleep(300 * time.Millisecond); os.RemoveAll(d) })
	return d
}
