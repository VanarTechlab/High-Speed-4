package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCueYSettings(t *testing.T) {
	d := t.TempDir()
	cue := filepath.Join(d, "Need for Speed - High Stakes (USA).cue")
	os.WriteFile(cue, []byte("FILE \"Need for Speed - High Stakes (USA).bin\" BINARY\r\n  TRACK 01 MODE2/2352\r\n"), 0o644)
	b, err := binOfCue(cue)
	if err != nil || b != filepath.Join(d, "Need for Speed - High Stakes (USA).bin") {
		t.Fatal(b, err)
	}
	if err := writeDiscSetting(d, `D:\X\a b.cue`); err != nil {
		t.Fatal(err)
	}
	s, _ := os.ReadFile(filepath.Join(d, "settings.toml"))
	if !strings.Contains(string(s), `path = "D:/X/a b.cue"`) {
		t.Fatal(string(s))
	}
	os.WriteFile(filepath.Join(d, "settings.toml"), []byte("[video]\nx = 1\n[disc]\npath = \"viejo\"\n[memcard]\ny = 2\n"), 0o644)
	writeDiscSetting(d, `E:\n.cue`)
	s, _ = os.ReadFile(filepath.Join(d, "settings.toml"))
	if !strings.Contains(string(s), `path = "E:/n.cue"`) || strings.Contains(string(s), "viejo") || !strings.Contains(string(s), "y = 2") {
		t.Fatal(string(s))
	}
}
