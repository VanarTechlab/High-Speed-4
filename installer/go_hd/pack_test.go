package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestInstalarPack(t *testing.T) {
	d := t.TempDir()
	zp := filepath.Join(d, "p.zip")
	f, _ := os.Create(zp)
	zw := zip.NewWriter(f)
	for _, n := range []string{"HighSpeed4_TexturasHD/texturas/a.png", "HighSpeed4_TexturasHD/texturas/b.PNG", "HighSpeed4_TexturasHD/LEEME.txt"} {
		w, _ := zw.Create(n)
		w.Write([]byte("x"))
	}
	zw.Close()
	f.Close()
	k := &Kit{Base: filepath.Join(d, "juego"), Log: func(string) {}}
	if err := k.InstalarPack(zp); err != nil {
		t.Fatal(err)
	}
	if n := len(listPNG(k.repl())); n != 2 {
		t.Fatal("instaladas", n)
	}
	if en, du := k.HDState(); !en || du {
		t.Fatal(en, du)
	}
}
