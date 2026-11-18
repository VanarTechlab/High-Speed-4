package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const estadoUsuario = "format_version = 2\r\n\r\n[[package]]\r\nid = \"nfshs.4p\"\r\n\r\n[[package]]\r\nid = \"psx.enhancement.hd-textures\"\r\n\r\n[[feature]]\r\npackage_id = \"nfshs.4p\"\r\nid = \"4p\"\r\nenabled = false\r\n\r\n[[feature]]\r\npackage_id = \"psx.enhancement.hd-textures\"\r\nid = \"hd-textures\"\r\nenabled = true\r\n[feature.values]\r\ndump = \"false\"\r\nreplacements = \"true\"\r\n\r\n[[feature]]\r\npackage_id = \"psx.enhancement.pgxp\"\r\nid = \"pgxp\"\r\nenabled = false\r\n"

func TestSetHD(t *testing.T) {
	d := t.TempDir()
	k := &Kit{Base: d, Log: func(string) {}}
	os.MkdirAll(filepath.Join(d, "mods"), 0o755)
	os.WriteFile(k.stateFile(), []byte(estadoUsuario), 0o644)
	if err := k.SetHD(true, true); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(k.stateFile())
	s := string(b)
	if strings.Count(s, "package_id = \"psx.enhancement.hd-textures\"") != 1 || !strings.Contains(s, "dump = \"true\"") || !strings.Contains(s, "id = \"pgxp\"") || !strings.Contains(s, "id = \"4p\"") {
		t.Fatal(s)
	}
	en, du := k.HDState()
	if !en || !du {
		t.Fatal("estado", en, du)
	}
	// sin state.toml
	d2 := t.TempDir()
	k2 := &Kit{Base: d2, Log: func(string) {}}
	k2.SetHD(true, false)
	b, _ = os.ReadFile(k2.stateFile())
	t.Log(string(b))
	en, du = k2.HDState()
	if !en || du {
		t.Fatal("estado2", en, du)
	}
}
