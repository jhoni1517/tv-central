//go:build !windows

// Versões de mentira para rodar e testar as telas fora do Windows.
package main

import (
	"errors"
	"log"
	"os/exec"
	"strconv"
	"syscall"
)

func avisar(titulo, msg string)      { log.Printf("[%s] %s", titulo, msg) }
func focarProcesso(pid int)          {}
func precisaInstalar() bool          { return false }
func instalar() error                { return errors.New("instalador só funciona no Windows") }
func remover() error                 { return errors.New("só no Windows") }
func apertarTecla(nome string) error { log.Printf("tecla %s", nome); return nil }
func digitarTexto(s string)          { log.Printf("texto %q", s) }
func moverMouse(dx, dy int)          { log.Printf("mouse %d,%d", dx, dy) }
func clicar(b string)                { log.Printf("clique %s", b) }
func rolar(dy int)                   { log.Printf("rolar %d", dy) }

func acaoSistema(acao string) error {
	log.Printf("sistema %s", acao)
	if acao != "desligar" && acao != "reiniciar" {
		return errors.New("ação desconhecida")
	}
	return nil
}

func matarArvore(pid int) {
	_ = syscall.Kill(-pid, syscall.SIGKILL)
	_ = exec.Command("kill", strconv.Itoa(pid)).Run()
}
