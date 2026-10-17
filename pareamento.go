package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"
)

type Celular struct {
	Nome   string    `json:"nome"`
	Criado time.Time `json:"criado"`
}

// Pareamento: o celular pede um código, o código aparece na TV e o celular digita.
// Só o hash dos tokens fica salvo em disco.
type Pareamento struct {
	mu         sync.Mutex
	arquivo    string
	celulares  map[string]Celular
	codigo     string
	expira     time.Time
	tentativas int
	bloqueado  time.Time // depois de muitas tentativas erradas, espera um pouco
}

func novoPareamento(arquivo string) (*Pareamento, error) {
	p := &Pareamento{arquivo: arquivo, celulares: map[string]Celular{}}
	b, err := os.ReadFile(arquivo)
	if err == nil {
		err = json.Unmarshal(b, &p.celulares)
	} else if os.IsNotExist(err) {
		err = nil
	}
	return p, err
}

func hashToken(t string) string {
	h := sha256.Sum256([]byte(t))
	return hex.EncodeToString(h[:])
}

func aleatorio(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// IniciarCodigo cria (ou reaproveita) um código de 6 dígitos válido por 3 minutos.
// Devolve true quando o código é novo.
func (p *Pareamento) IniciarCodigo() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.codigo != "" && time.Now().Before(p.expira) {
		return false
	}
	if time.Now().Before(p.bloqueado) {
		return false
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	p.codigo = fmt.Sprintf("%06d", n.Int64())
	p.expira = time.Now().Add(3 * time.Minute)
	p.tentativas = 0
	return true
}

// Pendente devolve o código que deve aparecer na TV.
func (p *Pareamento) Pendente() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.codigo == "" || time.Now().After(p.expira) {
		p.codigo = ""
		return ""
	}
	return p.codigo
}

func (p *Pareamento) Confirmar(codigo, nome string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.codigo == "" || time.Now().After(p.expira) {
		return "", fmt.Errorf("código expirado: peça outro")
	}
	if subtle.ConstantTimeCompare([]byte(codigo), []byte(p.codigo)) != 1 {
		p.tentativas++
		if p.tentativas >= 5 {
			p.codigo = ""
			p.bloqueado = time.Now().Add(time.Minute)
			return "", fmt.Errorf("muitas tentativas: peça outro código")
		}
		return "", fmt.Errorf("código errado")
	}
	p.codigo = ""
	token := aleatorio(32)
	if nome == "" {
		nome = "Celular"
	}
	p.celulares[hashToken(token)] = Celular{Nome: nome, Criado: time.Now()}
	return token, p.salvar()
}

func (p *Pareamento) Valido(token string) bool {
	if len(token) < 32 {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.celulares[hashToken(token)]
	return ok
}

func (p *Pareamento) Quantos() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.celulares)
}

func (p *Pareamento) salvar() error {
	b, _ := json.MarshalIndent(p.celulares, "", "  ")
	tmp := p.arquivo + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p.arquivo)
}
