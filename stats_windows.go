//go:build windows

package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32               = windows.NewLazySystemDLL("kernel32.dll")
	procGetSystemTimes     = kernel32.NewProc("GetSystemTimes")
	procGlobalMemoryStatus = kernel32.NewProc("GlobalMemoryStatusEx")
)

func tempoCPU() (ocioso, total uint64) {
	var idle, kernel, user windows.Filetime
	procGetSystemTimes.Call(uintptr(unsafe.Pointer(&idle)), uintptr(unsafe.Pointer(&kernel)), uintptr(unsafe.Pointer(&user)))
	f := func(t windows.Filetime) uint64 { return uint64(t.HighDateTime)<<32 | uint64(t.LowDateTime) }
	return f(idle), f(kernel) + f(user) // o tempo de kernel já inclui o ocioso
}

func memoriaEDisco() Stats {
	var m struct {
		Length, MemoryLoad                         uint32
		TotalPhys, AvailPhys, TotalPage, AvailPage uint64
		TotalVirtual, AvailVirtual, AvailExtended  uint64
	}
	m.Length = uint32(unsafe.Sizeof(m))
	procGlobalMemoryStatus.Call(uintptr(unsafe.Pointer(&m)))
	s := Stats{MemLivreMB: int(m.AvailPhys >> 20), MemTotalMB: int(m.TotalPhys >> 20)}
	var livre, totalDisco, totalLivre uint64
	raiz, _ := windows.UTF16PtrFromString(`C:\`)
	if windows.GetDiskFreeSpaceEx(raiz, &livre, &totalDisco, &totalLivre) == nil {
		s.DiscoGB = float64(livre>>20) / 1024
	}
	return s
}
