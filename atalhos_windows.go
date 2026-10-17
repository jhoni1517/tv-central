//go:build windows

package main

import (
	"log"
	"runtime"
	"unsafe"
)

var (
	procRegisterHotKey = user32.NewProc("RegisterHotKey")
	procGetMessage     = user32.NewProc("GetMessageW")
)

// Teclas globais (funcionam com qualquer janela na frente):
//   - "Home do navegador" (tecla de mídia de teclados e de receptores de controle remoto, como o FLIRC)
//   - Ctrl+Alt+H
//
// As duas voltam para a tela inicial.
func atalhosGlobais(inicio func()) {
	go func() {
		runtime.LockOSThread() // as teclas registradas chegam na fila desta thread
		const modCtrlAlt, modSemRepetir = 0x0002 | 0x0001, 0x4000
		ok1, _, _ := procRegisterHotKey.Call(0, 1, modSemRepetir, 0xAC)           // VK_BROWSER_HOME
		ok2, _, _ := procRegisterHotKey.Call(0, 2, modCtrlAlt|modSemRepetir, 'H') // Ctrl+Alt+H
		log.Printf("atalhos globais: home=%v ctrl+alt+h=%v", ok1 != 0, ok2 != 0)
		var msg struct {
			Hwnd    uintptr
			Message uint32
			WParam  uintptr
			LParam  uintptr
			Time    uint32
			X, Y    int32
			_       uint32
		}
		for {
			r, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
			if int32(r) <= 0 {
				return
			}
			if msg.Message == 0x0312 { // WM_HOTKEY
				go inicio()
			}
		}
	}()
}
