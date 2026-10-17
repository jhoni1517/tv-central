//go:build windows

package main

import (
	"fmt"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32        = windows.NewLazySystemDLL("user32.dll")
	procSendInput = user32.NewProc("SendInput")
)

const (
	inputMouse   = 0
	inputTeclado = 1
	kfEstendida  = 0x0001
	kfSolta      = 0x0002
	kfUnicode    = 0x0004
	mfMover      = 0x0001
	mfEsqDesce   = 0x0002
	mfEsqSobe    = 0x0004
	mfDirDesce   = 0x0008
	mfDirSobe    = 0x0010
	mfRoda       = 0x0800
	deltaRoda    = 120
	tamanhoInput = unsafe.Sizeof(entradaWin{})
)

type tecladoWin struct {
	Vk, Scan    uint16
	Flags, Hora uint32
	Extra       uintptr
}

type mouseWin struct {
	DX, DY             int32
	Dados, Flags, Hora uint32
	Extra              uintptr
}

// entradaWin é o INPUT do Windows (40 bytes em 64 bits).
type entradaWin struct {
	Tipo uint32
	_    uint32
	U    [4]uint64
}

func teclado(k tecladoWin) entradaWin {
	var e entradaWin
	e.Tipo = inputTeclado
	*(*tecladoWin)(unsafe.Pointer(&e.U[0])) = k
	return e
}

func mouse(m mouseWin) entradaWin {
	var e entradaWin
	e.Tipo = inputMouse
	*(*mouseWin)(unsafe.Pointer(&e.U[0])) = m
	return e
}

func enviar(es []entradaWin) {
	if len(es) == 0 {
		return
	}
	procSendInput.Call(uintptr(len(es)), uintptr(unsafe.Pointer(&es[0])), tamanhoInput)
}

type tecla struct {
	vk  uint16
	ext bool
}

var teclas = map[string]tecla{
	"cima": {0x26, true}, "baixo": {0x28, true}, "esquerda": {0x25, true}, "direita": {0x27, true},
	"ok": {0x0D, false}, "voltar": {0x1B, false}, "apagar": {0x08, false}, "espaco": {0x20, false},
	"tab": {0x09, false}, "telacheia": {0x7A, false},
	"volmais": {0xAF, true}, "volmenos": {0xAE, true}, "mudo": {0xAD, true},
	"playpause": {0xB3, true}, "proximo": {0xB0, true}, "anterior": {0xB1, true},
	"alt": {0x12, false},
}

func apertarTecla(nome string) error {
	t, ok := teclas[nome]
	if !ok {
		return fmt.Errorf("tecla %q desconhecida", nome)
	}
	var f uint32
	if t.ext {
		f = kfEstendida
	}
	enviar([]entradaWin{
		teclado(tecladoWin{Vk: t.vk, Flags: f}),
		teclado(tecladoWin{Vk: t.vk, Flags: f | kfSolta}),
	})
	return nil
}

// digitarTexto manda cada caractere como Unicode (funciona com acentos e emoji).
func digitarTexto(s string) {
	var es []entradaWin
	for _, u := range utf16.Encode([]rune(s)) {
		if u == '\n' {
			es = append(es, teclado(tecladoWin{Vk: 0x0D}), teclado(tecladoWin{Vk: 0x0D, Flags: kfSolta}))
			continue
		}
		es = append(es,
			teclado(tecladoWin{Scan: u, Flags: kfUnicode}),
			teclado(tecladoWin{Scan: u, Flags: kfUnicode | kfSolta}))
	}
	enviar(es)
}

func moverMouse(dx, dy int) {
	enviar([]entradaWin{mouse(mouseWin{DX: int32(dx), DY: int32(dy), Flags: mfMover})})
}

func clicar(botao string) {
	desce, sobe := uint32(mfEsqDesce), uint32(mfEsqSobe)
	if botao == "direito" {
		desce, sobe = mfDirDesce, mfDirSobe
	}
	enviar([]entradaWin{mouse(mouseWin{Flags: desce}), mouse(mouseWin{Flags: sobe})})
}

func rolar(dy int) {
	enviar([]entradaWin{mouse(mouseWin{Dados: uint32(int32(-dy * deltaRoda)), Flags: mfRoda})})
}
