//go:build !windows

package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func tempoCPU() (ocioso, total uint64) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0
	}
	defer f.Close()
	l, _ := bufio.NewReader(f).ReadString('\n')
	campos := strings.Fields(l)
	for i, c := range campos[1:] {
		v, _ := strconv.ParseUint(c, 10, 64)
		total += v
		if i == 3 || i == 4 { // idle + iowait
			ocioso += v
		}
	}
	return
}

func memoriaEDisco() Stats {
	var s Stats
	if b, err := os.ReadFile("/proc/meminfo"); err == nil {
		for _, l := range strings.Split(string(b), "\n") {
			c := strings.Fields(l)
			if len(c) < 2 {
				continue
			}
			v, _ := strconv.Atoi(c[1])
			switch c[0] {
			case "MemTotal:":
				s.MemTotalMB = v / 1024
			case "MemAvailable:":
				s.MemLivreMB = v / 1024
			}
		}
	}
	var st syscall.Statfs_t
	if syscall.Statfs("/", &st) == nil {
		s.DiscoGB = float64(st.Bavail*uint64(st.Bsize)>>20) / 1024
	}
	return s
}
