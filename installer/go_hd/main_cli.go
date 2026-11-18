//go:build !gui

package main

// texturas_hd: el kit HD por consola.  Uso: texturas_hd [pack ZIP|recoger|crear|activar|desactivar|estado]

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	exe, _ := os.Executable()
	k := &Kit{Base: filepath.Dir(exe), Log: func(s string) { fmt.Println(s) }}
	if b := os.Getenv("HS4_JUEGO"); b != "" {
		k.Base = b
	}
	acc := "estado"
	if len(os.Args) > 1 {
		acc = os.Args[1]
	}
	var err error
	switch acc {
	case "recoger":
		err = k.SetHD(true, true)
		fmt.Println("Recogida de texturas ACTIVADA: juega y cierra el juego.")
	case "pack":
		if len(os.Args) < 3 {
			fmt.Println("Uso: texturas_hd pack ARCHIVO.zip")
			os.Exit(2)
		}
		err = k.InstalarPack(os.Args[2])
	case "crear":
		err = k.Crear()
	case "activar":
		err = k.SetHD(true, false)
	case "desactivar":
		err = k.SetHD(false, false)
	default:
		c := k.Count()
		en, du := k.HDState()
		fmt.Printf("HD activo: %v  recogida: %v  recogidas: %d  instaladas: %d  pendientes: %d\n", en, du, c.Recogidas, c.Instaladas, c.Pendientes)
	}
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}
