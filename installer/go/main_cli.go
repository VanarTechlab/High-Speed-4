//go:build !gui

package main

// instalar_hs4: la misma instalacion que la ventana, desde la consola.
// Uso: instalar_hs4 DISCO.cue CARPETA_DESTINO [--sin-preparar] [--sin-accesos]

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	args := []string{}
	o := &Options{Preparar: true, Accesos: true, Log: func(s string) { fmt.Println(s) }}
	for _, a := range os.Args[1:] {
		switch a {
		case "--sin-preparar":
			o.Preparar = false
		case "--sin-accesos":
			o.Accesos = false
		default:
			args = append(args, a)
		}
	}
	if len(args) != 2 {
		fmt.Println("Uso: instalar_hs4 DISCO.cue CARPETA_DESTINO [--sin-preparar] [--sin-accesos]")
		os.Exit(2)
	}
	exe, _ := os.Executable()
	o.Base = filepath.Dir(exe)
	if b := os.Getenv("HS4_BASE"); b != "" {
		o.Base = b
	}
	o.Cue, o.Dest = args[0], args[1]
	if err := Run(o); err != nil {
		fmt.Println("\n*** ERROR:", err)
		os.Exit(1)
	}
}
